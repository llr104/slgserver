package model

import (
	"encoding/json"
	"time"
	"xorm.io/xorm"
)

const (
	ArmyIdle  	= 0
	ArmyAttack  = 1
	ArmyDefend  = 2
	ArmyBack  	= 3
)

//军队
type Army struct {
	DB 					dbSync		`xorm:"-"`
	Id               	int  		`xorm:"id pk autoincr"`
	RId              	int  		`xorm:"rid"`
	CityId           	int  		`xorm:"cityId"`
	Order            	int8 		`xorm:"order"`
	Generals			string		`xorm:"generals"`
	Soldiers			string		`xorm:"soldiers"`
	GeneralArray		[]int		`xorm:"-"`
	SoldierArray		[]int		`xorm:"-"`
	State            	int8  		`xorm:"state"` //状态，0:空闲 1:攻击 2：驻军 3:返回
	FromX            	int       	`xorm:"from_x"`
	FromY            	int       	`xorm:"from_y"`
	ToX              	int       	`xorm:"to_x"`
	ToY              	int       	`xorm:"to_y"`
	Start            	time.Time 	`xorm:"start"`
	End              	time.Time 	`xorm:"end"`
}

func (this *Army) TableName() string {
	return "army"
}

func (this *Army) AfterSet(name string, cell xorm.Cell){
	this.SoldierArray = []int{0,0,0}
	this.GeneralArray = []int{0,0,0}
	if name == "generals"{
		if cell != nil{
			gs, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(gs, &this.GeneralArray)
			}
		}
	}else if name == "soldiers"{
		if cell != nil{
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.SoldierArray)
			}
		}
	}
}

func (this* Army) BeforeInsert() {

	data, _ := json.Marshal(this.GeneralArray)
	this.Generals = string(data)

	data, _ = json.Marshal(this.SoldierArray)
	this.Soldiers = string(data)
}

func (this* Army) BeforeUpdate() {

	data, _ := json.Marshal(this.GeneralArray)
	this.Generals = string(data)

	data, _ = json.Marshal(this.SoldierArray)
	this.Soldiers = string(data)
}
