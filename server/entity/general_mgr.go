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
	mutex sync.RWMutex
	gen   map[int][]*model.General
}

var GMgr = &GeneralMgr{
	gen: make(map[int][]*model.General),
}

func (this* GeneralMgr) Get(rid int) ([]*model.General, error){
	this.mutex.Lock()
	r, ok := this.gen[rid]
	this.mutex.Unlock()

	if ok {
		return r, nil
	}

	m := make([]*model.General, 0)
	err := db.MasterDB.Table(new(model.General)).Where("rid=?", rid).Find(&m)
	if err == nil {
		if len(m) > 0 {
			this.mutex.Lock()
			this.gen[rid] = m
			this.mutex.Unlock()
			return m, nil
		}else{
			return nil, errors.New(fmt.Sprintf("rid: %d general not fount", rid))
		}
	}else{
		return nil, err
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
				Speed: v.Speed, Cost: v.Cost, ArmyId: -1, CityId: -1,
				CreatedAt: time.Now(),
			}
			gs = append(gs, r)

			if _, err := db.MasterDB.Table(model.General{}).Insert(r); err != nil {
				sess.Rollback()
				log.DefaultLog.Warn("db error", zap.Error(err))
				return nil, err
			}
		}
		sess.Commit()
		return gs, nil
	}
}
