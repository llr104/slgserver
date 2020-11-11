package model

import (
	"time"
)

const (
	Login = iota
	Logout
)

type LoginHistory struct {
	Id       int       `json:"id" xorm:"pk autoincr"`
	UId      int       `json:"uid" xorm:"uid"`
	Time     time.Time `json:"time"`
	Ip       string    `json:"ip"`
	State    int8      `json:"state"`
	Hardware string    `json:"hardware"`
}

func (this *LoginHistory) TableName() string {
	return "login_history"
}


type LoginLast struct {
	Id         int       `json:"id" xorm:"pk autoincr"`
	UId        int       `json:"uid" xorm:"uid"`
	LoginTime  time.Time `json:"login_time"`
	LogoutTime time.Time `json:"logout_time"`
	Ip         string    `json:"ip"`
	Session    string    `json:"session"`
	IsLogout   int8      `json:"is_logout"`
	Hardware   string    `json:"hardware"`
}

func (this *LoginLast) TableName() string {
	return "login_last"
}