package model

import (
	"encoding/json"
	"time"
	"xorm.io/xorm"
)

type Coalition struct {
	Id           int       `xorm:"id pk autoincr"`
	Name         string    `xorm:"name"`
	Members      string    `xorm:"members"`
	MemberArray  []int     `xorm:"-"`
	CreateId     int       `xorm:"create_id"`
	Chairman     int       `xorm:"chairman"`
	ViceChairman int       `xorm:"vice_chairman"`
	Notice       string    `xorm:"notice"`
	Ctime        time.Time `xorm:"ctime"`
}


func (this *Coalition) TableName() string {
	return "coalition"
}

func (this *Coalition) AfterSet(name string, cell xorm.Cell){
	if name == "members"{
		if cell != nil{
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.MemberArray)
			}
		}
	}
}

func (this *Coalition) BeforeInsert() {
	data, _ := json.Marshal(this.MemberArray)
	this.Members = string(data)
}

func (this *Coalition) BeforeUpdate() {
	data, _ := json.Marshal(this.MemberArray)
	this.Members = string(data)
}

func (this* Coalition) Cnt() int{
	return len(this.MemberArray)
}

type CoalitionApply struct {
	Id          int       `xorm:"id pk autoincr"`
	CoalitionId int       `xorm:"coalition_id"`
	RId         int       `xorm:"rid"`
	State       int8      `xorm:"state"`
	Ctime       time.Time `xorm:"ctime"`
}

func (this *CoalitionApply) TableName() string {
	return "coalition_apply"
}
