package logic

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
)

type RoleMgr struct {
	mutex  sync.RWMutex
	roles map[int]*model.Role
}

var RMgr = &RoleMgr{
	roles: make(map[int]*model.Role),
}

func (this* RoleMgr) Get(rid int) (*model.Role, bool){
	this.mutex.Lock()
	r, ok := this.roles[rid]
	this.mutex.Unlock()

	if ok {
		return r, true
	}

	m := &model.Role{}
	ok, err := db.MasterDB.Table(new(model.Role)).Where("rid=?", rid).Get(m)
	log.DefaultLog.Warn("db error", zap.Error(err))
	if ok {
		this.mutex.Lock()
		this.roles[rid] = m
		this.mutex.Unlock()
		return m, true
	}else{
		return nil, false
	}
}

