package model

import (
	"slgserver/server/conn"
	"slgserver/server/proto"
	"slgserver/server/static_conf/general"
	"time"
)

type General struct {
	DB            dbSync    `xorm:"-"`
	Id            int       `xorm:"id pk autoincr"`
	RId           int       `xorm:"rid"`
	CfgId         int       `xorm:"cfgId"`
	PhysicalPower int       `xorm:"physical_power"`
	Level         int8      `xorm:"level"`
	Cost          int       `xorm:"cost"`
	Exp           int       `xorm:"exp"`
	Order         int8      `xorm:"order"`
	CityId        int       `xorm:"cityId"`
	CreatedAt     time.Time `xorm:"created_at"`
	CurArms       int       `xorm:"arms"`
	HasPrPoint    int       `xorm:"has_pr_point"`
	UsePrPoint    int       `xorm:"use_pr_point"`
	AttackDis     int       `xorm:"attack_distance"`
	ForceAdded    int       `xorm:"force_added"`
	StrategyAdded int       `xorm:"strategy_added"`
	DefenseAdded  int       `xorm:"defense_added"`
	SpeedAdded    int       `xorm:"speed_added"`
	DestroyAdded  int       `xorm:"destroy_added"`


}

func (this *General) TableName() string {
	return "general"
}

func (this *General) GetDestroy() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return (cfg.Destroy+cfg.DestroyGrow*int(this.Level))/100
	}
	return 0
}

func (this *General) GetSpeed() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return (cfg.Speed+cfg.SpeedGrow*int(this.Level))/100
	}
	return 0
}


/* 推送同步 begin */
func (this *General) IsCellView() bool{
	return false
}

func (this *General) BelongToRId() []int{
	return []int{this.RId}
}

func (this *General) PushMsgName() string{
	return "general.push"
}

func (this *General) Position() (int, int){
	return -1, -1
}

func (this *General) ToProto() interface{}{
	p := proto.General{}
	p.CityId = this.CityId
	p.Order = this.Order
	p.Cost = this.Cost
	p.PhysicalPower = this.PhysicalPower
	p.Id = this.Id
	p.CfgId = this.CfgId
	p.Level = this.Level
	p.Exp = this.Exp
	return p
}

func (this *General) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *General) SyncExecute() {
	this.DB.Sync()
	this.Push()
}