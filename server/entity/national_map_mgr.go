package entity

import (
	"go.uber.org/zap"
	"math/rand"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"sync"
	"time"
)

type NationalMapMgr struct {
	mutex sync.RWMutex
	conf map[int]model.NationalMap
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

}