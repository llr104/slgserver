package model

import (
	"slgserver/server/conn"
	"slgserver/server/proto"
	"time"
)

type MapRoleCity struct {
	DB          dbSync 		`json:"-" xorm:"-"`
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

/* 推送同步 begin */
func (this*MapRoleCity) IsCellView() bool{
	return true
}

func (this*MapRoleCity) BelongToRId() []int{
	return []int{this.RId}
}

func (this*MapRoleCity) PushMsgName() string{
	return "roleCity.push"
}

func (this*MapRoleCity) ToProto() interface{}{
	p := proto.MapRoleCity{}
	p.X = this.X
	p.Y = this.Y
	p.CityId = this.CityId
	p.CurDurable = this.CurDurable
	p.MaxDurable = this.MaxDurable
	p.Level = this.Level
	p.RId = this.RId
	p.Name = this.Name
	p.IsMain = this.IsMain == 1
	return p
}

func (this*MapRoleCity) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this*MapRoleCity) SyncExecute() {
	this.DB.Sync()
	this.Push()
}