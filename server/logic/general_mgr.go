package logic

import (
	"go.uber.org/zap"
	"math/rand"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/model"
	"slgserver/server/static_conf"
	"slgserver/server/static_conf/general"
	"slgserver/util"
	"sync"
	"time"
)

type GeneralMgr struct {
	mutex     sync.RWMutex
	genByRole map[int][]*model.General
	genByGId  map[int]*model.General
}

var GMgr = &GeneralMgr{
	genByRole: make(map[int][]*model.General),
	genByGId: make(map[int]*model.General),
}



func (this* GeneralMgr) toDatabase() {
	for true {
		time.Sleep(2*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.genByGId {
			if v.DB.NeedSync() {
				v.DB.BeginSync()
				_, err := db.MasterDB.Table(model.General{}).ID(v.Id).Cols("level",
					"exp", "order", "cityId", "physical_power").Update(v)

				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
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

func (this* GeneralMgr) updatePhysicalPower() {
	limit := static_conf.Basic.General.PhysicalPowerLimit
	recoverCnt := static_conf.Basic.General.RecoveryPhysicalPower
	for true {
		time.Sleep(1*time.Hour)
		this.mutex.RLock()
		for _, g := range this.genByGId {
			if g.PhysicalPower < limit{
				g.PhysicalPower = util.MinInt(limit, g.PhysicalPower+recoverCnt)
				g.SyncExecute()
			}
		}
		this.mutex.RUnlock()
	}
}

func (this* GeneralMgr) createNPC() ([]*model.General, bool){
	//创建
	gs := make([]*model.General, 0)
	sess := db.MasterDB.NewSession()
	sess.Begin()

	for _, v := range general.General.GMap {
		r, ok := this.NewGeneral(v.CfgId, 0)
		if ok == false {
			sess.Rollback()
			return nil, false
		}
		gs = append(gs, r)

	}
	if err := sess.Commit(); err != nil{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, false
	}else{
		return gs, true
	}
}

func (this* GeneralMgr) add(g *model.General) {
	this.mutex.Lock()

	if _,ok := this.genByRole[g.RId]; ok == false{
		this.genByRole[g.RId] = make([]*model.General, 0)
	}
	this.genByRole[g.RId] = append(this.genByRole[g.RId], g)
	this.genByGId[g.Id] = g

	this.mutex.Unlock()
}

func (this* GeneralMgr) Load(){

	err := db.MasterDB.Table(model.General{}).Find(this.genByGId)
	if err != nil {
		log.DefaultLog.Warn("db error", zap.Error(err))
		return
	}

	for _, v := range this.genByGId {
		if _, ok := this.genByRole[v.RId]; ok==false {
			this.genByRole[v.RId] = make([]*model.General, 0)
		}
		this.genByRole[v.RId] = append(this.genByRole[v.RId], v)
	}

	if len(this.genByGId) == 0{
		this.createNPC()
	}

	go this.toDatabase()
	go this.updatePhysicalPower()
}

func (this* GeneralMgr) GetByRId(rid int) ([]*model.General, bool){
	this.mutex.Lock()
	r, ok := this.genByRole[rid]
	this.mutex.Unlock()

	if ok {
		return r, true
	}

	gs := make([]*model.General, 0)
	err := db.MasterDB.Table(new(model.General)).Where("rid=?", rid).Find(&gs)
	if err == nil {
		if len(gs) > 0 {
			for _, g := range gs {
				this.add(g)
			}
			return gs, true
		}else{
			log.DefaultLog.Warn("general not fount", zap.Int("rid", rid))
			return nil, false
		}
	}else{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, false
	}
}

//查找将领
func (this* GeneralMgr) GetByGId(gid int) (*model.General, bool){
	this.mutex.RLock()
	g, ok := this.genByGId[gid]
	this.mutex.RUnlock()
	if ok {
		return g, true
	}else{

		g := &model.General{}
		r, err := db.MasterDB.Table(new(model.General)).Where("id=?", gid).Get(g)
		if err == nil{
			if r {
				this.add(g)
				return g, true
			}else{
				log.DefaultLog.Warn("general gid not found", zap.Int("gid", gid))
				return nil, false
			}

		}else{
			log.DefaultLog.Warn("general gid not found", zap.Int("gid", gid))
			return nil, false
		}
	}
}

func (this* GeneralMgr) NewGeneral(cfgId int, rid int) (*model.General, bool) {
	cfg, ok := general.General.GMap[cfgId]
	if ok {
		g := &model.General{RId: rid, CfgId: cfg.CfgId, Cost: cfg.Cost, Order: 0, CityId: 0,
			PhysicalPower: static_conf.Basic.General.PhysicalPowerLimit,
			Level: 1, CreatedAt: time.Now(),
		}

		if _, err := db.MasterDB.Table(model.General{}).Insert(g); err != nil {
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}else{
			this.add(g)
			return g, true
		}
	}else{
		return nil, false
	}
}

/*
如果不存在尝试去创建
*/
func (this* GeneralMgr) GetByRIdTryCreate(rid int) ([]*model.General, bool){
	r, ok := this.GetByRId(rid)
	if ok {
		return r, true
	}else{
		//创建
		gs := make([]*model.General, 0)
		sess := db.MasterDB.NewSession()
		sess.Begin()

		for _, v := range general.General.GMap {
			g, ok := this.NewGeneral(v.CfgId, rid)
			if ok == false{
				sess.Rollback()
				return nil, false
			}
			gs = append(gs, g)
		}
		if err := sess.Commit(); err != nil{
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}else{
			return gs, true
		}
	}
}




/*
随机创建一个
*/
func (this* GeneralMgr) RandCreateGeneral(rid int,nums int) ([]*model.General, bool){
	//创建
	gs := make([]*model.General, 0)
	sess := db.MasterDB.NewSession()
	sess.Begin()



	for i := 0; i < nums; i++ {
		r := rand.Intn(len(general.General.GArr))
		g, ok := this.NewGeneral(general.General.GArr[r].CfgId, rid)
		if ok == false{
			sess.Rollback()
			return nil, false
		}
		gs = append(gs, g)
	}



	if err := sess.Commit(); err != nil{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, false
	}else{
		return gs, true
	}
}


//获取npc武将
func (this* GeneralMgr) GetNPCGenerals(cnt int) ([]model.General, bool) {
	gs, ok := this.GetByRId(0)
	if ok == false {
		return make([]model.General, 0), false
	}else{
		if cnt > len(gs){
			return make([]model.General, 0), false
		}else{
			m := make(map[int]int)
			for true {
				 r := rand.Intn(len(gs))
				 m[r] = r
				 if len(m) == cnt{
				 	break
				 }
			}
			rgs := make([]model.General, 0)
			for _, v := range m {
				t := *gs[v]
				//npc不需要更新到数据库
				t.DB.Disable(true)
				rgs = append(rgs, t)
			}
			return rgs, true
		}
	}
}

func (this *GeneralMgr) GetDestroy(army *model.Army) int{
	destroy := 0
	for _, gid := range army.GeneralArray {
		g, ok := this.GetByGId(gid)
		if ok {
			destroy += g.GetDestroy()
		}
	}
	return destroy
}

//体力是否足够
func (this* GeneralMgr) PhysicalPowerIsEnough(army *model.Army, cost int) bool{
	for _, gid := range army.GeneralArray {
		if gid == 0{
			continue
		}

		g, ok := this.GetByGId(gid)
		if ok {
			if g.PhysicalPower < cost{
				return false
			}
		}else{
			return false
		}
	}
	return true
}

//尝试使用体力
func (this *GeneralMgr) TryUsePhysicalPower(army *model.Army, cost int) bool{

	if this.PhysicalPowerIsEnough(army, cost) == false{
		return false
	}

	for _, gid := range army.GeneralArray {
		if gid == 0{
			continue
		}

		g, _ := this.GetByGId(gid)
		g.PhysicalPower -= cost
		g.SyncExecute()
	}

	return true
}
