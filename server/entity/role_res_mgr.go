package entity

import (
	"errors"
	"fmt"
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


func (this* RoleResMgr) Get(rid int) (*model.RoleRes, error){
	this.mutex.RLock()
	r, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if ok {
		return r, nil
	}

	m := &model.RoleRes{}
	ok, err := db.MasterDB.Table(new(model.RoleRes)).Where("rid=?", rid).Get(m)
	if ok {
		this.mutex.Lock()
		this.rolesRes[rid] = m
		this.mutex.Unlock()
		return m, nil
	}else{
		if err == nil{
			return nil, errors.New(fmt.Sprintf("RoleRes %d not found", rid))
		}else{
			return nil, err
		}
	}
}

func (this* RoleResMgr) Add(res *model.RoleRes) (){
	this.mutex.Lock()
	defer this.mutex.Unlock()
	res.NeedUpdate = true
	this.rolesRes[res.RId] = res
}

func (this* RoleResMgr) TryUseNeed(rid int, need*facility.LevelNeedRes) bool{
	this.mutex.Lock()
	defer this.mutex.Unlock()
	rr, ok := this.rolesRes[rid]
	if ok {
		if need.Decree <= rr.Decree && need.Grain <= rr.Grain &&
			need.Stone <= rr.Stone && need.Wood <= rr.Wood && need.Iron <= rr.Iron {

			rr.NeedUpdate = true
			rr.Decree -= need.Decree
			rr.Iron -= need.Iron
			rr.Wood -= need.Wood
			rr.Stone -= need.Stone
			rr.Grain -= need.Grain
			return true
		}else{
			return false
		}
	}else {
		return false
	}

}


func (this* RoleResMgr) produce() {
	index := 1
	for true {
		time.Sleep(60*10*time.Second)
		//每个10分钟处理一次资源更新
		this.mutex.Lock()
		for _, v := range this.rolesRes {
			v.Wood += util.MinInt(v.WoodYield/6, v.DepotCapacity)
			v.Iron += util.MinInt(v.IronYield/6, v.DepotCapacity)
			v.Stone += util.MinInt(v.StoneYield/6, v.DepotCapacity)
			v.Grain += util.MinInt(v.GrainYield/6, v.DepotCapacity)
			v.Gold += util.MinInt(v.GoldYield/6, v.DepotCapacity)

			if index%6 == 0{
				v.Decree+=1
			}

			v.NeedUpdate = true
		}
		index++

		this.mutex.Unlock()
	}
}

func (this* RoleResMgr) toDatabase() {
	for true {
		time.Sleep(5*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.rolesRes {
			if v.NeedUpdate {
				_, err := db.MasterDB.Table(v).Cols("wood", "iron", "stone",
					"grain", "gold", "decree", "wood_yield",
					"iron_yield", "stone_yield", "gold_yield",
					"gold_yield", "depot_capacity").Update(v)
				if err != nil{
					log.DefaultLog.Error("RoleResMgr toDatabase error", zap.Error(err))
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