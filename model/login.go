package model

import (
	"time"
)

const (
	Login = iota
	Logout
)

type LoginHistory struct {
	Id       int       `xorm:"id pk autoincr"`
	UId      int       `xorm:"uid"`
	Time     time.Time `xorm:"time"`
	Ip       string    `xorm:"ip"`
	State    int8      `xorm:"state"`
	Hardware string    `xorm:"hardware"`
}

func (this *LoginHistory) TableName() string {
	return "login_history"
}


type LoginLast struct {
	Id         int       `xorm:"id pk autoincr"`
	UId        int       `xorm:"uid"`
	LoginTime  time.Time `xorm:"login_time"`
	LogoutTime time.Time `xorm:"logout_time"`
	Ip         string    `xorm:"ip"`
	Session    string    `xorm:"session"`
	IsLogout   int8      `xorm:"is_logout"`
	Hardware   string    `xorm:"hardware"`
}

func (this *LoginLast) TableName() string {
	return "login_last"
}