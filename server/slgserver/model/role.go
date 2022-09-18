package model

import (
	"fmt"
	"time"

	"github.com/llr104/slgserver/server/slgserver/proto"
)

type Role struct {
	RId			int			`xorm:"rid pk autoincr"`
	UId			int			`xorm:"uid"`
	NickName	string		`xorm:"nick_name" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	Balance		int			`xorm:"balance"`
	HeadId		int16		`xorm:"headId"`
	Sex			int8		`xorm:"sex"`
	Profile		string		`xorm:"profile"`
	LoginTime   time.Time	`xorm:"login_time"`
	LogoutTime  time.Time	`xorm:"logout_time"`
	CreatedAt	time.Time	`xorm:"created_at"`
}

func (this *Role) TableName() string {
	return "tb_role" + fmt.Sprintf("_%d", ServerId)
}

func (this *Role) ToProto() interface{}{
	p := proto.Role{}
	p.UId = this.UId
	p.RId = this.RId
	p.Sex = this.Sex
	p.NickName = this.NickName
	p.HeadId = this.HeadId
	p.Balance = this.Balance
	p.Profile = this.Profile
	return p
}
