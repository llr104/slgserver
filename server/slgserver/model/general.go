package model

import (
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/net"
	"slgserver/server/slgserver/proto"
	"slgserver/server/slgserver/static_conf/general"
	"time"
)


const (
	GeneralNormal      	= 0 //正常
	GeneralComposeStar 	= 1 //星级合成
	GeneralConvert 		= 2 //转换
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


func (this*generalDBMgr) running()  {
	for true {
		select {
		case g := <- this.gs:
			if g.Id > 0 && g.RId > 0 {
				_, err := db.MasterDB.Table(g).ID(g.Id).Cols(
					"level", "exp", "order", "cityId",
					"physical_power", "star_lv", "has_pr_point",
					"use_pr_point", "force_added", "strategy_added",
					"defense_added", "speed_added", "destroy_added",
					"parentId", "compose_type", "state").Update(g)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update general fail, because id <= 0")
			}
		}
	}
}

func (this*generalDBMgr) push(g *General)  {
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
	AttackDis     int  		`xorm:"attack_distance"`
	ForceAdded    int  		`xorm:"force_added"`
	StrategyAdded int  		`xorm:"strategy_added"`
	DefenseAdded  int  		`xorm:"defense_added"`
	SpeedAdded    int  		`xorm:"speed_added"`
	DestroyAdded  int  		`xorm:"destroy_added"`
	StarLv        int8  	`xorm:"star_lv"`
	Star          int8  	`xorm:"star"`
	ParentId      int  		`xorm:"parentId"`
	State         int8 		`xorm:"state"`
}

func (this *General) TableName() string {
	return "tb_general" + fmt.Sprintf("_%d", ServerId)
}

func (this *General) GetDestroy() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Destroy+cfg.DestroyGrow*int(this.Level) + this.DestroyAdded
	}
	return 0
}

func (this* General) IsActive() bool  {
	return this.State == GeneralNormal
}

func (this *General) GetSpeed() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Speed+cfg.SpeedGrow*int(this.Level) + this.SpeedAdded
	}
	return 0
}

func (this *General) GetForce() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Force+cfg.ForceGrow*int(this.Level) + this.ForceAdded
	}
	return 0
}

func (this *General) GetDefense() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Defense+cfg.DefenseGrow*int(this.Level) + this.DefenseAdded
	}
	return 0
}

func (this *General) GetStrategy() int{
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Strategy+cfg.StrategyGrow*int(this.Level) + this.StrategyAdded
	}
	return 0
}

//获取阵营
func (this*General) GetCamp() int8 {
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Camp
	}
	return 0
}

/* 推送同步 begin */
func (this *General) IsCellView() bool{
	return false
}

func (this *General) IsCanView(rid, x, y int) bool{
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

func (this *General) TPosition() (int, int){
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
	p.State = this.State
	p.ParentId = this.ParentId
	return p
}

func (this *General) Push(){
	net.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *General) SyncExecute() {
	dbGeneralMgr.push(this)
	this.Push()
}