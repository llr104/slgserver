package entity

import (
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

type RoleCityMgr struct {
	mutex sync.RWMutex
	conf map[int]model.RoleCity
}

var RCMgr = &RoleCityMgr{
	conf: make(map[int]model.RoleCity),
}

func (this* RoleCityMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.conf)
	if err != nil {
		log.DefaultLog.Error("RoleCityMgr load role_city table error")
	}

}