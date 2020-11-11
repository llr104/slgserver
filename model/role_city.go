package model

import "time"

type RoleCity struct {
	CityId		int			`json:"cityId" xorm:"cityId pk autoincr"`
	RId			int			`json:"rid" xorm:"rid"`
	Name		string		`json:"nickName" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	X			int			`json:"x"`
	Y			int			`json:"y"`
	IsMain		int8		`json:"is_main"`
	CreatedAt	time.Time	`json:"created_at"`
}

func (this *RoleCity) TableName() string {
	return "role_city"
}