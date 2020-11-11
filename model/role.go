package model

import "time"

type Role struct {
	RId			int			`json:"rid" xorm:"rid pk autoincr"`
	UId			int			`json:"uid" xorm:"uid"`
	SId			int			`json:"sid" xorm:"sid"`
	NickName	string		`json:"nickName" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	Balance		int			`json:"balance"`
	HeadId		int16		`json:"headId" xorm:"headId"`
	Sex			int8		`json:"sex"`
	Profile		string		`json:"profile"`
	LoginTime   time.Time	`json:"login_time"`
	LogoutTime  time.Time	`json:"logout_time"`
	CreatedAt	time.Time	`json:"created_at"`
}

func (this *Role) TableName() string {
	return "role"
}

