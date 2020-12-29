package model

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/static_conf/facility"
	"time"
)

/*******db 操作begin********/
var dbCFMgr *cfDBMgr
func init() {
	dbCFMgr = &cfDBMgr{cs: make(chan *CityFacility, 100)}
	go dbCFMgr.running()
}

type cfDBMgr struct {
	cs   chan *CityFacility
}

func (this*cfDBMgr) running()  {
	for true {
		select {
		case c := <- this.cs:
			if c.Id >0 {
				_, err := db.MasterDB.Table(c).ID(c.Id).Cols("facilities").Update(c)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update CityFacility fail, because id <= 0")
			}
		}
	}
}

func (this*cfDBMgr) push(c *CityFacility)  {
	this.cs <- c
}
/*******db 操作end********/


type Facility struct {
	Name   string 	`json:"name"`
	Level  int8   	`json:"level"`
	Type   int8   	`json:"type"`
	UpTime int64	`json:"up_time"`	//升级的时间戳，0表示该等级已经升级完成了
}

func (this* Facility) GetLevel() int8  {
	//升级这里做成被动触发产生，不做定时
	if this.UpTime > 0{
		cur := time.Now().Unix()
		cost := facility.FConf.CostTime(this.Type, this.Level+1)
		if cur >= this.UpTime + int64(cost){
			this.Level+=1
			this.UpTime = 0
		}
	}
	return this.Level
}

type CityFacility struct {
	Id         int    `xorm:"id pk autoincr"`
	RId        int    `xorm:"rid"`
	CityId     int    `xorm:"cityId"`
	Facilities string `xorm:"facilities"`
}

func (this *CityFacility) TableName() string {
	return "tb_city_facility" + fmt.Sprintf("_%d", ServerId)
}


func (this *CityFacility) SyncExecute() {
	dbCFMgr.push(this)
}

func (this*CityFacility) Facility()[]Facility {
	facilities := make([]Facility, 0)
	json.Unmarshal([]byte(this.Facilities), &facilities)
	return facilities
}