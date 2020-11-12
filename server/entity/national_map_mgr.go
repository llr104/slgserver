package entity

import (
	"go.uber.org/zap"
	"math/rand"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/util"
	"sort"
	"sync"
	"time"
)

const MapWith = 40
const MapHeight = 40
const ScanWith = 3
const ScanHeight = 3

type NMArray struct {
	arr []model.NationalMap
}

func (this NMArray) Len() int {
	return len(this.arr)
}

func (this NMArray) Swap(i, j int) {
	this.arr[i], this.arr[j] = this.arr[j], this.arr[i]
}

func (this NMArray) Less(i, j int) bool {
	if this.arr[i].X < this.arr[j].X{
		return true
	}else if this.arr[i].X == this.arr[j].X {
		return this.arr[i].Y < this.arr[j].Y
	}else{
		return false
	}
}

type NationalMapMgr struct {
	mutex sync.RWMutex
	conf map[int]model.NationalMap
	confArr NMArray
}

var NMMgr = &NationalMapMgr{
	conf: make(map[int]model.NationalMap),
}

func (this* NationalMapMgr) Load() {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	err := db.MasterDB.Find(this.conf)
	if err != nil {
		log.DefaultLog.Error("NationalMapMgr load national_map table error")
	}else{
		if len(this.conf) == 0 {
			session := db.MasterDB.NewSession()
			//随机初始化数据
			rand.Seed(time.Now().UnixNano())
			needInsert := MapWith*MapHeight
			for i := 0; i < needInsert ; i++ {
				x := i/MapWith
				y := i%MapWith

				t := rand.Intn(4)+1
				m := &model.NationalMap{X: x, Y: y, Type: int8(t), Level: 1}
				_, err := db.MasterDB.Table(m).InsertOne(m)
				if err != nil{
					session.Rollback()
					return
				}
			}
			session.Commit()

			db.MasterDB.Find(this.conf)
			log.DefaultLog.Info("NationalMapMgr load", zap.Int("len", len(this.conf)))
		}
	}

	this.confArr.arr = make([]model.NationalMap, len(this.conf))
	i := 0
	for _, v := range this.conf {
		this.confArr.arr[i] = v
		i++
	}

	sort.Sort(this.confArr)

}

func (this* NationalMapMgr) Scan(x, y int) []model.NationalMap {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	minX := util.MaxInt(0, x-ScanWith)
	maxX := util.MinInt(40, x+ScanWith)

	minY := util.MaxInt(0, y-ScanHeight)
	maxY := util.MinInt(40, y+ScanHeight)

	c := (maxX-minX+1)*(maxY-minY+1)
	r := make([]model.NationalMap, c)

	index := 0
	for i := minX; i <= maxX; i++ {
		for j := minY; j <= maxY; j++ {
			r[index] = this.confArr.arr[i*ScanWith+j]
			index++
		}
	}
	return r
}