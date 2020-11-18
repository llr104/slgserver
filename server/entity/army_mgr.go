package entity

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

type ArmyMgr struct {
	mutex     sync.RWMutex
	armyById  map[int]*model.Army
	armByCityId map[int][]*model.Army

}

var AMgr = &ArmyMgr{
	armyById: make(map[int]*model.Army),
	armByCityId: make(map[int][]*model.Army),
}

func (this* ArmyMgr) Get(aid int) (*model.Army, error){
	this.mutex.RLock()
	a, ok := this.armyById[aid]
	this.mutex.RUnlock()
	if ok {
		return a, nil
	}else{
		army := &model.Army{}
		ok, err := db.MasterDB.Table(model.Army{}).Where("id=?", aid).Get(army)
		if ok {
			this.mutex.Lock()
			this.armyById[aid] = army
			if _, r:= this.armByCityId[army.CityId]; r == false{
				this.armByCityId[army.CityId] = make([]*model.Army, 0)
			}
			this.armByCityId[army.CityId] = append(this.armByCityId[army.CityId], army)
			this.mutex.Unlock()
			return army, nil
		}else{
			return nil, err
		}
	}
}

func (this* ArmyMgr) GetByCity(cid int) ([]*model.Army, error){
	this.mutex.RLock()
	a,ok := this.armByCityId[cid]
	this.mutex.RUnlock()
	if ok {
		return a, nil
	}else{
		m := make([]*model.Army, 0)
		err := db.MasterDB.Table(model.Army{}).Where("cityId=?", cid).Find(&m)
		if err!=nil{
			log.DefaultLog.Warn("ArmyMgr GetByCity db error", zap.Int("cityId", cid))
			return m, err
		}else{
			this.mutex.Lock()
			this.armByCityId[cid] = m
			this.mutex.Unlock()
			return m, nil
		}
	}
}

func (this* ArmyMgr) GetOrCreate(rid int, cid int, order int8) (*model.Army, error){

	this.mutex.RLock()
	armys, ok := this.armByCityId[cid]
	this.mutex.RUnlock()

	if ok {
		for _, v := range armys {
			if v.Order == order{
				return v, nil
			}
		}
	}

	//需要创建
	a := &model.Army{RId: rid, Order: order, CityId: cid,
		FirstId: 0, SecondId: 0, ThirdId: 0,
		FirstSoldierCnt: 0, SecondSoldierCnt: 0, ThirdSoldierCnt: 0}
	_, err := db.MasterDB.Insert(a)
	if err == nil{
		return a, nil
	}else{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, err
	}
}


