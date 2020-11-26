package model

import (
	"slgserver/server/static_conf/general"
	"time"
)

type General struct {
	DB 				dbSync		`xorm:"-"`
	Id        		int     	`xorm:"id pk autoincr"`
	RId       		int     	`xorm:"rid"`
	CfgId     		int     	`xorm:"cfgId"`
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

func (this* General) GetDestroy() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return (cfg.Destroy+cfg.DestroyGrow*int(this.Level))/100
	}
	return 0
}
