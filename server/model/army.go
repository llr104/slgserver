package model

import (
	"encoding/json"
	"slgserver/server/conn"
	"slgserver/server/proto"
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

//军队
type Army struct {
	DB           		dbSync 		`json:"-" xorm:"-"`
	Id           		int    		`xorm:"id pk autoincr"`
	RId          		int    		`xorm:"rid"`
	CityId       		int    		`xorm:"cityId"`
	Order        		int8   		`xorm:"order"`
	Generals     		string 		`xorm:"generals"`
	Soldiers     		string 		`xorm:"soldiers"`
	GeneralArray 		[]int  		`json:"-" xorm:"-"`
	SoldierArray 		[]int  		`json:"-" xorm:"-"`
	Cmd          		int8   		`xorm:"cmd"` //执行命令0:空闲 1:攻击 2：驻军 3:返回
	State        		int8   		`xorm:"-"` //状态:0:running,1:stop
	FromX        		int    		`xorm:"from_x"`
	FromY            	int       	`xorm:"from_y"`
	ToX              	int       	`xorm:"to_x"`
	ToY              	int       	`xorm:"to_y"`
	Start            	time.Time 	`json:"-"xorm:"start"`
	End              	time.Time 	`json:"-"xorm:"end"`
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

func (this*Army) ToSoldier() {
	if this.SoldierArray != nil {
		data, _ := json.Marshal(this.SoldierArray)
		this.Soldiers = string(data)
	}
}

func (this*Army) ToGeneral() {
	if this.GeneralArray != nil {
		data, _ := json.Marshal(this.GeneralArray)
		this.Generals = string(data)
	}
}

func (this*Army) BeforeInsert() {

	data, _ := json.Marshal(this.GeneralArray)
	this.Generals = string(data)

	data, _ = json.Marshal(this.SoldierArray)
	this.Soldiers = string(data)
}

func (this*Army) BeforeUpdate() {

	data, _ := json.Marshal(this.GeneralArray)
	this.Generals = string(data)

	data, _ = json.Marshal(this.SoldierArray)
	this.Soldiers = string(data)
}

/* 推送同步 begin */
func (this*Army) IsCellView() bool{
	return true
}

func (this*Army) BelongToRId() []int{
	return []int{this.RId}
}

func (this*Army) PushMsgName() string{
	return "army.push"
}

func (this*Army) ToProto() interface{}{
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

func (this*Army) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this*Army) SyncExecute() {
	this.DB.Sync()
	this.Push()
}