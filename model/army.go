package model

import "time"

const (
	ArmyIdle  	= 0
	ArmyAttack  = 1
	ArmyDefend  = 2
	ArmyBack  	= 3
)

//军队
type Army struct {
	Id               int  		`json:"id" xorm:"id pk autoincr"`
	RId              int  		`json:"rid" xorm:"rid"`
	CityId           int  		`json:"cityId" xorm:"cityId"`
	Order            int8 		`json:"order"`
	FirstId          int       	`json:"firstId" xorm:"firstId"`
	SecondId         int       	`json:"secondId" xorm:"secondId"`
	ThirdId          int       	`json:"thirdId" xorm:"thirdId"`
	FirstSoldierCnt  int       	`json:"first_soldier_cnt" xorm:"first_soldier_cnt"`
	SecondSoldierCnt int       	`json:"second_soldier_cnt" xorm:"second_soldier_cnt"`
	ThirdSoldierCnt  int       	`json:"third_soldier_cnt" xorm:"third_soldier_cnt"`
	State            int8  		`json:"state"` //状态，0:空闲 1:攻击 2：驻军 3:返回
	FromX            int       	`json:"from_x"`
	FromY            int       	`json:"from_y"`
	ToX              int       	`json:"to_x"`
	ToY              int       	`json:"to_y"`
	Start            time.Time 	`json:"start"`
	End              time.Time 	`json:"end"`
	NeedUpdate       bool      	`json:"-" xorm:"-"`
}

func (this *Army) TableName() string {
	return "army"
}


