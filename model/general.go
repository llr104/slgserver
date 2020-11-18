package model

import "time"

type General struct {
	Id        int       `json:"id" xorm:"id pk autoincr"`
	RId       int       `json:"rid" xorm:"rid"`
	Name      string    `json:"name"`
	CfgId     int		`json:"cfgId" xorm:"cfgId"`
	Force     int       `json:"force"`
	Strategy  int       `json:"strategy"`
	Defense   int       `json:"defense"`
	Speed     int       `json:"speed"`
	Cost      int       `json:"cost"`
	ArmyId    int       `json:"armyId"`
	CityId    int       `json:"cityId"`
	CreatedAt time.Time `json:"created_at"`
}

func (this *General) TableName() string {
	return "general"
}

