package logic

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/server/static_conf/facility"
	"slgserver/util"
	"sync"
	"time"
)

type RoleResMgr struct {
	mutex  sync.RWMutex
	rolesRes map[int]*model.RoleRes
}

var RResMgr = &RoleResMgr{
	rolesRes: make(map[int]*model.RoleRes),
}

func (this* RoleResMgr) Load() {

	rr := make([]model.RoleRes, 0)
	err := db.MasterDB.Find(&rr)
	if err != nil {
		log.DefaultLog.Error("RoleResMgr load role_res table error")
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	for _, v := range rr {
		this.rolesRes[v.RId] = &v
	}

	go this.produce()
	go this.toDatabase()

}


func (this* RoleResMgr) Get(rid int) (*model.RoleRes, bool){
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

func (this* RoleResMgr) Add(res *model.RoleRes) (){
	this.mutex.Lock()
	defer this.mutex.Unlock()
	res.DB.Sync()
	this.rolesRes[res.RId] = res
}

func (this* RoleResMgr) TryUseNeed(rid int, need*facility.NeedRes) bool{
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rr, ok := this.rolesRes[rid]
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

			rr.DB.Sync()
			return true
		}else{
			return false
		}
	}else {
		return false
	}
}

func (this* RoleResMgr) TryUseDecree(rid int, decree int) bool{
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rr, ok := this.rolesRes[rid]
	if ok {
		if rr.Decree >= decree {
			rr.Decree -= decree
			rr.DB.Sync()
			return true
		}else{
			return false
		}
	}else{
		return false
	}
}

func (this* RoleResMgr) CutDown(rid int, b *model.MapRoleBuild) (*model.RoleRes, bool)  {
	rr, ok := this.Get(rid)
	if ok {
		rr.GrainYield = util.MaxInt(rr.GrainYield-b.Grain, 0)
		rr.IronYield = util.MaxInt(rr.IronYield-b.Iron, 0)
		rr.StoneYield = util.MaxInt(rr.StoneYield-b.Stone, 0)
		rr.WoodYield = util.MaxInt(rr.WoodYield-b.Wood, 0)
		rr.DB.Sync()
		return rr, true
	}
	return nil, false
}
func (this* RoleResMgr) produce() {
	index := 1
	for true {
		//每个10分钟处理一次资源更新
		time.Sleep(60*10*time.Second)
		this.mutex.Lock()
		for _, v := range this.rolesRes {
			//加判断是因为爆仓了，资源不无故减少
			if v.WoodYield < v.DepotCapacity{
				v.Wood += util.MinInt(v.WoodYield/6, v.DepotCapacity)
			}

			if v.IronYield < v.DepotCapacity{
				v.Iron += util.MinInt(v.IronYield/6, v.DepotCapacity)
			}

			if v.StoneYield < v.DepotCapacity{
				v.Stone += util.MinInt(v.StoneYield/6, v.DepotCapacity)
			}

			if v.GrainYield < v.DepotCapacity{
				v.Grain += util.MinInt(v.GrainYield/6, v.DepotCapacity)
			}

			if v.GoldYield < v.DepotCapacity{
				v.Grain += util.MinInt(v.GoldYield/6, v.DepotCapacity)
			}

			if index%6 == 0{
				v.Decree+=1
			}

			v.DB.Sync()
		}
		index++

		this.mutex.Unlock()
	}
}

func (this* RoleResMgr) toDatabase() {
	for true {
		time.Sleep(2*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.rolesRes {
			if v.DB.NeedSync() {
				v.DB.BeginSync()
				_, err := db.MasterDB.Table(v).ID(v.Id).Cols("wood", "iron", "stone",
					"grain", "gold", "decree", "wood_yield",
					"iron_yield", "stone_yield", "gold_yield",
					"gold_yield", "depot_capacity").Update(v)
				if err != nil{
					log.DefaultLog.Error("RoleResMgr toDatabase error", zap.Error(err))
				}
				v.DB.EndSync()
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