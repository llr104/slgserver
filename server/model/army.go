package model

import (
	"encoding/json"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/conn"
	"slgserver/server/global"
	"slgserver/server/proto"
	"slgserver/util"
	"time"
	"xorm.io/xorm"
)

const (
	ArmyCmdIdle   		= 0	//空闲
	ArmyCmdAttack 		= 1	//攻击
	ArmyCmdDefend 		= 2	//驻守
	ArmyCmdReclamation 	= 3	//屯垦
	ArmyCmdBack   		= 4 //撤退
)

const (
	ArmyStop  		= 0
	ArmyRunning  	= 1
)

/*******db 操作begin********/
var dbArmyMgr *armyDBMgr
func init() {
	dbArmyMgr = &armyDBMgr{armys: make(chan *Army, 100)}
	go dbArmyMgr.running()
}

type armyDBMgr struct {
	armys    chan *Army
}

func (this* armyDBMgr) running()  {
	for true {
		select {
		case army := <- this.armys:
			if army.Id >0 {
				_, err := db.MasterDB.Table(army).ID(army.Id).Cols("soldiers",
					"generals", "cmd", "from_x", "from_y", "to_x", "to_y", "start", "end").Update(army)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update army fail, because id <= 0")
			}
		}
	}
}

func (this* armyDBMgr) push(army *Army)  {
	this.armys <- army
}
/*******db 操作end********/

//军队
type Army struct {
	Id           int        `xorm:"id pk autoincr"`
	RId          int        `xorm:"rid"`
	CityId       int        `xorm:"cityId"`
	Order        int8       `xorm:"order"`
	Generals     string     `xorm:"generals"`
	Soldiers     string     `xorm:"soldiers"`
	GeneralArray []int      `json:"-" xorm:"-"`
	SoldierArray []int      `json:"-" xorm:"-"`
	Gens         []*General `json:"-" xorm:"-"`
	Cmd          int8       `xorm:"cmd"` //执行命令0:空闲 1:攻击 2：驻军 3:返回
	State        int8       `xorm:"-"` //状态:0:running,1:stop
	FromX        int        `xorm:"from_x"`
	FromY        int        `xorm:"from_y"`
	ToX          int        `xorm:"to_x"`
	ToY          int        `xorm:"to_y"`
	Start        time.Time  `json:"-"xorm:"start"`
	End          time.Time  `json:"-"xorm:"end"`
	CellX        int        `json:"-" xorm:"-"`
	CellY        int        `json:"-" xorm:"-"`
}

func (this *Army) TableName() string {
	return "army"
}

func (this *Army) AfterSet(name string, cell xorm.Cell){
	if name == "generals"{
		this.GeneralArray = []int{0,0,0}
		if cell != nil{
			gs, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(gs, &this.GeneralArray)
			}
		}
	}else if name == "soldiers"{
		this.SoldierArray = []int{0,0,0}
		if cell != nil{
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.SoldierArray)
			}
		}
	}
}

func (this *Army) ToSoldier() {
	if this.SoldierArray != nil {
		data, _ := json.Marshal(this.SoldierArray)
		this.Soldiers = string(data)
	}
}

func (this *Army) ToGeneral() {
	if this.GeneralArray != nil {
		data, _ := json.Marshal(this.GeneralArray)
		this.Generals = string(data)
	}
}

func (this *Army) BeforeInsert() {

	data, _ := json.Marshal(this.GeneralArray)
	this.Generals = string(data)

	data, _ = json.Marshal(this.SoldierArray)
	this.Soldiers = string(data)
}

func (this *Army) BeforeUpdate() {

	data, _ := json.Marshal(this.GeneralArray)
	this.Generals = string(data)

	data, _ = json.Marshal(this.SoldierArray)
	this.Soldiers = string(data)
}

/* 推送同步 begin */
func (this *Army) IsCellView() bool{
	return true
}

func (this *Army) BelongToRId() []int{
	return []int{this.RId}
}

func (this *Army) PushMsgName() string{
	return "army.push"
}

func (this *Army) Position() (int, int){
	diffTime := this.End.Unix()-this.Start.Unix()
	passTime := time.Now().Unix()-this.Start.Unix()
	rate := float32(passTime)/float32(diffTime)
	x := 0
	y := 0
	if this.Cmd == ArmyCmdBack{
		diffX := this.ToX - this.FromX
		diffY := this.ToX - this.FromY
		x = int(rate*float32(diffX)) + this.FromX
		y = int(rate*float32(diffY)) + this.FromY
	}else{
		diffX := this.FromX - this.ToX
		diffY := this.FromY - this.ToY
		x = int(rate*float32(diffX)) + this.FromX
		y = int(rate*float32(diffY)) + this.FromY
	}

	x = util.MinInt(util.MaxInt(x, 0), global.MapWith)
	y = util.MinInt(util.MaxInt(y, 0), global.MapHeight)
	return x, y
}

func (this *Army) ToProto() interface{}{
	p := proto.Army{}
	p.CityId = this.CityId
	p.Id = this.Id
	p.Order = this.Order
	p.Generals = this.GeneralArray
	p.Soldiers = this.SoldierArray
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

func (this *Army) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *Army) SyncExecute() {
	dbArmyMgr.push(this)
	this.Push()
	this.CellX, this.CellY = this.Position()
}

func (this *Army) CheckSyncCell() {
	x, y := this.Position()
	if x != this.CellX || y != this.CellY{
		this.SyncExecute()
	}
}

