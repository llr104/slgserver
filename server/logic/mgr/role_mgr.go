package mgr

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/model"
	"sync"
)

func RoleNickName(rid int) string {
	vRole, ok := RMgr.Get(rid)
	if ok {
		return vRole.NickName
	}
	return ""
}

type roleMgr struct {
	mutex  sync.RWMutex
	roles map[int]*model.Role
}

var RMgr = &roleMgr{
	roles: make(map[int]*model.Role),
}

func (this*roleMgr) Get(rid int) (*model.Role, bool){
	this.mutex.RLock()
	r, ok := this.roles[rid]
	this.mutex.RUnlock()

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

