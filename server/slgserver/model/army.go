package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/util"
	"go.uber.org/zap"
	"xorm.io/xorm"
)

const (
	ArmyCmdIdle        = 0 //空闲
	ArmyCmdAttack      = 1 //攻击
	ArmyCmdDefend      = 2 //驻守
	ArmyCmdReclamation = 3 //屯垦
	ArmyCmdBack        = 4 //撤退
	ArmyCmdConscript   = 5 //征兵
	ArmyCmdTransfer    = 6 //调动
)

const (
	ArmyStop    = 0
	ArmyRunning = 1
)

/*******db 操作begin********/
var dbArmyMgr *armyDBMgr

func init() {
	dbArmyMgr = &armyDBMgr{armys: make(chan *Army, 100)}
	go dbArmyMgr.running()
}

type armyDBMgr struct {
	armys chan *Army
}

func (this *armyDBMgr) running() {
	for true {
		select {
		case army := <-this.armys:
			if army.Id > 0 {
				_, err := db.MasterDB.Table(army).ID(army.Id).Cols(
					"soldiers", "generals", "conscript_times",
					"conscript_cnts", "cmd", "from_x", "from_y", "to_x",
					"to_y", "start", "end").Update(army)
				if err != nil {
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			} else {
				log.DefaultLog.Warn("update army fail, because id <= 0")
			}
		}
	}
}

func (this *armyDBMgr) push(army *Army) {
	this.armys <- army
}

/*******db 操作end********/

//军队
type Army struct {
	Id                 int                            `xorm:"id pk autoincr"`
	RId                int                            `xorm:"rid"`
	CityId             int                            `xorm:"cityId"`
	Order              int8                           `xorm:"order"`
	Generals           string                         `xorm:"generals"`
	Soldiers           string                         `xorm:"soldiers"`
	ConscriptTimes     string                         `xorm:"conscript_times"` //征兵结束时间，json数组
	ConscriptCnts      string                         `xorm:"conscript_cnts"`  //征兵数量，json数组
	Cmd                int8                           `xorm:"cmd"`
	FromX              int                            `xorm:"from_x"`
	FromY              int                            `xorm:"from_y"`
	ToX                int                            `xorm:"to_x"`
	ToY                int                            `xorm:"to_y"`
	Start              time.Time                      `json:"-"xorm:"start"`
	End                time.Time                      `json:"-"xorm:"end"`
	State              int8                           `xorm:"-"` //状态:0:running,1:stop
	GeneralArray       [static_conf.ArmyGCnt]int      `json:"-" xorm:"-"`
	SoldierArray       [static_conf.ArmyGCnt]int      `json:"-" xorm:"-"`
	ConscriptTimeArray [static_conf.ArmyGCnt]int64    `json:"-" xorm:"-"`
	ConscriptCntArray  [static_conf.ArmyGCnt]int      `json:"-" xorm:"-"`
	Gens               [static_conf.ArmyGCnt]*General `json:"-" xorm:"-"`
	CellX              int                            `json:"-" xorm:"-"`
	CellY              int                            `json:"-" xorm:"-"`
}

func (this *Army) TableName() string {
	return "tb_army" + fmt.Sprintf("_%d", ServerId)
}

//是否能出战
func (this *Army) IsCanOutWar() bool {
	return this.Gens[0] != nil && this.Cmd == ArmyCmdIdle
}

func (this *Army) AfterSet(name string, cell xorm.Cell) {
	if name == "generals" {
		this.GeneralArray = [static_conf.ArmyGCnt]int{0, 0, 0}
		if cell != nil {
			gs, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(gs, &this.GeneralArray)
				fmt.Println(this.GeneralArray)
			}
		}
	} else if name == "soldiers" {
		this.SoldierArray = [static_conf.ArmyGCnt]int{0, 0, 0}
		if cell != nil {
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.SoldierArray)
				fmt.Println(this.SoldierArray)
			}
		}
	} else if name == "conscript_times" {
		this.ConscriptTimeArray = [static_conf.ArmyGCnt]int64{0, 0, 0}
		if cell != nil {
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.ConscriptTimeArray)
				fmt.Println(this.ConscriptTimeArray)
			}
		}
	} else if name == "conscript_cnts" {
		this.ConscriptCntArray = [static_conf.ArmyGCnt]int{0, 0, 0}
		if cell != nil {
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.ConscriptCntArray)
				fmt.Println(this.ConscriptCntArray)
			}
		}
	}
}

func (this *Army) beforeModify() {
	data, _ := json.Marshal(this.GeneralArray)
	this.Generals = string(data)

	data, _ = json.Marshal(this.SoldierArray)
	this.Soldiers = string(data)

	data, _ = json.Marshal(this.ConscriptTimeArray)
	this.ConscriptTimes = string(data)

	data, _ = json.Marshal(this.ConscriptCntArray)
	this.ConscriptCnts = string(data)
}

func (this *Army) BeforeInsert() {
	this.beforeModify()
}

func (this *Army) BeforeUpdate() {
	this.beforeModify()
}

func (this *Army) ToSoldier() {
	data, _ := json.Marshal(this.SoldierArray)
	this.Soldiers = string(data)
}

func (this *Army) ToGeneral() {
	data, _ := json.Marshal(this.GeneralArray)
	this.Generals = string(data)
}

//获取军队阵营
func (this *Army) GetCamp() int8 {
	var camp int8 = 0
	for _, g := range this.Gens {
		if g == nil {
			return 0
		}
		if camp == 0 {
			camp = g.GetCamp()
		} else {
			if camp != g.GetCamp() {
				return 0
			}
		}
	}
	return camp
}

//检测征兵是否完成，服务器不做定时任务，用到的时候再检测
func (this *Army) CheckConscript() {
	if this.Cmd == ArmyCmdConscript {
		curTime := time.Now().Unix()
		finish := true
		for i, endTime := range this.ConscriptTimeArray {
			if endTime > 0 {
				if endTime <= curTime {
					this.SoldierArray[i] += this.ConscriptCntArray[i]
					this.ConscriptCntArray[i] = 0
					this.ConscriptTimeArray[i] = 0
				} else {
					finish = false
				}
			}
		}

		if finish {
			this.Cmd = ArmyCmdIdle
		}
	}
}

//队伍指定的位置是否能变化（上下阵）
func (this *Army) PositionCanModify(position int) bool {
	if position >= 3 || position < 0 {
		return false
	}

	if this.Cmd == ArmyCmdIdle {
		return true
	} else if this.Cmd == ArmyCmdConscript {
		endTime := this.ConscriptTimeArray[position]
		return endTime == 0
	} else {
		return false
	}
}

func (this *Army) ClearConscript() {
	if this.Cmd == ArmyCmdConscript {
		for i, _ := range this.ConscriptTimeArray {
			this.ConscriptCntArray[i] = 0
			this.ConscriptTimeArray[i] = 0
		}
		this.Cmd = ArmyCmdIdle
	}
}

func (this *Army) IsIdle() bool {
	return this.Cmd == ArmyCmdIdle
}

/* 推送同步 begin */
func (this *Army) IsCellView() bool {
	return true
}

func (this *Army) IsCanView(rid, x, y int) bool {
	if ArmyIsInView != nil {
		return ArmyIsInView(rid, x, y)
	}
	return false
}

func (this *Army) BelongToRId() []int {
	return []int{this.RId}
}

func (this *Army) PushMsgName() string {
	return "army.push"
}

func (this *Army) Position() (int, int) {
	diffTime := this.End.Unix() - this.Start.Unix()
	passTime := time.Now().Unix() - this.Start.Unix()
	rate := float32(passTime) / float32(diffTime)
	x := 0
	y := 0
	if this.Cmd == ArmyCmdBack {
		diffX := this.FromX - this.ToX
		diffY := this.FromY - this.ToY
		x = int(rate*float32(diffX)) + this.ToX
		y = int(rate*float32(diffY)) + this.ToY
	} else {
		diffX := this.ToX - this.FromX
		diffY := this.ToY - this.FromY
		x = int(rate*float32(diffX)) + this.FromX
		y = int(rate*float32(diffY)) + this.FromY
	}

	x = util.MinInt(util.MaxInt(x, 0), global.MapWith)
	y = util.MinInt(util.MaxInt(y, 0), global.MapHeight)
	log.DefaultLog.Info("army Position:", zap.Int("x", x), zap.Int("y", y))
	return x, y
}

func (this *Army) TPosition() (int, int) {
	return this.ToX, this.ToY
}

func (this *Army) ToProto() interface{} {

	p := proto.Army{}
	p.CityId = this.CityId
	p.Id = this.Id
	p.UnionId = GetUnionId(this.RId)
	p.Order = this.Order
	p.Generals = this.GeneralArray
	p.Soldiers = this.SoldierArray
	p.ConTimes = this.ConscriptTimeArray
	p.ConCnts = this.ConscriptCntArray
	p.Cmd = this.Cmd
	p.State = this.State
	p.FromX = this.FromX
	p.FromY = this.FromY
	p.ToX = this.ToX
	p.ToY = this.ToY
	p.Start = this.Start.Unix()
	p.End = this.End.Unix()
	return p
}

func (this *Army) Push() {
	net.ConnMgr.Push(this)
}

/* 推送同步 end */

func (this *Army) SyncExecute() {
	dbArmyMgr.push(this)
	this.Push()
	this.CellX, this.CellY = this.Position()
}

func (this *Army) CheckSyncCell() {
	x, y := this.Position()
	if x != this.CellX || y != this.CellY {
		this.SyncExecute()
	}
}
