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
	dbRB  map[int]*model.MapRoleBuild //key:dbId
	posRB map[int]*model.MapRoleBuild //key:posId
}


var RBMgr = &RoleBuildMgr{
	dbRB: make(map[int]*model.MapRoleBuild),
	posRB: make(map[int]*model.MapRoleBuild),
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
		posId := ToPosition(v.X, v.Y)
		this.posRB[posId] = v
	}
}

/*
该位置是否被角色占领
*/
func (this* RoleBuildMgr) IsEmpty(x, y int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := ToPosition(x, y)
	_, ok := this.posRB[posId]
	return !ok
}

func (this* RoleBuildMgr) Scan(x, y int) []*model.MapRoleBuild {
	if x < 0 || x >= MapWith || y < 0 || y >= MapHeight {
		return nil
	}


	this.mutex.RLock()
	defer this.mutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(MapWith, x+ScanWith)
	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(MapHeight, y+ScanHeight)

	rb := make([]*model.MapRoleBuild, 0)
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			posId := ToPosition(i, j)
			v, ok := this.posRB[posId]
			if ok {
				rb = append(rb, v)
			}
		}
	}

	return rb
}

func (this* RoleBuildMgr) ScanBlock(x, y, length int) []*model.MapRoleBuild {
	if x < 0 || x >= MapWith || y < 0 || y >= MapHeight {
		return nil
	}


	this.mutex.RLock()
	defer this.mutex.RUnlock()

	maxX := util.MinInt(MapWith, x+length)
	maxY := util.MinInt(MapHeight, y+length)

	rb := make([]*model.MapRoleBuild, 0)
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			posId := ToPosition(i, j)
			v, ok := this.posRB[posId]
			if ok {
				rb = append(rb, v)
			}
		}
	}

	return rb
}