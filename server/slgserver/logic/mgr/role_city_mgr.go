package mgr

import (
	"sync"
	"time"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/util"
	"go.uber.org/zap"
)


type roleCityMgr struct {
	mutex  sync.RWMutex
	dbCity map[int]*model.MapRoleCity     //key: cid
	posCity map[int]*model.MapRoleCity    //key: pos
	roleCity map[int][]*model.MapRoleCity //key: rid
}

var RCMgr = &roleCityMgr{
	dbCity: make(map[int]*model.MapRoleCity),
	posCity: make(map[int]*model.MapRoleCity),
	roleCity: make(map[int][]*model.MapRoleCity),
}

func GetCityCost(cid int) int8 {
	return RFMgr.GetCost(cid) + static_conf.Basic.City.Cost
}

func GetMaxDurable(cid int) int {
	return RFMgr.GetMaxDurable(cid) + static_conf.Basic.City.Durable
}

func GetCityLV(cid int) int8  {
	return RFMgr.GetCityLV(cid)
}

func (this*roleCityMgr) Load() {

	err := db.MasterDB.Find(this.dbCity)
	if err != nil {
		log.DefaultLog.Error("roleCityMgr load role_city table error")
	}

	//转成posCity、roleCity
	for _, v := range this.dbCity {
		posId := global.ToPosition(v.X, v.Y)
		this.posCity[posId] = v
		_, ok := this.roleCity[v.RId]
		if ok == false{
			this.roleCity[v.RId] = make([]*model.MapRoleCity, 0)
		}
		this.roleCity[v.RId] = append(this.roleCity[v.RId], v)
	}

	go this.running()
}

func (this*roleCityMgr) running() {
	for true {
		t := static_conf.Basic.City.RecoveryTime
		time.Sleep(time.Duration(t) * time.Second)
		this.mutex.RLock()
		for _, city := range this.dbCity {
			if city.CurDurable < GetMaxDurable(city.CityId){
				city.DurableChange(100)
				city.SyncExecute()
			}
		}
		this.mutex.RUnlock()
	}
}

/*
该位置是否被角色建立城池
*/
func (this*roleCityMgr) IsEmpty(x, y int) bool {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := global.ToPosition(x, y)
	_, ok := this.posCity[posId]
	return !ok
}

func (this*roleCityMgr) PositionCity(x, y int) (*model.MapRoleCity, bool) {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	posId := global.ToPosition(x, y)
	c,ok := this.posCity[posId]
	return c, ok
}

func (this*roleCityMgr) Add(city *model.MapRoleCity) {

	this.mutex.Lock()
	defer this.mutex.Unlock()
	this.dbCity[city.CityId] = city
	this.posCity[global.ToPosition(city.X, city.Y)] = city

	_, ok := this.roleCity[city.RId]
	if ok == false{
		this.roleCity[city.RId] = make([]*model.MapRoleCity, 0)
	}
	this.roleCity[city.RId] = append(this.roleCity[city.RId], city)
}

func (this*roleCityMgr) Scan(x, y int) []*model.MapRoleCity {
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
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
			posId := global.ToPosition(i, j)
			v, ok := this.posCity[posId]
			if ok {
				cb = append(cb, v)
			}
		}
	}
	return cb
}

func (this*roleCityMgr) ScanBlock(x, y, length int) []*model.MapRoleCity {
	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}

	this.mutex.RLock()
	defer this.mutex.RUnlock()

	maxX := util.MinInt(global.MapWith, x+length-1)
	maxY := util.MinInt(global.MapHeight, y+length-1)

	cb := make([]*model.MapRoleCity, 0)
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			posId := global.ToPosition(i, j)
			v, ok := this.posCity[posId]
			if ok {
				cb = append(cb, v)
			}
		}
	}
	return cb
}

func (this*roleCityMgr) GetByRId(rid int) ([]*model.MapRoleCity, bool){
	this.mutex.RLock()
	r, ok := this.roleCity[rid]
	this.mutex.RUnlock()
	return r, ok
}

func (this*roleCityMgr) GetMainCity(rid int) (*model.MapRoleCity, bool){
	citys, ok := this.GetByRId(rid)
	if ok == false {
		return nil, false
	}else{
		for _, city := range citys {
			if city.IsMain == 1{
				return city, true
			}
		}
	}
	return nil, false
}

func (this*roleCityMgr) Get(cid int) (*model.MapRoleCity, bool){
	this.mutex.RLock()
	r, ok := this.dbCity[cid]
	this.mutex.RUnlock()

	if ok {
		return r, true
	}


	r = &model.MapRoleCity{}
	ok, err := db.MasterDB.Table(r).Where("cityId=?", cid).Get(r)
	if err != nil{
		log.DefaultLog.Warn("db error", zap.Error(err))
	}

	if ok {
		this.mutex.Lock()
		this.dbCity[cid] = r
		this.mutex.Unlock()
		return r, true
	}else{
		return nil, false
	}
}
