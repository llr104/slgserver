package entity

import (
	"errors"
	"fmt"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
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

	go this.running()

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
	this.rolesRes[res.RId] = res
}

func (this* RoleResMgr) running() {
	for true {
		time.Sleep(60*10*time.Second)

		//每个10分钟处理一次资源更新
		index := 1
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
		}
		index++

		this.mutex.Unlock()
	}
}
