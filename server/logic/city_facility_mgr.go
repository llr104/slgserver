package logic

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/server/static_conf/facility"
	"sync"
	"time"
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

	go this.toDatabase()
}

func (this* FacilityMgr) Get(cid int) (*model.CityFacility, bool){
	this.mutex.RLock()
	r, ok := this.facilities[cid]
	this.mutex.RUnlock()

	if ok {
		return r, true
	}

	r = &model.CityFacility{}
	ok, err := db.MasterDB.Table(r).Where("cityId=?", cid).Get(r)
	log.DefaultLog.Warn("db error", zap.Error(err))
	if ok {
		this.mutex.Lock()
		this.facilities[cid] = r
		this.mutex.Unlock()
		return r, true
	}else{
		return nil, false
	}
}

/*
如果不存在尝试去创建
*/
func (this* FacilityMgr) GetAndTryCreate(cid int) (*model.CityFacility, bool){
	r, ok := this.Get(cid)
	if ok {
		return r, true
	}else{
		if _, ok:= RCMgr.Get(cid); ok {
			//创建
			fs := make([]Facility, len(facility.FConf.List))

			for i, v := range facility.FConf.List {
				f := Facility{Type: v.Type, Level: int8(1), Name: v.Name}
				fs[i] = f
			}

			sdata, _ := json.Marshal(fs)
			cf := &model.CityFacility{CityId: cid, Facilities: string(sdata)}
			db.MasterDB.Table(cf).Insert(cf)

			this.mutex.Lock()
			this.facilities[cid] = cf
			this.mutex.Unlock()

			return cf, true
		}else{
			log.DefaultLog.Warn("cid not found", zap.Int("cid", cid))
			return nil, false
		}
	}
}

func (this* FacilityMgr) UpFacility(rid, cid int, fType int8) (*Facility, int){
	this.mutex.Lock()
	defer this.mutex.Unlock()
	f, ok := this.facilities[cid]
	if ok == false{
		str := fmt.Sprintf("UpFacility cityId %d not found", cid)
		log.DefaultLog.Warn(str)
		return nil, constant.CityNotExist
	}else{
		suss := false
		fa := make([]*Facility, 0)
		var out *Facility
		json.Unmarshal([]byte(f.Facilities), &fa)
		for _, v := range fa {

			if v.Type == fType && v.Level < facility.FConf.MaxLevel(fType){
				need, err := facility.FConf.Need(fType, int(v.Level+1))
				if err != nil {
					break
				}

				if RResMgr.TryUseNeed(rid, need) {
					v.Level += 1
					suss = true
					out = v
					f.NeedUpdate = true
				}else{
					break
				}
			}
		}
		if suss {
			if t, err := json.Marshal(fa); err == nil{
				f.Facilities = string(t)
				return out, constant.OK
			}else{
				return nil, constant.UpError
			}
		}else{
			str := fmt.Sprintf("UpFacility error")
			log.DefaultLog.Warn(str)
			return nil, constant.UpError
		}

	}
}
func (this* FacilityMgr) toDatabase() {
	for true {
		time.Sleep(5*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.facilities {
			if v.NeedUpdate {
				_, err := db.MasterDB.Table(v).Cols("facilities").Update(v)
				if err != nil{
					log.DefaultLog.Error("FacilityMgr toDatabase error", zap.Error(err))
				}else{
					v.NeedUpdate = false
				}
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