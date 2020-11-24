package model

import (
	"encoding/json"
	"fmt"
	"time"
	"xorm.io/xorm"
)

const (
	ArmyCmdIdle   = 0
	ArmyCmdAttack = 1
	ArmyCmdDefend = 2
	ArmyCmdBack   = 3
)

const (
	ArmyRunning  	= 0
	ArmyStop  		= 1
)

//军队
type Army struct {
	DB           		dbSync 		`xorm:"-"`
	Id           		int    		`xorm:"id pk autoincr"`
	RId          		int    		`xorm:"rid"`
	CityId       		int    		`xorm:"cityId"`
	Order        		int8   		`xorm:"order"`
	Generals     		string 		`xorm:"generals"`
	Soldiers     		string 		`xorm:"soldiers"`
	GeneralArray 		[]int  		`xorm:"-"`
	SoldierArray 		[]int  		`xorm:"-"`
	Cmd         		int8   		`xorm:"cmd"` //执行命令0:空闲 1:攻击 2：驻军 3:返回
	State				int8		`xorm:"-"` //状态:0:running,1:stop
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
	if name == "generals"{
		this.GeneralArray = []int{0,0,0}
		if cell != nil{
			gs, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(gs, &this.GeneralArray)
				fmt.Println(this.GeneralArray)
			}
		}
	}else if name == "soldiers"{
		this.SoldierArray = []int{0,0,0}
		if cell != nil{
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.SoldierArray)
				fmt.Println(this.SoldierArray)
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
