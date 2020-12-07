package logic

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/model"
	"sync"
)

type roleMgr struct {
	mutex  sync.RWMutex
	roles map[int]*model.Role
}

var RMgr = &roleMgr{
	roles: make(map[int]*model.Role),
}

func (this*roleMgr) Get(rid int) (*model.Role, bool){
	this.mutex.Lock()
	r, ok := this.roles[rid]
	this.mutex.Unlock()

	if ok {
		return r, true
	}

	m := &model.Role{}
	ok, err := db.MasterDB.Table(new(model.Role)).Where("rid=?", rid).Get(m)
	if ok {
		this.mutex.Lock()
		this.roles[rid] = m
		this.mutex.Unlock()
		return m, true
	}else{
		log.DefaultLog.Warn("db error", zap.Error(err), zap.Int("rid", rid))
		return nil, false
	}
}

