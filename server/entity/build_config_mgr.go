package entity

import (
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

type BuildConfigMgr struct {
	mutex sync.RWMutex
	conf map[int]model.MapBuildConfig
}

var BCMgr = &BuildConfigMgr{
	conf: make(map[int]model.MapBuildConfig),
}

func (this* BuildConfigMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.conf)
	if err != nil {
		log.DefaultLog.Error("BuildConfigMgr load build_config table error")
	}
}

func (this* BuildConfigMgr) Maps() map[int]model.MapBuildConfig {
	return this.conf
}

func (this* BuildConfigMgr) GetDurable(t int8, level int8) int {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	for _, v := range this.conf {
		if v.Type == t && v.Level == level{
			return v.Durable
		}
	}
	return 0
}