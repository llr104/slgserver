package model

import "sync"

type dbSync struct {
	mutex 			sync.RWMutex
	disable			bool
	needSyncToDB 	bool
}

func (this*dbSync) Disable(b bool) {
	this.disable = b
}

func (this *dbSync) NeedSync() bool {
	if this.disable {
		return false
	}else{
		this.mutex.RLock()
		defer this.mutex.RUnlock()
		return this.needSyncToDB
	}
}

func (this *dbSync) BeginSync() {
	if this.disable {
		return
	}
	this.mutex.Lock()
}

func (this *dbSync) EndSync()  {
	if this.disable {
		return
	}
	this.needSyncToDB = false
	this.mutex.Unlock()
}

func (this *dbSync) Sync()  {
	if this.disable {
		return
	}

	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.needSyncToDB = true
}

type PushAndDB interface {
	SyncExecute()
}

