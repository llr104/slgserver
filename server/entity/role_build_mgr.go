package entity

import (
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

type RoleBuildMgr struct {
	mutex sync.RWMutex
	conf map[int]model.RoleBuild
}

var RBMgr = &RoleBuildMgr{
	conf: make(map[int]model.RoleBuild),
}

func (this* RoleBuildMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.conf)
	if err != nil {
		log.DefaultLog.Error("RoleBuildMgr load role_build table error")
	}

}