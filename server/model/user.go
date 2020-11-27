package model

import "time"

type User struct {
	UId      int       `xorm:"uid pk autoincr"`
	Username string    `xorm:"username" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	Passcode string    `xorm:"passcode"`
	Passwd   string    `xorm:"passwd"`
	Hardware string    `xorm:"hardware"`
	Status   int       `xorm:"status"`
	Ctime    time.Time `xorm:"ctime"`
	Mtime    time.Time `xorm:"mtime"`
	IsOnline bool      `xorm:"-"`
}

func (this *User) TableName() string {
	return "user_info"
}