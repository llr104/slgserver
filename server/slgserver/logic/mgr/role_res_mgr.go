package mgr

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/static_conf"
	"slgserver/server/slgserver/static_conf/facility"
	"slgserver/util"
	"sync"
	"time"
)

type roleResMgr struct {
	mutex  sync.RWMutex
	rolesRes map[int]*model.RoleRes
}

var RResMgr = &roleResMgr{
	rolesRes: make(map[int]*model.RoleRes),
}

func (this*roleResMgr) Load() {

	rr := make([]*model.RoleRes, 0)
	err := db.MasterDB.Find(&rr)
	if err != nil {
		log.DefaultLog.Error("roleResMgr load role_res table error")
	}

	
	this.mutex.Lock()
	for _, v := range rr {
		this.rolesRes[v.RId] = v
	}
	this.mutex.Unlock()
	go this.produce()

}


func (this*roleResMgr) Get(rid int) (*model.RoleRes, bool){
	
	this.mutex.RLock()
	r, ok := this.rolesRes[rid]
	this.mutex.RUnlock()
	
	if ok {
		return r, true
	}

	m := &model.RoleRes{}
	ok, err := db.MasterDB.Table(new(model.RoleRes)).Where("rid=?", rid).Get(m)
	if ok {
		
		this.mutex.Lock()
		this.rolesRes[rid] = m
		this.mutex.Unlock()
		
		return m, true
	}else{
		if err == nil{
			log.DefaultLog.Warn("RoleRes not found", zap.Int("rid", rid))
			return nil, false
		}else{
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}
	}
}

func (this*roleResMgr) Add(res *model.RoleRes) (){
	
	this.mutex.Lock()
	this.rolesRes[res.RId] = res
	this.mutex.Unlock()
}

func (this*roleResMgr) TryUseNeed(rid int, need*facility.NeedRes) bool{
	
	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()
	
	if ok {
		if need.Decree <= rr.Decree && need.Grain <= rr.Grain &&
			need.Stone <= rr.Stone && need.Wood <= rr.Wood &&
			need.Iron <= rr.Iron && need.Gold <= rr.Gold {
			rr.Decree -= need.Decree
			rr.Iron -= need.Iron
			rr.Wood -= need.Wood
			rr.Stone -= need.Stone
			rr.Grain -= need.Grain
			rr.Gold -= need.Gold

			rr.SyncExecute()
			return true
		}else{
			return false
		}
	}else {
		return false
	}
}

//政令是否足够
func (this*roleResMgr) DecreeIsEnough(rid int, cost int) bool {
	
	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()
	
	if ok {
		if rr.Decree >= cost {
			return true
		}else{
			return false
		}
	}else{
		return false
	}
}

func (this*roleResMgr) TryUseDecree(rid int, decree int) bool{
	
	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()
	
	if ok {
		if rr.Decree >= decree {
			rr.Decree -= decree
			rr.SyncExecute()
			return true
		}else{
			return false
		}
	}else{
		return false
	}
}



//金币是否足够
func (this*roleResMgr) GoldIsEnough(rid int, cost int) bool {
	
	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if ok {
		if rr.Gold >= cost {
			return true
		}else{
			return false
		}
	}else{
		return false
	}
}


func (this*roleResMgr) TryUseGold(rid int, gold int) bool{
	
	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()
	
	if ok {
		if rr.Gold >= gold {
			rr.Gold -= gold
			rr.SyncExecute()
			return true
		}else{
			return false
		}
	}else{
		return false
	}
}



func (this*roleResMgr) CutDown(rid int, b *model.MapRoleBuild) (*model.RoleRes, bool)  {
	rr, ok := this.Get(rid)
	if ok {
		rr.GrainYield = util.MaxInt(rr.GrainYield-b.Grain, 0)
		rr.IronYield = util.MaxInt(rr.IronYield-b.Iron, 0)
		rr.StoneYield = util.MaxInt(rr.StoneYield-b.Stone, 0)
		rr.WoodYield = util.MaxInt(rr.WoodYield-b.Wood, 0)
		rr.SyncExecute()
		return rr, true
	}
	return nil, false
}
func (this*roleResMgr) produce() {
	index := 1
	for true {
		t := static_conf.Basic.City.RecoveryTime
		time.Sleep(time.Duration(t) * time.Second)
		this.mutex.RLock()
		for _, v := range this.rolesRes {
			//加判断是因为爆仓了，资源不无故减少
			if v.Wood < v.DepotCapacity{
				v.Wood += util.MinInt(v.WoodYield/6, v.DepotCapacity)
			}

			if v.Iron < v.DepotCapacity{
				v.Iron += util.MinInt(v.IronYield/6, v.DepotCapacity)
			}

			if v.Stone < v.DepotCapacity{
				v.Stone += util.MinInt(v.StoneYield/6, v.DepotCapacity)
			}

			if v.Grain < v.DepotCapacity{
				v.Grain += util.MinInt(v.GrainYield/6, v.DepotCapacity)
			}

			if v.Gold < v.DepotCapacity{
				v.Grain += util.MinInt(v.GoldYield/6, v.DepotCapacity)
			}

			if index%6 == 0{
				if v.Decree < static_conf.Basic.Role.DecreeLimit{
					v.Decree+=1
				}
			}

			v.SyncExecute()
		}
		index++
		this.mutex.RUnlock()
		
	}
}

