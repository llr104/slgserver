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
	Id               int  		`xorm:"id pk autoincr"`
	RId              int  		`xorm:"rid"`
	CityId           int  		`xorm:"cityId"`
	Order            int8 		`xorm:"order"`
	FirstId          int       	`xorm:"firstId"`
	SecondId         int       	`xorm:"secondId"`
	ThirdId          int       	`xorm:"thirdId"`
	FirstSoldierCnt  int       	`xorm:"first_soldier_cnt"`
	SecondSoldierCnt int       	`xorm:"second_soldier_cnt"`
	ThirdSoldierCnt  int       	`xorm:"third_soldier_cnt"`
	State            int8  		`xorm:"state"` //状态，0:空闲 1:攻击 2：驻军 3:返回
	FromX            int       	`xorm:"from_x"`
	FromY            int       	`xorm:"from_y"`
	ToX              int       	`xorm:"to_x"`
	ToY              int       	`xorm:"to_y"`
	Start            time.Time 	`xorm:"start"`
	End              time.Time 	`xorm:"end"`
	NeedUpdate       bool      	`xorm:"-"`
}

func (this *Army) TableName() string {
	return "army"
}


