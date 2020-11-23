package logic

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/util"
	"sync"
	"time"
)


type RoleBuildMgr struct {
	mutex sync.RWMutex
	dbRB  map[int]*model.MapRoleBuild //key:dbId
	posRB map[int]*model.MapRoleBuild //key:posId
	roleRB map[int][]*model.MapRoleBuild //key:roleId
}


var RBMgr = &RoleBuildMgr{
	dbRB: make(map[int]*model.MapRoleBuild),
	posRB: make(map[int]*model.MapRoleBuild),
	roleRB: make(map[int][]*model.MapRoleBuild),
}

func (this* RoleBuildMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.dbRB)
	if err != nil {
		log.DefaultLog.Error("RoleBuildMgr load role_build table error", zap.Error(err))
	}

	//转成posRB 和 roleRB
	for _, v := range this.dbRB {
		posId := ToPosition(v.X, v.Y)
		this.posRB[posId] = v
		_,ok := this.roleRB[v.RId]
		if ok == false{
			this.roleRB[v.RId] = make([]*model.MapRoleBuild, 0)
		}
		this.roleRB[v.RId] = append(this.roleRB[v.RId], v)
	}
	go this.toDatabase()
}


func (this* RoleBuildMgr) toDatabase() {
	for true {
		time.Sleep(5*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.dbRB {
			if v.DB.NeedSync() {
				v.DB.BeginSync()
				_, err := db.MasterDB.Table(v).ID(v.Id).Cols("rid",
					"cur_durable", "max_durable").Update(v)
				if err != nil{
					log.DefaultLog.Error("RoleResMgr toDatabase error", zap.Error(err))
				}
				v.DB.EndSync()
				cnt+=1
			}

			//一次最多更新20个
			if cnt>20{
				break
			}
		}
		this.mutex.RUnlock()
	}
}

/*
该位置是否被角色占领
*/
func (this* RoleBuildMgr) IsEmpty(x, y int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := ToPosition(x, y)
	_, ok := this.posRB[posId]
	return !ok
}

func (this* RoleBuildMgr) PositionBuild(x, y int) (*model.MapRoleBuild, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := ToPosition(x, y)
	b,ok := this.posRB[posId]
	if ok && b.RId != 0 {
		return b, ok
	}else{
		return nil, false
	}
}


func (this* RoleBuildMgr) AddBuild(rid, x, y int) (*model.MapRoleBuild, bool) {

	posId := ToPosition(x, y)
	this.mutex.Lock()
	rb, ok := this.posRB[posId]
	this.mutex.Unlock()
	if ok {
		rb.RId = rid
		rb.DB.Sync()

		this.mutex.Lock()
		if _, ok := this.roleRB[rid]; ok == false{
			this.roleRB[rid] = make([]*model.MapRoleBuild, 0)
		}
		this.roleRB[rid] = append(this.roleRB[rid], rb)
		this.mutex.Unlock()
		return rb, true

	}else{

		if b, ok := NMMgr.PositionBuild(x, y); ok {
			if cfg, _ := BCMgr.BuildConfig(b.Type, b.Level); cfg != nil {
				rb := &model.MapRoleBuild{RId: rid, X: x, Y: y,
					Type: b.Type, Level: b.Level, Name: cfg.Name,
					Wood: cfg.Wood, Iron: cfg.Iron, Stone: cfg.Stone,
					Grain: cfg.Grain, CurDurable: cfg.Durable,
					MaxDurable: cfg.Durable}

				if _, err := db.MasterDB.Table(model.MapRoleBuild{}).Insert(rb); err == nil{
					this.mutex.Lock()
					this.posRB[posId] = rb
					this.dbRB[rb.Id] = rb
					if _, ok := this.roleRB[rid]; ok == false{
						this.roleRB[rid] = make([]*model.MapRoleBuild, 0)
					}
					this.roleRB[rid] = append(this.roleRB[rid], rb)
					this.mutex.Unlock()
					return rb, true
				}else{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}
		}
	}
	return nil, false
}

func (this* RoleBuildMgr) RemoveFromRole(build *model.MapRoleBuild)  {
	this.mutex.Lock()
	rb,ok := this.roleRB[build.RId]
	if ok {
		for i, v := range rb {
			if v.Id == build.Id{
				rb = append(rb[:i], rb[i+1:]...)
				break
			}
		}
	}
	this.mutex.Unlock()

	build.RId = 0
	build.DB.Sync()
}

func (this* RoleBuildMgr) GetRoleBuild(rid int) ([]*model.MapRoleBuild, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	ra, ok := this.roleRB[rid]
	return ra, ok
}

func (this* RoleBuildMgr) Scan(x, y int) []*model.MapRoleBuild {
	if x < 0 || x >= MapWith || y < 0 || y >= MapHeight {
		return nil
	}

	this.mutex.RLock()
	defer this.mutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(MapWith, x+ScanWith)
	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(MapHeight, y+ScanHeight)

	rb := make([]*model.MapRoleBuild, 0)
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			posId := ToPosition(i, j)
			v, ok := this.posRB[posId]
			if ok && v.RId != 0 {
				rb = append(rb, v)
			}
		}
	}

	return rb
}

func (this* RoleBuildMgr) ScanBlock(x, y, length int) []*model.MapRoleBuild {
	if x < 0 || x >= MapWith || y < 0 || y >= MapHeight {
		return nil
	}


	this.mutex.RLock()
	defer this.mutex.RUnlock()

	maxX := util.MinInt(MapWith, x+length-1)
	maxY := util.MinInt(MapHeight, y+length-1)

	rb := make([]*model.MapRoleBuild, 0)
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			posId := ToPosition(i, j)
			v, ok := this.posRB[posId]
			if ok && v.RId != 0 {
				rb = append(rb, v)
			}
		}
	}

	return rb
}