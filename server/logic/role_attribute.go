package logic

import (
	"slgserver/server/model"
	"sync"
)

type roleAttributeMgr struct {
	mutex  sync.RWMutex
	attribute map[int]*model.RoleAttribute
}

var RAttributeMgr = &roleAttributeMgr{
	attribute: make(map[int]*model.RoleAttribute),
}

func (this* roleAttributeMgr) Load() {
	//加载
	this.mutex.Lock()
	l := UnionMgr.List()
	for _, c := range l {
		for _, rid := range c.MemberArray {
			a := &model.RoleAttribute{UnionId: c.Id}
			this.attribute[rid] = a
		}
	}
	this.mutex.Unlock()
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


func (this* roleAttributeMgr) IsHasUnion(rid int) bool{

	this.mutex.RLock()
	r, ok := this.attribute[rid]
	this.mutex.RUnlock()

	if ok {
		return r.UnionId != 0
	}else {
		return  false
	}
}

func (this* roleAttributeMgr) EnterUnion(rid, unionId int) {
	attr, ok := this.Get(rid)
	if ok {
		attr.UnionId = unionId
	}else{
		this.mutex.Lock()
		this.attribute[rid] = &model.RoleAttribute{UnionId: unionId}
		this.mutex.Unlock()
	}
}



