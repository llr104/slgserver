package model

import "time"

type General struct {
	DB 				dbSync		`xorm:"-"`
	Id        		int     	`xorm:"id pk autoincr"`
	RId       		int     	`xorm:"rid"`
	Name      		string  	`xorm:"name"`
	CfgId     		int     	`xorm:"cfgId"`
	Force         	int       	`xorm:"force"`
	Strategy      	int       	`xorm:"strategy"`
	Defense       	int       	`xorm:"defense"`
	Speed         	int       	`xorm:"speed"`
	Destroy       	int       	`xorm:"destroy"`
	ForceGrow     	int       	`xorm:"force_grow"`
	StrategyGrow  	int       	`xorm:"strategy_grow"`
	DefenseGrow   	int       	`xorm:"defense_grow"`
	SpeedGrow     	int       	`xorm:"speed_grow"`
	DestroyGrow   	int       	`xorm:"destroy_grow"`
	PhysicalPower 	int       	`xorm:"physical_power"`
	Level         	int8      	`xorm:"level"`
	Cost          	int       	`xorm:"cost"`
	Exp           	int       	`xorm:"exp"`
	Order         	int8      	`xorm:"order"`
	CityId        	int       	`xorm:"cityId"`
	CreatedAt     	time.Time 	`xorm:"created_at"`
}

func (this *General) TableName() string {
	return "general"
}

