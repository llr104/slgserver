package model

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/server/slgserver/static_conf/general"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

const (
	GeneralNormal      = 0 //正常
	GeneralComposeStar = 1 //星级合成
	GeneralConvert     = 2 //转换
)

/*******db 操作begin********/
var dbGeneralMgr *generalDBMgr

func init() {
	dbGeneralMgr = &generalDBMgr{gs: make(chan *General, 100)}
	go dbGeneralMgr.running()
}

type generalDBMgr struct {
	gs chan *General
}

func (this *generalDBMgr) running() {
	for true {
		select {
		case g := <-this.gs:
			if g.Id > 0 && g.RId > 0 {
				_, err := db.MasterDB.Table(g).ID(g.Id).Cols(
					"level", "exp", "order", "cityId",
					"physical_power", "star_lv", "has_pr_point",
					"use_pr_point", "force_added", "strategy_added",
					"defense_added", "speed_added", "destroy_added",
					"parentId", "compose_type", "skills", "state").Update(g)
				if err != nil {
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			} else {
				log.DefaultLog.Warn("update general fail, because id <= 0")
			}
		}
	}
}

func (this *generalDBMgr) push(g *General) {
	this.gs <- g
}

/*******db 操作end********/

const SkillLimit = 3

func NewGeneral(cfgId int, rid int, level int8) (*General, bool) {

	cfg, ok := general.General.GMap[cfgId]
	if ok {
		sa := make([]*proto.GSkill, SkillLimit)
		ss, _ := json.Marshal(sa)
		g := &General{
			PhysicalPower: static_conf.Basic.General.PhysicalPowerLimit,
			RId:           rid,
			CfgId:         cfg.CfgId,
			Order:         0,
			CityId:        0,
			Level:         level,
			CreatedAt:     time.Now(),
			CurArms:       cfg.Arms[0],
			HasPrPoint:    0,
			UsePrPoint:    0,
			AttackDis:     0,
			ForceAdded:    0,
			StrategyAdded: 0,
			DefenseAdded:  0,
			SpeedAdded:    0,
			DestroyAdded:  0,
			Star:          cfg.Star,
			StarLv:        0,
			ParentId:      0,
			SkillsArray:   sa,
			Skills:        string(ss),
			State:         GeneralNormal,
		}

		if _, err := db.MasterDB.Table(General{}).Insert(g); err != nil {
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		} else {
			return g, true
		}
	} else {
		return nil, false
	}
}

type General struct {
	Id            int             `xorm:"id pk autoincr"`
	RId           int             `xorm:"rid"`
	CfgId         int             `xorm:"cfgId"`
	PhysicalPower int             `xorm:"physical_power"`
	Level         int8            `xorm:"level"`
	Exp           int             `xorm:"exp"`
	Order         int8            `xorm:"order"`
	CityId        int             `xorm:"cityId"`
	CreatedAt     time.Time       `xorm:"created_at"`
	CurArms       int             `xorm:"arms"`
	HasPrPoint    int             `xorm:"has_pr_point"`
	UsePrPoint    int             `xorm:"use_pr_point"`
	AttackDis     int             `xorm:"attack_distance"`
	ForceAdded    int             `xorm:"force_added"`
	StrategyAdded int             `xorm:"strategy_added"`
	DefenseAdded  int             `xorm:"defense_added"`
	SpeedAdded    int             `xorm:"speed_added"`
	DestroyAdded  int             `xorm:"destroy_added"`
	StarLv        int8            `xorm:"star_lv"`
	Star          int8            `xorm:"star"`
	ParentId      int             `xorm:"parentId"`
	Skills        string          `xorm:"skills"`
	SkillsArray   []*proto.GSkill `xorm:"-"`
	State         int8            `xorm:"state"`
}

func (this *General) TableName() string {
	return "tb_general" + fmt.Sprintf("_%d", ServerId)
}

func (this *General) AfterSet(name string, cell xorm.Cell) {
	if name == "skills" {
		this.SkillsArray = make([]*proto.GSkill, 3)
		if cell != nil {
			gs, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(gs, &this.SkillsArray)

				if this.SkillsArray[0] != nil {
					fmt.Println(this.SkillsArray)
				}
			}
		}
	}
}

func (this *General) beforeModify() {
	data, _ := json.Marshal(this.SkillsArray)
	this.Skills = string(data)
}

func (this *General) BeforeInsert() {
	this.beforeModify()
}

func (this *General) BeforeUpdate() {
	this.beforeModify()
}

func (this *General) GetDestroy() int {
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Destroy + cfg.DestroyGrow*int(this.Level) + this.DestroyAdded
	}
	return 0
}

func (this *General) IsActive() bool {
	return this.State == GeneralNormal
}

func (this *General) GetSpeed() int {
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Speed + cfg.SpeedGrow*int(this.Level) + this.SpeedAdded
	}
	return 0
}

func (this *General) GetForce() int {
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Force + cfg.ForceGrow*int(this.Level) + this.ForceAdded
	}
	return 0
}

func (this *General) GetDefense() int {
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Defense + cfg.DefenseGrow*int(this.Level) + this.DefenseAdded
	}
	return 0
}

func (this *General) GetStrategy() int {
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Strategy + cfg.StrategyGrow*int(this.Level) + this.StrategyAdded
	}
	return 0
}

//获取阵营
func (this *General) GetCamp() int8 {
	cfg, ok := general.General.GMap[this.CfgId]
	if ok {
		return cfg.Camp
	}
	return 0
}

func (this *General) UpSkill(skillId int, cfgId int, pos int) bool {
	if pos < 0 || pos >= SkillLimit {
		return false
	}

	for _, skill := range this.SkillsArray {
		if skill != nil && skill.Id == skillId {
			//已经上过同类型的技能了
			return false
		}
	}

	s := this.SkillsArray[pos]
	if s == nil {
		this.SkillsArray[pos] = &proto.GSkill{Id: skillId, Lv: 1, CfgId: cfgId}
		return true
	} else {
		if s.Id == 0 {
			s.Id = skillId
			s.CfgId = cfgId
			s.Lv = 1
			return true
		} else {
			return false
		}
	}
}

func (this *General) DownSkill(skillId int, pos int) bool {
	if pos < 0 || pos >= SkillLimit {
		return false
	}
	s := this.SkillsArray[pos]
	if s != nil && s.Id == skillId {
		s.Id = 0
		s.Lv = 0
		s.CfgId = 0
		return true
	} else {
		return false
	}
}

func (this *General) PosSkill(pos int) (*proto.GSkill, error) {
	if pos >= len(this.SkillsArray) {
		return nil, errors.New("skill index out of range")
	}
	return this.SkillsArray[pos], nil
}

/* 推送同步 begin */
func (this *General) IsCellView() bool {
	return false
}

func (this *General) IsCanView(rid, x, y int) bool {
	return false
}

func (this *General) BelongToRId() []int {
	return []int{this.RId}
}

func (this *General) PushMsgName() string {
	return "general.push"
}

func (this *General) Position() (int, int) {
	return -1, -1
}

func (this *General) TPosition() (int, int) {
	return -1, -1
}

func (this *General) ToProto() interface{} {
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
	p.Skills = this.SkillsArray
	return p
}

func (this *General) Push() {
	net.ConnMgr.Push(this)
}

/* 推送同步 end */

func (this *General) SyncExecute() {
	dbGeneralMgr.push(this)
	this.Push()
}
