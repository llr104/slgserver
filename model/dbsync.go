package model

import "sync"

type dbSync struct {
	mutex 			sync.RWMutex
	needSyncToDB 	bool
}

func (this *dbSync) NeedSync() bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.needSyncToDB
}

func (this *dbSync) BeginSync() {
	this.mutex.Lock()
}

func (this *dbSync) EndSync()  {
	this.needSyncToDB = false
	this.mutex.Unlock()
}

func (this *dbSync) Sync()  {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.needSyncToDB = true
}

