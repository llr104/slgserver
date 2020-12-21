package logic

import "time"

type User struct {
	rid			int
	nickName 	string
}

func NewUser(rid int, nickName string) *User {
	return &User{
		rid: rid,
		nickName: nickName,
	}
}

type Msg struct {
	RId      int		`json:"rid"`
	NickName string		`json:"nickName"`
	Msg      string		`json:"msg"`
	Time     time.Time	`json:"time"`
}