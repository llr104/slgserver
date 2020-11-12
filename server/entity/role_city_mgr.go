package entity

import (
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

type RoleCityMgr struct {
	mutex  sync.RWMutex
	dbCity map[int]*model.RoleCity
	posCity map[int]*model.RoleCity
}

var RCMgr = &RoleCityMgr{
	dbCity: make(map[int]*model.RoleCity),
	posCity: make(map[int]*model.RoleCity),
}

func (this* RoleCityMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.dbCity)
	if err != nil {
		log.DefaultLog.Error("RoleCityMgr load role_city table error")
	}

	//转成posCity
	for _, v := range this.dbCity {
		posId := v.X*MapWith+v.Y
		this.posCity[posId] = v
	}
}

/*
该位置是否被角色建立城池
*/
func (this* RoleCityMgr) IsEmpty(x, y int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := MapWith*x+y
	_, ok := this.posCity[posId]
	return !ok
}