package mgr

import (
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/global"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/static_conf"
	"slgserver/util"
	"sync"
	"time"
)


type roleBuildMgr struct {
	baseMutex 		sync.RWMutex
	giveUpMutex 	sync.RWMutex
	destroyMutex 	sync.RWMutex
	dbRB  			map[int]*model.MapRoleBuild    	//key:dbId
	posRB 			map[int]*model.MapRoleBuild    	//key:posId
	roleRB 			map[int][]*model.MapRoleBuild 	//key:roleId
	giveUpRB 		map[int64]map[int]*model.MapRoleBuild //key:time
	destroyRB 		map[int64]map[int]*model.MapRoleBuild //key:time

}


var RBMgr = &roleBuildMgr{
	dbRB: make(map[int]*model.MapRoleBuild),
	posRB: make(map[int]*model.MapRoleBuild),
	roleRB: make(map[int][]*model.MapRoleBuild),
	giveUpRB: make(map[int64]map[int]*model.MapRoleBuild),
	destroyRB: make(map[int64]map[int]*model.MapRoleBuild),
}

func (this*roleBuildMgr) Load() {

	if total, err := db.MasterDB.Where("type = ?",
		model.MapBuildSysCity).Count(new(model.MapRoleBuild)); err != nil{
		log.DefaultLog.Panic("db error")
	}else{
		//初始化系统城池进数据库
		if int64(len(NMMgr.sysCity)) != total{
			db.MasterDB.Where("type = ?",
				model.MapBuildSysCity).Delete(new(model.MapRoleBuild))
			for _, sysCity := range NMMgr.sysCity {

				build := model.MapRoleBuild{
					RId: 0,
					Type: sysCity.Type,
					Level: sysCity.Level,
					X: sysCity.X,
					Y: sysCity.Y,
				}
				build.ConvertToRes()
				db.MasterDB.InsertOne(&build)
			}
		}
	}


	err := db.MasterDB.Find(this.dbRB)
	if err != nil {
		log.DefaultLog.Error("roleBuildMgr load role_build table error", zap.Error(err))
	}

	curTime := time.Now().Unix()

	//转成posRB 和 roleRB
	for _, v := range this.dbRB {
		v.Init()

		//恢复正在放弃的土地
		if v.GiveUpTime != 0 {
			_, ok := this.giveUpRB[v.GiveUpTime]
			if ok == false{
				this.giveUpRB[v.GiveUpTime] = make(map[int]*model.MapRoleBuild)
			}
			this.giveUpRB[v.GiveUpTime][v.Id] = v
		}

		//恢复正在拆除的建筑
		if v.OPLevel == 0 && v.Level != v.OPLevel {
			t := v.EndTime.Unix()
			if curTime >= t {
				v.ConvertToRes()
			}else{
				_, ok := this.destroyRB[t]
				if ok == false{
					this.destroyRB[t] = make(map[int]*model.MapRoleBuild)
				}
				this.destroyRB[t][v.Id] = v
			}
		}

		posId := global.ToPosition(v.X, v.Y)
		this.posRB[posId] = v
		_,ok := this.roleRB[v.RId]
		if ok == false{
			this.roleRB[v.RId] = make([]*model.MapRoleBuild, 0)
		}
		this.roleRB[v.RId] = append(this.roleRB[v.RId], v)

		//过滤掉到了放弃时间的领地
		if v.GiveUpTime != 0 && v.GiveUpTime <= curTime{
			this.RemoveFromRole(v)
		}

	}

}

//检测正在放弃的土地是否到期了
func (this*roleBuildMgr) CheckGiveUp() []int {
	var ret []int
	var builds []*model.MapRoleBuild

	curTime := time.Now().Unix()
	this.giveUpMutex.Lock()
	for i := curTime-10; i <= curTime ; i++ {
		gs, ok := this.giveUpRB[i]
		if ok {
			for _, g := range gs {
				builds = append(builds, g)
				ret = append(ret, global.ToPosition(g.X, g.Y))
			}
		}
	}
	this.giveUpMutex.Unlock()

	for _, build := range builds {
		this.RemoveFromRole(build)
	}

	return ret
}

//检测正在拆除的建筑是否到期
func (this*roleBuildMgr) CheckDestroy() []int {
	var ret []int
	var builds []*model.MapRoleBuild

	curTime := time.Now().Unix()
	this.destroyMutex.Lock()
	for i := curTime-10; i <= curTime ; i++ {
		gs, ok := this.destroyRB[i]
		if ok {
			for _, g := range gs {
				builds = append(builds, g)
				ret = append(ret, global.ToPosition(g.X, g.Y))
			}
		}
	}
	this.destroyMutex.Unlock()

	for _, build := range builds {
		build.ConvertToRes()
		build.SyncExecute()
	}
	return ret
}
/*
该位置是否被角色占领
*/
func (this*roleBuildMgr) IsEmpty(x, y int) bool {
	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()
	posId := global.ToPosition(x, y)
	_, ok := this.posRB[posId]
	return !ok
}

func (this*roleBuildMgr) PositionBuild(x, y int) (*model.MapRoleBuild, bool) {
	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()
	posId := global.ToPosition(x, y)
	b,ok := this.posRB[posId]
	if ok && b.RId != 0 {
		return b, ok
	}else{
		return nil, false
	}
}

func (this*roleBuildMgr) RoleFortressCnt(rid int)int{
	bs, ok := this.GetRoleBuild(rid)
	cnt := 0
	if ok == false {
		return 0
	}else{
		for _, b := range bs {
			if b.IsRoleFortress(){
				cnt +=1
			}
		}
	}
	return cnt
}


func (this*roleBuildMgr) AddBuild(rid, x, y int) (*model.MapRoleBuild, bool) {

	posId := global.ToPosition(x, y)
	this.baseMutex.Lock()
	rb, ok := this.posRB[posId]
	this.baseMutex.Unlock()
	if ok {
		rb.RId = rid
		this.baseMutex.Lock()
		if _, ok := this.roleRB[rid]; ok == false{
			this.roleRB[rid] = make([]*model.MapRoleBuild, 0)
		}
		this.roleRB[rid] = append(this.roleRB[rid], rb)
		this.baseMutex.Unlock()
		return rb, true

	}else{

		if b, ok := NMMgr.PositionBuild(x, y); ok {
			if cfg, _ := static_conf.MapBuildConf.BuildConfig(b.Type, b.Level); cfg != nil {
				rb := &model.MapRoleBuild{
					RId: rid, X: x, Y: y,
					Type: b.Type, Level: b.Level, OPLevel: b.Level,
					Name: cfg.Name, CurDurable: cfg.Durable,
					MaxDurable: cfg.Durable,
				}
				rb.Init()

				if _, err := db.MasterDB.Table(model.MapRoleBuild{}).Insert(rb); err == nil{
					this.baseMutex.Lock()
					this.posRB[posId] = rb
					this.dbRB[rb.Id] = rb
					if _, ok := this.roleRB[rid]; ok == false{
						this.roleRB[rid] = make([]*model.MapRoleBuild, 0)
					}
					this.roleRB[rid] = append(this.roleRB[rid], rb)
					this.baseMutex.Unlock()
					return rb, true
				}else{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}
		}
	}
	return nil, false
}

func (this*roleBuildMgr) RemoveFromRole(build *model.MapRoleBuild)  {
	this.baseMutex.Lock()
	rb,ok := this.roleRB[build.RId]
	if ok {
		for i, v := range rb {
			if v.Id == build.Id{
				this.roleRB[build.RId] = append(rb[:i], rb[i+1:]...)
				break
			}
		}
	}
	this.baseMutex.Unlock()

	t := build.EndTime.Unix()
	//移除放弃事件
	this.giveUpMutex.Lock()
	if ms, ok := this.giveUpRB[t]; ok{
		delete(ms, build.Id)
	}
	this.giveUpMutex.Unlock()

	//移除拆除事件
	this.destroyMutex.Lock()
	if ms, ok := this.destroyRB[t]; ok{
		delete(ms, build.Id)
	}
	this.destroyMutex.Unlock()

	build.Reset()
	build.SyncExecute()
}

func (this*roleBuildMgr) GetRoleBuild(rid int) ([]*model.MapRoleBuild, bool) {
	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()
	ra, ok := this.roleRB[rid]
	return ra, ok
}

func (this*roleBuildMgr) BuildCnt(rid int) int {
	bs, ok := this.GetRoleBuild(rid)
	if ok {
		return len(bs)
	}else{
		return 0
	}
}

func (this*roleBuildMgr) Scan(x, y int) []*model.MapRoleBuild {
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}

	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(global.MapWith, x+ScanWith)
	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(global.MapHeight, y+ScanHeight)

	rb := make([]*model.MapRoleBuild, 0)
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			posId := global.ToPosition(i, j)
			v, ok := this.posRB[posId]
			if ok && v.RId != 0 {
				rb = append(rb, v)
			}
		}
	}

	return rb
}

func (this*roleBuildMgr) ScanBlock(x, y, length int) []*model.MapRoleBuild {
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}


	this.baseMutex.RLock()
	defer this.baseMutex.RUnlock()

	maxX := util.MinInt(global.MapWith, x+length-1)
	maxY := util.MinInt(global.MapHeight, y+length-1)

	rb := make([]*model.MapRoleBuild, 0)
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			posId := global.ToPosition(i, j)
			v, ok := this.posRB[posId]
			if ok && v.RId != 0 {
				rb = append(rb, v)
			}
		}
	}

	return rb
}

func (this*roleBuildMgr) BuildIsRId(x, y, rid int) bool {
	b, ok := this.PositionBuild(x, y)
	if ok {
		return b.RId == rid
	}else{
		return false
	}
}

func (this*roleBuildMgr) GetYield(rid int)model.Yield{
	builds, ok := this.GetRoleBuild(rid)
	var y model.Yield
	if ok {
		for _, b := range builds {
			y.Iron += b.Iron
			y.Wood += b.Wood
			y.Grain += b.Grain
			y.Stone += b.Grain
		}
	}
	return y
}

func (this* roleBuildMgr) GiveUp(x, y int) int {
	b, ok := this.PositionBuild(x, y)
	if ok == false{
		return constant.CannotGiveUp
	}

	if b.IsWarFree() {
		return constant.BuildWarFree
	}

	if b.GiveUpTime > 0{
		return constant.BuildGiveUpAlready
	}

	b.GiveUpTime = time.Now().Unix() + static_conf.Basic.Build.GiveUpTime
	b.SyncExecute()

	this.giveUpMutex.Lock()
	_, ok = this.giveUpRB[b.GiveUpTime]
	if ok == false {
		this.giveUpRB[b.GiveUpTime] = make(map[int]*model.MapRoleBuild)
	}
	this.giveUpRB[b.GiveUpTime][b.Id] = b
	this.giveUpMutex.Unlock()

	return constant.OK
}

func (this* roleBuildMgr) Destroy(x, y int) int {

	b, ok := this.PositionBuild(x, y)
	if ok == false {
		return constant.BuildNotMe
	}

	if b.IsHaveModifyLVAuth() == false || b.IsInGiveUp() || b.IsBusy() {
		return constant.CanNotDestroy
	}

	cfg, ok := static_conf.MapBCConf.BuildConfig(b.Type, b.Level)
	if ok == false{
		return constant.InvalidParam
	}

	code := RResMgr.TryUseNeed(b.RId, cfg.Need)
	if code != constant.OK {
		return code
	}

	b.EndTime = time.Now().Add(time.Duration(cfg.Time)*time.Second)
	this.destroyMutex.Lock()
	t := b.EndTime.Unix()
	_, ok = this.destroyRB[t]
	if ok == false {
		this.destroyRB[t] = make(map[int]*model.MapRoleBuild)
	}
	this.destroyRB[t][b.Id] = b
	this.destroyMutex.Unlock()

	b.DelBuild(*cfg)
	b.SyncExecute()

	return constant.OK
}