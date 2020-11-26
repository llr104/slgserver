package logic

import (
	"go.uber.org/zap"
	"math/rand"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/server/static_conf"
	"slgserver/server/static_conf/general"
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

}
func (this* GeneralMgr) toDatabase() {
	for true {
		time.Sleep(2*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.genByGId {
			if v.DB.NeedSync() {
				v.DB.BeginSync()
				_, err := db.MasterDB.Table(model.General{}).ID(v.Id).Cols( "force", "strategy",
					"defense", "speed", "destroy", "level", "exp", "order", "cityId").Update(v)
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

func (this* GeneralMgr) Get(rid int) ([]*model.General, bool){
	this.mutex.Lock()
	r, ok := this.genByRole[rid]
	this.mutex.Unlock()

	if ok {
		return r, true
	}

	m := make([]*model.General, 0)
	err := db.MasterDB.Table(new(model.General)).Where("rid=?", rid).Find(&m)
	if err == nil {
		if len(m) > 0 {
			this.mutex.Lock()
			this.genByRole[rid] = m
			for _, v := range m {
				this.genByGId[v.Id] = v
			}
			this.mutex.Unlock()
			return m, true
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
func (this* GeneralMgr) FindGeneral(gid int) (*model.General, bool){
	this.mutex.RLock()
	g, ok := this.genByGId[gid]
	this.mutex.RUnlock()
	if ok {
		return g, true
	}else{

		m := &model.General{}
		r, err := db.MasterDB.Table(new(model.General)).Where("id=?", gid).Get(m)
		if err == nil{
			if r {
				this.mutex.Lock()
				this.genByGId[m.Id] = m

				if rg,ok := this.genByRole[m.RId];ok{
					this.genByRole[m.RId] = append(rg, m)
				}else{
					this.genByRole[m.RId] = make([]*model.General, 0)
					this.genByRole[m.RId] = append(this.genByRole[m.RId], m)
				}

				this.mutex.Unlock()
				return m, true
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
		g := &model.General{RId: rid, Name: cfg.Name, CfgId: cfg.CfgId,
			Force: cfg.Force, Strategy: cfg.Strategy, Defense: cfg.Defense, Speed: cfg.Speed,
			ForceGrow: cfg.ForceGrow, StrategyGrow: cfg.StrategyGrow,
			DefenseGrow: cfg.DefenseGrow, DestroyGrow: cfg.DestroyGrow, SpeedGrow: cfg.SpeedGrow,
			Cost: cfg.Cost, Order: 0, CityId: 0,
			PhysicalPower: static_conf.Basic.General.PhysicalPowerLimit,
			Level: 1, CreatedAt: time.Now(),
		}
		return g, true
	}else{
		return nil, false
	}
}

/*
如果不存在尝试去创建
*/
func (this* GeneralMgr) GetAndTryCreate(rid int) ([]*model.General, bool){
	r, ok := this.Get(rid)
	if ok {
		return r, true
	}else{
		//创建
		gs := make([]*model.General, 0)
		sess := db.MasterDB.NewSession()
		sess.Begin()

		for _, v := range general.General.GMap {
			r, ok := this.NewGeneral(v.CfgId, rid)
			if ok == false{
				continue
			}

			gs = append(gs, r)

			if _, err := db.MasterDB.Table(model.General{}).Insert(r); err != nil {
				sess.Rollback()
				log.DefaultLog.Warn("db error", zap.Error(err))
				return nil, false
			}
		}
		if err := sess.Commit(); err != nil{
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}else{
			this.mutex.Lock()
			this.genByRole[rid] = gs
			for _, v := range gs {
				this.genByGId[v.Id] = v
			}
			this.mutex.Unlock()
			return gs, true
		}
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
			continue
		}
		gs = append(gs, r)
		if _, err := db.MasterDB.Table(model.General{}).Insert(r); err != nil {
			sess.Rollback()
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}
	}
	if err := sess.Commit(); err != nil{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, false
	}else{
		this.mutex.Lock()
		this.genByRole[0] = gs
		for _, v := range gs {
			this.genByGId[v.Id] = v
		}
		this.mutex.Unlock()
		return gs, true
	}
}

//获取npc武将
func (this* GeneralMgr) GetNPCGenerals(cnt int) ([]model.General, bool) {
	gs, ok := this.Get(0)
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
		g, ok := this.FindGeneral(gid)
		if ok {
			destroy += g.Destroy
		}
	}
	return destroy
}
