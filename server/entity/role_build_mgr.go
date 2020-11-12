package entity

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/util"
	"sync"
)


type RoleBuildMgr struct {
	mutex sync.RWMutex
	dbRB  map[int]*model.RoleBuild //key:dbId
	posRB map[int]*model.RoleBuild //key:posId
}


var RBMgr = &RoleBuildMgr{
	dbRB: make(map[int]*model.RoleBuild),
	posRB: make(map[int]*model.RoleBuild),
}

func (this* RoleBuildMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.dbRB)
	if err != nil {
		log.DefaultLog.Error("RoleBuildMgr load role_build table error", zap.Error(err))
	}

	//转成posRB
	for _, v := range this.dbRB {
		posId := v.X*MapWith+v.Y
		this.posRB[posId] = v
	}
}

/*
该位置是否被角色占领
*/
func (this* RoleBuildMgr) IsEmpty(x, y int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := MapWith*x+y
	_, ok := this.posRB[posId]
	return !ok
}

func (this* RoleBuildMgr) Scan(x, y int) []*model.RoleBuild{
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(40, x+ScanWith)
	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(40, y+ScanHeight)

	rb := make([]*model.RoleBuild, 0)
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			posId := i*ScanWith+j
			v, ok := this.posRB[posId]
			if ok {
				rb = append(rb, v)
			}
		}
	}

	return rb
}