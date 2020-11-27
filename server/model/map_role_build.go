package model

import (
	"slgserver/server/conn"
	"slgserver/server/proto"
)

type MapRoleBuild struct {
	DB    		dbSync 		`xorm:"-"`
	Id    		int    		`xorm:"id pk autoincr"`
	RId   		int    		`xorm:"rid"`
	RNick		string		`xorm:"-"`
	Type  		int8   		`xorm:"type"`
	Level 		int8   		`xorm:"level"`
	X     		int    		`xorm:"x"`
	Y     		int    		`xorm:"y"`
	Name  		string 		`xorm:"name"`
	Wood  		int    		`xorm:"Wood"`
	Iron  		int    		`xorm:"iron"`
	Stone 		int    		`xorm:"stone"`
	Grain		int			`xorm:"grain"`
	CurDurable	int			`xorm:"cur_durable"`
	MaxDurable	int			`xorm:"max_durable"`
	Defender	int			`xorm:"defender"`
}

func (this *MapRoleBuild) TableName() string {
	return "map_role_build"
}


/* 推送同步 begin */
func (this*MapRoleBuild) IsCellView() bool{
	return true
}

func (this*MapRoleBuild) BelongToRId() []int{
	return []int{this.RId}
}

func (this*MapRoleBuild) PushMsgName() string{
	return "roleBuild.push"
}

func (this*MapRoleBuild) ToProto() interface{}{
	p := proto.MapRoleBuild{}
	p.RNick = this.RNick
	p.X = this.X
	p.Y = this.Y
	p.Type = this.Type
	p.CurDurable = this.CurDurable
	p.MaxDurable = this.MaxDurable
	p.Level = this.Level
	p.RId = this.RId
	p.Name = this.Name
	p.Defender = this.Defender
	return p
}

func (this*MapRoleBuild) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this*MapRoleBuild) Execute() {
	this.DB.Sync()
	this.Push()
}