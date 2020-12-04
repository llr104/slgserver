package model

import (
	"encoding/json"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
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

func (this* cfDBMgr) running()  {
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

func (this* cfDBMgr) push(c *CityFacility)  {
	this.cs <- c
}
/*******db 操作end********/


type Facility struct {
	Name   string `json:"name"`
	Level  int8   `json:"level"`
	Type   int8   `json:"type"`
}


type CityFacility struct {
	Id         int    `xorm:"id pk autoincr"`
	RId        int    `xorm:"rid"`
	CityId     int    `xorm:"cityId"`
	Facilities string `xorm:"facilities"`
}

func (this *CityFacility) TableName() string {
	return "city_facility"
}


func (this *CityFacility) SyncExecute() {
	dbCFMgr.push(this)
}

func (this* CityFacility) Facility()[]Facility {
	facilities := make([]Facility, 0)
	json.Unmarshal([]byte(this.Facilities), &facilities)
	return facilities
}