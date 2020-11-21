package model

import "time"

type MapRoleCity struct {
	CityId		int			`xorm:"cityId pk autoincr"`
	RId			int			`xorm:"rid"`
	Name		string		`xorm:"name" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	X			int			`xorm:"x"`
	Y			int			`xorm:"y"`
	IsMain		int8		`xorm:"is_main"`
	Level		int8		`xorm:"level"`
	CurDurable	int			`xorm:"cur_durable"`
	MaxDurable	int			`xorm:"max_durable"`
	CreatedAt	time.Time	`xorm:"created_at"`
}

func (this *MapRoleCity) TableName() string {
	return "map_role_city"
}