package entity

import (
	"encoding/json"
	"errors"
	"fmt"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/server/static_conf"
	"sync"
)



type Facility struct {
	Name   string `json:"name"`
	Level  int8   `json:"level"`
	Type   int8   `json:"type"`
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
	this.mutex.RUnlock()

	if ok {
		return r, nil
	}

	r = &model.CityFacility{}
	ok, err := db.MasterDB.Table(r).Where("cityId=?", cid).Get(r)
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
			fs := make([]Facility, len(static_conf.FConf.List))

			for i, v := range static_conf.FConf.List {
				f := Facility{Type: int8(i), Level: int8(1), Name: v.Name}
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

func (this* FacilityMgr) UpFacility(cid int, fType int8) (*Facility, error){
	this.mutex.Lock()
	defer this.mutex.Unlock()
	f, ok := this.facilities[cid]
	if ok == false{
		str := fmt.Sprintf("UpFacility cityId %d not found", cid)
		log.DefaultLog.Warn(str)
		return nil, errors.New(str)
	}else{
		suss := false
		fa := make([]*Facility, 0)
		var out *Facility
		json.Unmarshal([]byte(f.Facilities), &fa)
		for _, v := range fa {
			if v.Type == fType && v.Level < static_conf.FConf.MaxLevel(fType){
				v.Level +=1
				suss = true
				out = v
				break
			}
		}
		if suss {
			if t, err := json.Marshal(fa); err == nil{
				f.Facilities = string(t)
				return out, nil
			}else{
				return nil, err
			}
		}else{
			str := fmt.Sprintf("UpFacility error")
			log.DefaultLog.Warn(str)
			return nil, errors.New(str)
		}

	}
}