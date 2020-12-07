package logic

import (
	"slgserver/server/model"
	"sync"
)

type roleAttributeMgr struct {
	mutex  sync.RWMutex
	attribute map[int]*model.RoleAttribute
}

var RoleAttributeMgr = &roleAttributeMgr{
	attribute: make(map[int]*model.RoleAttribute),
}

func (this* roleAttributeMgr) Load() {

}


func (this* roleAttributeMgr) Get(rid int) (*model.RoleAttribute, bool){

	this.mutex.RLock()
	r, ok := this.attribute[rid]
	this.mutex.RUnlock()

	if ok {
		return r, true
	}else {
		return nil, false
	}
}

