package logic

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/model"
	"sync"
	"time"
)

type coalitionMgr struct {
	mutex  sync.RWMutex
	unions map[int]*model.Coalition
}

var UnionMgr = &coalitionMgr{
	unions: make(map[int]*model.Coalition),
}

func (this* coalitionMgr) Load() {

	rr := make([]*model.Coalition, 0)
	err := db.MasterDB.Find(&rr)
	if err != nil {
		log.DefaultLog.Error("coalitionMgr load union table error")
	}

	this.mutex.Lock()
	for _, v := range rr {
		this.unions[v.Id] = v
	}
	this.mutex.Unlock()
}


func (this* coalitionMgr) Get(id int) (*model.Coalition, bool){

	this.mutex.RLock()
	r, ok := this.unions[id]
	this.mutex.RUnlock()

	if ok {
		return r, true
	}

	m := &model.Coalition{}
	ok, err := db.MasterDB.Table(new(model.Coalition)).Where("id=?", id).Get(m)
	if ok {

		this.mutex.Lock()
		this.unions[id] = m
		this.mutex.Unlock()

		return m, true
	}else{
		if err == nil{
			log.DefaultLog.Warn("coalitionMgr not found", zap.Int("id", id))
			return nil, false
		}else{
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}
	}
}

func (this* coalitionMgr) Create(name string, rid int) (*model.Coalition, bool){
	m := &model.Coalition{Name: name, Ctime: time.Now(), CreateId: rid, Chairman: rid}
	_, err := db.MasterDB.Table(new(model.Coalition)).InsertOne(m)
	if err == nil {

		this.mutex.Lock()
		this.unions[m.Id] = m
		this.mutex.Unlock()

		return m, true
	}else{
		log.DefaultLog.Error("db error", zap.Error(err))
		return nil, false
	}
}

func (this* coalitionMgr) List() []*model.Coalition {
	r := make([]*model.Coalition, 0)
	this.mutex.Lock()
	for _, coalition := range this.unions {
		r = append(r, coalition)
	}
	this.mutex.Unlock()
	return r
}


