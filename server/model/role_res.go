package model

import (
	"slgserver/server/conn"
	"slgserver/server/proto"
)

type RoleRes struct {
	DB        dbSync `xorm:"-"`
	Id        int    `xorm:"id pk autoincr"`
	RId       int    `xorm:"rid"`
	Wood      int    `xorm:"wood"`
	Iron      int    `xorm:"iron"`
	Stone     int    `xorm:"stone"`
	Grain     int    `xorm:"grain"`
	Gold      int    `xorm:"gold"`
	Decree    int    `xorm:"decree"`	//令牌
	WoodYield int    `xorm:"wood_yield"`
	IronYield int    `xorm:"iron_yield"`
	StoneYield		int			`xorm:"stone_yield"`
	GrainYield		int			`xorm:"grain_yield"`
	GoldYield		int			`xorm:"gold_yield"`
	DepotCapacity	int			`xorm:"depot_capacity"`	//仓库容量
}

func (this *RoleRes) TableName() string {
	return "role_res"
}


/* 推送同步 begin */
func (this*RoleRes) IsCellView() bool{
	return false
}

func (this*RoleRes) BelongToRId() []int{
	return []int{this.RId}
}

func (this*RoleRes) PushMsgName() string{
	return "roleRes.push"
}

func (this*RoleRes) ToProto() interface{}{
	p := proto.RoleRes{}
	p.Gold = this.Gold
	p.Grain = this.Grain
	p.Stone = this.Stone
	p.Iron = this.Iron
	p.Wood = this.Wood
	p.Decree = this.Decree
	p.GoldYield = this.GoldYield
	p.GrainYield = this.GrainYield
	p.StoneYield = this.StoneYield
	p.IronYield = this.IronYield
	p.WoodYield = this.WoodYield
	p.DepotCapacity = this.DepotCapacity
	return p
}

func (this*RoleRes) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this*RoleRes) Execute() {
	this.DB.Sync()
	this.Push()
}