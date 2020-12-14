package logic

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
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
	t := make(map[int]*model.RoleAttribute)
	err := db.MasterDB.Find(t)
	if err != nil {
		log.DefaultLog.Error("roleAttributeMgr load role_attribute table error", zap.Error(err))
	}

	//获取联盟id
	this.mutex.Lock()
	for _, v:= range t {
		this.attribute[v.RId] = v
	}

	l := UnionMgr.List()
	for _, c := range l {
		for _, rid := range c.MemberArray {
			attr, ok := this.attribute[rid]
			if ok {
				attr.UnionId = c.Id
			}else{
				attr := this.create(rid)
				if attr != nil{
					attr.UnionId = c.Id
				}
			}
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

func (this* roleAttributeMgr) TryCreate(rid int) (*model.RoleAttribute, bool){
	attr, ok := this.Get(rid)
	if ok {
		return attr, true
	}else{
		this.mutex.Lock()
		defer this.mutex.Unlock()
		attr := this.create(rid)
		return attr, attr != nil
	}
}

func (this* roleAttributeMgr) create(rid int) *model.RoleAttribute{
	roleAttr := &model.RoleAttribute{RId: rid, ParentId: 0, UnionId: 0}
	if _ , err := db.MasterDB.Insert(roleAttr); err != nil {
		log.DefaultLog.Error("insert RoleAttribute error", zap.Error(err))
		return roleAttr
	}else{
		this.attribute[rid] = roleAttr
		return nil
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

func (this* roleAttributeMgr) UnionId(rid int) int{

	this.mutex.RLock()
	r, ok := this.attribute[rid]
	this.mutex.RUnlock()

	if ok {
		return r.UnionId
	}else {
		return  0
	}
}

func (this* roleAttributeMgr) EnterUnion(rid, unionId int) {
	attr, ok := this.TryCreate(rid)
	if ok {
		attr.UnionId = unionId
	}else{
		log.DefaultLog.Warn("EnterUnion not found roleAttribute", zap.Int("rid", rid))
	}
}



