package model

import "time"

type General struct {
	Id        	int     	`xorm:"id pk autoincr"`
	RId       	int     	`xorm:"rid"`
	Name      	string  	`xorm:"name"`
	CfgId     	int     	`xorm:"cfgId"`
	Force     	int     	`xorm:"force"`
	Strategy  	int     	`xorm:"strategy"`
	Defense   	int     	`xorm:"defense"`
	Speed     	int     	`xorm:"speed"`
	Destroy   	int     	`xorm:"destroy"`
	Level		int8     	`xorm:"level"`
	Cost      	int     	`xorm:"cost"`
	Exp      	int     	`xorm:"exp"`
	Order     	int8     	`xorm:"order"`
	CityId    	int     	`xorm:"cityId"`
	CreatedAt 	time.Time	`xorm:"created_at"`
	NeedUpdate	bool		`xorm:"-"`
}

func (this *General) TableName() string {
	return "general"
}

