package entity

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
func (this* GeneralMgr) FindGeneral(gid int) (*model.General, error){
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	g, ok := this.genByGId[gid]
	if ok {
		return g, nil
	}else{
		str := fmt.Sprintf("general %d not found", gid)
		log.DefaultLog.Warn(str, zap.Int("gid", gid))
		return nil, errors.New(str)
	}
}

/*
如果不存在尝试去创建
*/
func (this* GeneralMgr) GetAndTryCreate(rid int) ([]*model.General, error){
	r, err := this.Get(rid)
	if err == nil {
		return r, nil
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
				return nil, err
			}
		}
		if err := sess.Commit(); err != nil{
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, err
		}else{
			this.mutex.Lock()
			this.genByRole[rid] = gs
			for _, v := range gs {
				this.genByGId[v.Id] = v
			}
			this.mutex.Unlock()
			return gs, nil
		}
	}
}
