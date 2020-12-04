package model

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/conn"
	"slgserver/server/proto"
	"slgserver/server/static_conf/general"
	"time"
)


const (
	ComposeNot   		= 0	//没有合成
	ComposeStar 		= 1	//星级合成
)

/*******db 操作begin********/
var dbGeneralMgr *generalDBMgr
func init() {
	dbGeneralMgr = &generalDBMgr{gs: make(chan *General, 100)}
	go dbGeneralMgr.running()
}

type generalDBMgr struct {
	gs   chan *General
}

func (this* generalDBMgr) running()  {
	for true {
		select {
		case g := <- this.gs:
			if g.Id > 0 && g.RId > 0 {
				_, err := db.MasterDB.Table(g).ID(g.Id).Cols("level",
					"exp", "order", "cityId", "physical_power","star_lv","has_pr_point",
					"use_pr_point","force_added","strategy_added","defense_added","speed_added",
					"destroy_added","parentId","compose_type").Update(g)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update general fail, because id <= 0")
			}
		}
	}
}

func (this* generalDBMgr) push(g *General)  {
	this.gs <- g
}
/*******db 操作end********/


type General struct {
	Id            int       `xorm:"id pk autoincr"`
	RId           int       `xorm:"rid"`
	CfgId         int       `xorm:"cfgId"`
	PhysicalPower int       `xorm:"physical_power"`
	Level         int8      `xorm:"level"`
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
	StarLv        int       `xorm:"star_lv"`
	Star          int       `xorm:"star"`
	ParentId      int       `xorm:"parentId"`
	ComposeType   int       `xorm:"compose_type"`
}

func (this *General) TableName() string {
	return "general"
}

func (this *General) GetDestroy() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return (cfg.Destroy+cfg.DestroyGrow*int(this.Level))/100 + this.DestroyAdded
	}
	return 0
}

func (this *General) GetSpeed() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return (cfg.Speed+cfg.SpeedGrow*int(this.Level))/100 + this.SpeedAdded
	}
	return 0
}

func (this *General) GetForce() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return (cfg.Force+cfg.ForceGrow*int(this.Level))/100 + this.ForceAdded
	}
	return 0
}

func (this *General) GetDefense() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return (cfg.Defense+cfg.DefenseGrow*int(this.Level))/100 + this.DefenseAdded
	}
	return 0
}

func (this *General) GetStrategy() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return (cfg.Strategy+cfg.StrategyGrow*int(this.Level))/100 + this.StrategyAdded
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
	p.PhysicalPower = this.PhysicalPower
	p.Id = this.Id
	p.CfgId = this.CfgId
	p.Level = this.Level
	p.Exp = this.Exp
	p.CurArms = this.CurArms
	p.HasPrPoint = this.HasPrPoint
	p.UsePrPoint = this.UsePrPoint
	p.AttackDis = this.AttackDis
	p.ForceAdded = this.ForceAdded
	p.StrategyAdded = this.StrategyAdded
	p.DefenseAdded = this.DefenseAdded
	p.SpeedAdded = this.SpeedAdded
	p.DestroyAdded = this.DestroyAdded
	p.StarLv = this.StarLv
	p.Star = this.Star
	p.ComposeType = this.ComposeType
	p.ParentId = this.ParentId
	return p
}

func (this *General) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *General) SyncExecute() {
	dbGeneralMgr.push(this)
	this.Push()
}