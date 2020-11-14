package entity

import (
	"errors"
	"fmt"
	"slgserver/db"
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

func (this* RoleMgr) Get(rid int) (*model.Role, error){
	this.mutex.Lock()
	r, ok := this.roles[rid]
	this.mutex.Unlock()

	if ok {
		return r, nil
	}

	m := &model.Role{}
	ok, err := db.MasterDB.Table(new(model.Role)).Where("rid=?", rid).Get(m)
	if ok {
		this.mutex.Lock()
		this.roles[rid] = m
		this.mutex.Unlock()
		return m, nil
	}else{
		if err == nil{
			return nil, errors.New(fmt.Sprintf("role %d not found", rid))
		}else{
			return nil, err
		}
	}
}

