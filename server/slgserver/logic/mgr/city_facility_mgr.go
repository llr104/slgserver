package mgr

import (
	"encoding/json"
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/static_conf/facility"
	"sync"
	"time"
)

var RFMgr = facilityMgr{
	facilities: make(map[int]*model.CityFacility),
	facilitiesByRId: make(map[int][]*model.CityFacility),
}

type facilityMgr struct {
	mutex sync.RWMutex
	facilities map[int]*model.CityFacility
	facilitiesByRId map[int][]*model.CityFacility	//key:rid
}

func (this*facilityMgr) Load() {

	err := db.MasterDB.Find(this.facilities)
	if err != nil {
		log.DefaultLog.Error("facilityMgr load city_facility table error")
	}

	for _, cityFacility := range this.facilities {
		rid := cityFacility.RId
		_, ok := this.facilitiesByRId[rid]
		if ok == false {
			this.facilitiesByRId[rid] = make([]*model.CityFacility, 0)
		}
		this.facilitiesByRId[rid] = append(this.facilitiesByRId[rid], cityFacility)
	}

}
func (this*facilityMgr) GetByRId(rid int) ([]*model.CityFacility, bool){
	this.mutex.RLock()
	r, ok := this.facilitiesByRId[rid]
	this.mutex.RUnlock()
	return r, ok
}

func (this*facilityMgr) Get(cid int) (*model.CityFacility, bool){
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

func (this*facilityMgr) GetFacility(cid int, fType int8) (*model.Facility, bool){
	cf, ok := this.Get(cid)
	if ok == false{
		return nil, false
	}

	facilities := cf.Facility()
	for _, v := range facilities {
		if v.Type == fType{
			return &v, true
		}
	}
	return nil, false
}

func (this*facilityMgr) GetFacilityLv(cid int, fType int8) int8{
	f, ok := this.GetFacility(cid, fType)
	if ok {
		return f.GetLevel()
	}else{
		return 0
	}
}

/*
获取城内设施加成
*/
func (this*facilityMgr) GetAdditions(cid int, additionType... int8 ) []int{
	cf, ok := this.Get(cid)
	ret := make([]int, len(additionType))
	if ok == false{
		return ret
	}

	for i, at := range additionType {
		total := 0
		facilities := cf.Facility()
		for _, f := range facilities {
			if f.GetLevel() > 0{
				adds := facility.FConf.GetAdditions(f.Type)
				values := facility.FConf.GetValues(f.Type, f.GetLevel())

				for i, add := range adds {
					if add == at {
						total += values[i]
					}
				}
			}
		}
		ret[i] = total
	}

	return ret
}
/*
如果不存在尝试去创建
*/
func (this*facilityMgr) GetAndTryCreate(cid, rid int) (*model.CityFacility, bool){
	r, ok := this.Get(cid)
	if ok {
		return r, true
	}else{
		if _, ok:= RCMgr.Get(cid); ok {
			//创建
			fs := make([]model.Facility, len(facility.FConf.List))

			for i, v := range facility.FConf.List {
				f := model.Facility{Type: v.Type, PrivateLevel: 0, Name: v.Name}
				fs[i] = f
			}

			sdata, _ := json.Marshal(fs)
			cf := &model.CityFacility{CityId: cid, RId: rid, Facilities: string(sdata)}
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

func (this*facilityMgr) UpFacility(rid, cid int, fType int8) (*model.Facility, int){
	this.mutex.RLock()
	f, ok := this.facilities[cid]
	this.mutex.RUnlock()

	if ok == false{
		log.DefaultLog.Warn("UpFacility cityId not found",
			zap.Int("cityId", cid),
			zap.Int("type", int(fType)))
		return nil, constant.CityNotExist
	}else{
		facilities := make([]*model.Facility, 0)
		var out *model.Facility
		json.Unmarshal([]byte(f.Facilities), &facilities)
		for _, fac := range facilities {
			if fac.Type == fType {
				maxLevel := facility.FConf.MaxLevel(fType)
				if  fac.CanLV() == false {
					//正在升级中了
					log.DefaultLog.Warn("UpFacility error because already in up",
						zap.Int("curLevel", int(fac.GetLevel())),
						zap.Int("maxLevel", int(maxLevel)),
						zap.Int("cityId", cid),
						zap.Int("type", int(fType)))
					return nil, constant.UpError
				}else if fac.GetLevel() >= maxLevel {
					log.DefaultLog.Warn("UpFacility error",
						zap.Int("curLevel", int(fac.GetLevel())),
						zap.Int("maxLevel", int(maxLevel)),
						zap.Int("cityId", cid),
						zap.Int("type", int(fType)))
					return nil, constant.UpError
				}else{
					need, ok := facility.FConf.Need(fType, fac.GetLevel()+1)
					if ok == false {
						log.DefaultLog.Warn("UpFacility Need config error",
							zap.Int("curLevel", int(fac.GetLevel())),
							zap.Int("cityId", cid),
							zap.Int("type", int(fType)))
						return nil, constant.UpError
					}

					code := RResMgr.TryUseNeed(rid, *need)
					if code == constant.OK {
						fac.UpTime = time.Now().Unix()
						out = fac
						if t, err := json.Marshal(facilities); err == nil{
							f.Facilities = string(t)
							f.SyncExecute()
							return out, constant.OK
						}else{
							return nil, constant.UpError
						}
					}else{
						log.DefaultLog.Warn("UpFacility Need Res Not Enough",
							zap.Int("curLevel", int(fac.GetLevel())),
							zap.Int("cityId", cid),
							zap.Int("type", int(fType)))
						return nil, code
					}
				}
			}
		}

		log.DefaultLog.Warn("UpFacility error not found type",
			zap.Int("cityId", cid),
			zap.Int("type", int(fType)))
		return nil, constant.UpError
	}
}


func (this*facilityMgr) GetYield(rid int)model.Yield{
	cfs, ok := this.GetByRId(rid)
	var y model.Yield
	if ok {
		for _, cf := range cfs {
			for _, f := range cf.Facility() {
				if f.GetLevel() > 0{
					values := facility.FConf.GetValues(f.Type, f.GetLevel())
					additions := facility.FConf.GetAdditions(f.Type)
					for i, aType := range additions {
						if aType == facility.TypeWood {
							y.Wood += values[i]
						}else if aType == facility.TypeGrain {
							y.Grain += values[i]
						}else if aType == facility.TypeIron {
							y.Iron += values[i]
						}else if aType == facility.TypeStone {
							y.Stone += values[i]
						}else if aType == facility.TypeTax {
							y.Gold += values[i]
						}
					}
				}
			}
		}
	}
	return y
}

func (this*facilityMgr) GetDepotCapacity(rid int)int{
	cfs, ok := this.GetByRId(rid)
	limit := 0
	if ok {
		for _, cf := range cfs {
			for _, f := range cf.Facility() {
				if f.GetLevel() > 0{
					values := facility.FConf.GetValues(f.Type, f.GetLevel())
					additions := facility.FConf.GetAdditions(f.Type)
					for i, aType := range additions {
						if aType == facility.TypeWarehouseLimit {
							limit += values[i]
						}
					}
				}
			}
		}
	}
	return limit
}

func (this*facilityMgr) GetCost(cid int) int8{
	cf, ok := this.Get(cid)
	limit := 0
	if ok {
		for _, f := range cf.Facility() {
			if f.GetLevel() > 0{
				values := facility.FConf.GetValues(f.Type, f.GetLevel())
				additions := facility.FConf.GetAdditions(f.Type)
				for i, aType := range additions {
					if aType == facility.TypeCost {
						limit += values[i]
					}
				}
			}
		}
	}
	return int8(limit)
}

func (this*facilityMgr) GetMaxDurable(cid int) int{
	cf, ok := this.Get(cid)
	limit := 0
	if ok {
		for _, f := range cf.Facility() {
			if f.GetLevel() > 0{
				values := facility.FConf.GetValues(f.Type, f.GetLevel())
				additions := facility.FConf.GetAdditions(f.Type)
				for i, aType := range additions {
					if aType == facility.TypeDurable {
						limit += values[i]
					}
				}
			}
		}
	}
	return limit
}

func (this*facilityMgr) GetCityLV(cid int) int8{
	return this.GetFacilityLv(cid, facility.Main)
}



