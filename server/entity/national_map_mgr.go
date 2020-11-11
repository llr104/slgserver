package entity

import (
	"go.uber.org/zap"
	"math/rand"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/util"
	"sync"
	"time"
)

type NationalMapMgr struct {
	mutex sync.RWMutex
	conf map[int]model.NationalMap
	confArr []model.NationalMap
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
			needInsert := 1600
			for i := 0; i < needInsert ; i++ {
				x := i/40
				y := i%40

				t := rand.Intn(4)+1
				m := &model.NationalMap{X: x, Y: y, Type: int8(t)}
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

	this.confArr = make([]model.NationalMap, len(this.conf))
	i := 0
	for _, v := range this.conf {
		this.confArr[i] = v
		i++
	}
}

func (this* NationalMapMgr) Scan(x, y int) []model.NationalMap {
	this.mutex.RLock()
	defer this.mutex.RUnlock()

	minX := util.MaxInt(0, x-3)
	maxX := util.MinInt(40, x+3)

	minY := util.MaxInt(0, y-3)
	maxY := util.MinInt(40, y+3)

	c := (maxX-minX+1)*(maxY-minY+1)
	r := make([]model.NationalMap, c)

	index := 0
	for i := minX; i < maxX; i++ {
		for j := minY; j < maxY; j++ {
			r[index] = this.confArr[i*40+j]
			index++
		}
	}
	return r
}