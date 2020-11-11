package entity

import (
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

type BuildConfigMgr struct {
	mutex sync.RWMutex
	conf map[int]model.BuildConfig
}

var BCMgr = &BuildConfigMgr{
	conf: make(map[int]model.BuildConfig),
}

func (this* BuildConfigMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.conf)
	if err != nil {
		log.DefaultLog.Error("BuildConfigMgr load build_config table error")
	}
}

func (this* BuildConfigMgr) Maps() map[int]model.BuildConfig {
	return this.conf
}