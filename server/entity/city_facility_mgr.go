package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

type Facility struct {
	Name	string		`json:"name"`
	MLevel	int8		`json:"mLevel"`
	CLevel	int8		`json:"cLevel"`
	Type	int8		`json:"type"`
}

var RFMgr = FacilityMgr{
	facilities: make(map[int]*model.CityFacility),
}

type FacilityMgr struct {
	mutex sync.RWMutex
	facilities map[int]*model.CityFacility
}

func (this* FacilityMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.facilities)
	if err != nil {
		log.DefaultLog.Error("FacilityMgr load city_facility table error")
	}
}

func (this* FacilityMgr) Get(cid int) (*model.CityFacility, error){
	this.mutex.RLock()
	r, ok := this.facilities[cid]
	if ok {
		return r, nil
	}
	this.mutex.RUnlock()

	r = &model.CityFacility{}
	ok, err := db.MasterDB.Table(r).Where("cid=?", cid).Get(r)
	if ok {
		this.mutex.Lock()
		this.facilities[cid] = r
		this.mutex.Unlock()
		return r, nil
	}else{
		if err != nil{
			return nil, err
		}else{
			str := fmt.Sprintf("cid:%d CityFacility not found", cid)
			return nil, errors.New(str)
		}
	}
}

/*
如果不存在尝试去创建
*/
func (this* FacilityMgr) GetAndTryCreate(cid int) (*model.CityFacility, error){
	r, err := this.Get(cid)
	if err == nil {
		return r, nil
	}else{
		if _, err:= RCMgr.Get(cid); err == nil {
			//创建
			fs := make([]Facility, 30)
			for i := 0; i <30 ; i++ {
				f := Facility{Type: int8(i+1), CLevel: int8(1), MLevel: int8(10)}
				fs[i] = f
			}
			sdata, _ := json.Marshal(fs)
			cf := &model.CityFacility{CityId: cid, Facilities: string(sdata)}
			db.MasterDB.Table(cf).Insert(cf)

			this.mutex.Lock()
			this.facilities[cid] = cf
			this.mutex.Unlock()

			return cf, nil
		}else{
			str := fmt.Sprintf("cid:%d not found", cid)
			return nil, errors.New(str)
		}
	}
}