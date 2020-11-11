package model

import "time"

type User struct {
	UId      int       `json:"uid" xorm:"pk autoincr uid"`
	Username string    `json:"username" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	Passcode string    `json:"passcode"`
	Passwd   string    `json:"passwd"`
	Hardware string    `json:"hardware"`
	Status   int       `json:"status"`
	Ctime    time.Time `json:"ctime" xorm:"created"`
	Mtime    time.Time `json:"mtime" xorm:"<-"`
	IsOnline bool      `json:"is_online" xorm:"-"`
}

func (this *User) TableName() string {
	return "user_info"
}