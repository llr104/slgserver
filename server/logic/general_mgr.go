package logic

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
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
	go this.toDatabase()
}
func (this* GeneralMgr) toDatabase() {
	for true {
		time.Sleep(5*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.genByGId {
			if v.NeedUpdate {
				cnt+=1
				_, err := db.MasterDB.Table(model.General{}).Cols( "force", "strategy",
					"defense", "speed", "destroy", "level", "exp", "order", "cityId").Update(v)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}else{
					v.NeedUpdate = false
				}
			}

			//一次最多更新20个
			if cnt>20{
				break
			}
		}

		this.mutex.RUnlock()
	}
}

func (this* GeneralMgr) Get(rid int) ([]*model.General, error){
	this.mutex.Lock()
	r, ok := this.genByRole[rid]
	this.mutex.Unlock()

	if ok {
		return r, nil
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
			return m, nil
		}else{
			return nil, errors.New(fmt.Sprintf("rid: %d general not fount", rid))
		}
	}else{
		return nil, err
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

/*
如果不存在尝试去创建
*/
func (this* GeneralMgr) GetAndTryCreate(rid int) ([]*model.General, bool){
	r, err := this.Get(rid)
	if err == nil {
		return r, true
	}else{
		//创建
		gs := make([]*model.General, 0)
		sess := db.MasterDB.NewSession()
		sess.Begin()

		for _, v := range general.General.List {
			r := &model.General{RId: rid, Name: v.Name, CfgId: v.CfgId,
				Force: v.Force, Strategy: v.Strategy, Defense: v.Defense,
				Speed: v.Speed, Cost: v.Cost, Order: 0, CityId: 0,
				Level: 1, CreatedAt: time.Now(),
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
