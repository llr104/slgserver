package model

import (
	"encoding/json"
	"slgserver/server/conn"
	"slgserver/server/proto"
)

type CityFacility struct {
	DB         dbSync `xorm:"-"`
	Id         int    `xorm:"id pk autoincr"`
	RId        int    `xorm:"rid"`
	CityId     int    `xorm:"cityId"`
	Facilities string `xorm:"facilities"`
}

func (this *CityFacility) TableName() string {
	return "city_facility"
}


/* 推送同步 begin */
func (this*CityFacility) IsCellView() bool{
	return false
}

func (this*CityFacility) BelongToRId() []int{
	return []int{this.RId}
}

func (this*CityFacility) PushMsgName() string{
	return "facility.push"
}

func (this*CityFacility) ToProto() interface{}{
	p := make([]proto.Facility, 0)
	json.Unmarshal([]byte(this.Facilities), &p)
	return p
}

func (this*CityFacility) Push(){
	conn.ConnMgr.Push(this)
}

/* 推送同步 end */

func (this*CityFacility) Execute() {
	this.DB.Sync()
	this.Push()
}