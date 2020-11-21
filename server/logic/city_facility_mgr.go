package logic

import (
	"encoding/json"
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
	if err != nil{
		log.DefaultLog.Warn("db error", zap.Error(err))
	}

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
		log.DefaultLog.Warn("UpFacility cityId not found", zap.Int("cityId", cid), zap.Int("type", int(fType)))
		return nil, constant.CityNotExist
	}else{
		fa := make([]*Facility, 0)
		var out *Facility
		json.Unmarshal([]byte(f.Facilities), &fa)
		for _, v := range fa {
			if v.Type == fType {
				maxLevel := facility.FConf.MaxLevel(fType)
				if v.Level >= maxLevel{
					log.DefaultLog.Warn("UpFacility error",
						zap.Int("curLevel", int(v.Level)), zap.Int("maxLevel", int(maxLevel)), zap.Int("cityId", cid), zap.Int("type", int(fType)))
					return nil, constant.UpError
				}else{
					need, ok := facility.FConf.Need(fType, int(v.Level+1))
					if ok == false {
						log.DefaultLog.Warn("UpFacility Need config error",
							zap.Int("curLevel", int(v.Level)), zap.Int("cityId", cid), zap.Int("type", int(fType)))
						return nil, constant.UpError
					}
					if RResMgr.TryUseNeed(rid, need) {
						v.Level += 1
						out = v
						f.NeedUpdate = true
						if t, err := json.Marshal(fa); err == nil{
							f.Facilities = string(t)
							return out, constant.OK
						}else{
							return nil, constant.UpError
						}
					}else{
						log.DefaultLog.Warn("UpFacility Need Res Not Enough",
							zap.Int("curLevel", int(v.Level)), zap.Int("cityId", cid), zap.Int("type", int(fType)))
						return nil, constant.ResNotEnough
					}
				}
			}
		}

		log.DefaultLog.Warn("UpFacility error not found type", zap.Int("cityId", cid), zap.Int("type", int(fType)))
		return nil, constant.UpError
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