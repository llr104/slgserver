package model

import (
	"encoding/json"
	"fmt"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"github.com/llr104/slgserver/server/slgserver/static_conf/skill"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

/*******db 操作begin********/
var dbSkillMgr *skillDBMgr

func init() {
	dbSkillMgr = &skillDBMgr{skills: make(chan *Skill, 100)}
	go dbSkillMgr.running()
}

type skillDBMgr struct {
	skills chan *Skill
}

func (this *skillDBMgr) running() {
	for true {
		select {
		case skill := <-this.skills:
			if skill.Id > 0 {
				_, err := db.MasterDB.Table(skill).ID(skill.Id).Cols(
					"cfgId", "belong_generals", "rid").Update(skill)
				if err != nil {
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			} else {
				db.MasterDB.Table(skill).InsertOne(skill)
			}
		}
	}
}

func (this *skillDBMgr) push(skill *Skill) {
	this.skills <- skill
}

/*******db 操作end********/

type Skill struct {
	Id             int    `xorm:"id pk autoincr"`
	RId            int    `xorm:"rid"`
	CfgId          int    `xorm:"cfgId"`
	BelongGenerals string `xorm:"belong_generals"`
	Generals       []int  `xorm:"-"`
}

func NewSkill(rid int, cfgId int) *Skill {
	return &Skill{
		CfgId:          cfgId,
		RId:            rid,
		Generals:       []int{},
		BelongGenerals: "[]",
	}
}

func (this *Skill) TableName() string {
	return "tb_skill" + fmt.Sprintf("_%d", ServerId)
}

func (this *Skill) AfterSet(name string, cell xorm.Cell) {
	if name == "belong_generals" {
		this.Generals = []int{}
		if cell != nil {
			gs, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(gs, &this.BelongGenerals)
			}
		}
	}
}

func (this *Skill) beforeModify() {
	data, _ := json.Marshal(this.Generals)
	this.BelongGenerals = string(data)
}

func (this *Skill) BeforeInsert() {
	this.beforeModify()
}

func (this *Skill) BeforeUpdate() {
	this.beforeModify()
}

/* 推送同步 begin */
func (this *Skill) IsCellView() bool {
	return false
}

func (this *Skill) IsCanView(rid, x, y int) bool {
	return false
}

func (this *Skill) BelongToRId() []int {
	return []int{this.RId}
}

func (this *Skill) PushMsgName() string {
	return "skill.push"
}

func (this *Skill) Position() (int, int) {
	return -1, -1
}

func (this *Skill) TPosition() (int, int) {
	return -1, -1
}

func (this *Skill) ToProto() interface{} {
	p := proto.Skill{}
	p.Id = this.Id
	p.CfgId = this.CfgId
	p.Generals = this.Generals
	return p
}
func (this *Skill) Push() {
	net.ConnMgr.Push(this)
}

/* 推送同步 end */

func (this *Skill) SyncExecute() {
	dbSkillMgr.push(this)
	this.Push()
}

func (this *Skill) Limit() int {
	cfg, _ := skill.Skill.GetCfg(this.CfgId)
	return cfg.Limit
}

func (this *Skill) IsInLimit() bool {
	//fmt.Println("this.BelongGenerals", this.BelongGenerals)
	return len(this.Generals) < this.Limit()
}

func (this *Skill) ArmyIsIn(armId int) bool {
	cfg, _ := skill.Skill.GetCfg(this.CfgId)
	for _, arm := range cfg.Arms {
		if arm == armId {
			return true
		}
	}
	return false
}

func (this *Skill) UpSkill(gId int) {
	this.Generals = append(this.Generals, gId)
	data, _ := json.Marshal(this.Generals)
	this.BelongGenerals = string(data)
}

func (this *Skill) DownSkill(gId int) {
	gs := make([]int, 0)
	for _, general := range this.Generals {
		if gId != general {
			gs = append(gs, general)
		}
	}
	this.Generals = gs

	data, _ := json.Marshal(this.Generals)
	this.BelongGenerals = string(data)
}
