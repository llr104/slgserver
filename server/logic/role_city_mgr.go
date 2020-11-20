package logic

import (
	"errors"
	"fmt"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/util"
	"sync"
)

type RoleCityMgr struct {
	mutex  sync.RWMutex
	dbCity map[int]*model.MapRoleCity
	posCity map[int]*model.MapRoleCity
}

var RCMgr = &RoleCityMgr{
	dbCity: make(map[int]*model.MapRoleCity),
	posCity: make(map[int]*model.MapRoleCity),
}

func (this* RoleCityMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.dbCity)
	if err != nil {
		log.DefaultLog.Error("RoleCityMgr load role_city table error")
	}

	//转成posCity
	for _, v := range this.dbCity {
		posId := ToPosition(v.X, v.Y)
		this.posCity[posId] = v
	}
}

/*
该位置是否被角色建立城池
*/
func (this* RoleCityMgr) IsEmpty(x, y int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := ToPosition(x, y)
	_, ok := this.posCity[posId]
	return !ok
}

func (this* RoleCityMgr) PositionCity(x, y int) (*model.MapRoleCity, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := ToPosition(x, y)
	c,ok := this.posCity[posId]
	return c, ok
}

func (this* RoleCityMgr) Add(city *model.MapRoleCity) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.dbCity[city.CityId] = city
	this.posCity[ToPosition(city.X, city.Y)] = city
}

func (this* RoleCityMgr) Scan(x, y int) []*model.MapRoleCity {
	if x < 0 || x >= MapWith || y < 0 || y >= MapHeight {
		return nil
	}

	this.mutex.RLock()
	defer this.mutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(40, x+ScanWith)
	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(40, y+ScanHeight)

	cb := make([]*model.MapRoleCity, 0)
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			posId := ToPosition(i, j)
			v, ok := this.posCity[posId]
			if ok {
				cb = append(cb, v)
			}
		}
	}
	return cb
}

func (this* RoleCityMgr) ScanBlock(x, y, length int) []*model.MapRoleCity {
	if x < 0 || x >= MapWith || y < 0 || y >= MapHeight {
		return nil
	}

	this.mutex.RLock()
	defer this.mutex.RUnlock()

	maxX := util.MinInt(MapWith, x+length)
	maxY := util.MinInt(MapHeight, y+length)

	cb := make([]*model.MapRoleCity, 0)
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			posId := ToPosition(i, j)
			v, ok := this.posCity[posId]
			if ok {
				cb = append(cb, v)
			}
		}
	}
	return cb
}



func (this* RoleCityMgr) Get(cid int) (*model.MapRoleCity, error){
	this.mutex.RLock()
	r, ok := this.dbCity[cid]
	this.mutex.RUnlock()

	if ok {
		return r, nil
	}


	r = &model.MapRoleCity{}
	ok, err := db.MasterDB.Table(r).Where("cityId=?", cid).Get(r)
	if ok {
		this.mutex.Lock()
		this.dbCity[cid] = r
		this.mutex.Unlock()
		return r, nil
	}else{
		if err != nil{
			return nil, err
		}else{
			str := fmt.Sprintf("cid:%d CityFacility not found", cid)
			return nil, errors.New(str)
		}
	}
}