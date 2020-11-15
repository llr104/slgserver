package model

import "time"

type MapRoleCity struct {
	CityId		int			`json:"cityId" xorm:"cityId pk autoincr"`
	RId			int			`json:"rid" xorm:"rid"`
	Name		string		`json:"nickName" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	X			int			`json:"x"`
	Y			int			`json:"y"`
	IsMain		int8		`json:"is_main"`
	Level		int8		`json:"level"`
	CurDurable	int			`json:"cur_durable"`
	MaxDurable	int			`json:"max_durable"`
	CreatedAt	time.Time	`json:"created_at"`
}

func (this *MapRoleCity) TableName() string {
	return "map_role_city"
}