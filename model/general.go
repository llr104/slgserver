package model

import "time"

type General struct {
	Id        	int     	`json:"id" xorm:"id pk autoincr"`
	RId       	int     	`json:"rid" xorm:"rid"`
	Name      	string  	`json:"name"`
	CfgId     	int     	`json:"cfgId" xorm:"cfgId"`
	Force     	int     	`json:"force"`
	Strategy  	int     	`json:"strategy"`
	Defense   	int     	`json:"defense"`
	Speed     	int     	`json:"speed"`
	Destroy   	int     	`json:"destroy"`
	Level		int8     	`json:"level"`
	Cost      	int     	`json:"cost"`
	Exp      	int     	`json:"exp"`
	Order     	int8     	`json:"order" xorm:"order"`
	CityId    	int     	`json:"cityId" xorm:"cityId"`
	CreatedAt 	time.Time	`json:"created_at"`
}

func (this *General) TableName() string {
	return "general"
}

