package model

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/conn"
	"slgserver/server/slgserver/proto"
	"time"
	"xorm.io/xorm"
)


const (
	UnionDismiss	= 0 //解散
	UnionRunning	= 1 //运行中
)

/*******db 操作begin********/
var dbCoalitionMgr *coalitionDBMgr
func init() {
	dbCoalitionMgr = &coalitionDBMgr{coalitions: make(chan *Coalition, 100)}
	go dbCoalitionMgr.running()
}

type coalitionDBMgr struct {
	coalitions    chan *Coalition
}

func (this*coalitionDBMgr) running()  {
	for true {
		select {
		case coalition := <- this.coalitions:
			if coalition.Id >0 {
				_, err := db.MasterDB.Table(coalition).ID(coalition.Id).Cols("name",
					"members", "chairman", "vice_chairman", "notice", "state").Update(coalition)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update coalition fail, because id <= 0")
			}
		}
	}
}

func (this*coalitionDBMgr) push(coalition *Coalition)  {
	this.coalitions <- coalition
}
/*******db 操作end********/



type Coalition struct {
	Id           int       `xorm:"id pk autoincr"`
	Name         string    `xorm:"name"`
	Members      string    `xorm:"members"`
	MemberArray  []int     `xorm:"-"`
	CreateId     int       `xorm:"create_id"`
	Chairman     int       `xorm:"chairman"`
	ViceChairman int       `xorm:"vice_chairman"`
	Notice       string    `xorm:"notice"`
	State        int8      `xorm:"state"`
	Ctime        time.Time `xorm:"ctime"`
}


func (this *Coalition) ToProto() interface{}{
	p := proto.Union{}

	p.Id = this.Id
	p.Name = this.Name
	p.Notice = this.Notice
	p.Cnt = this.Cnt()

	return p
}

func (this *Coalition) TableName() string {
	return "tb_coalition" + fmt.Sprintf("_%d", ServerId)
}

func (this *Coalition) AfterSet(name string, cell xorm.Cell){
	if name == "members"{
		if cell != nil{
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.MemberArray)
			}
			if this.MemberArray == nil{
				this.MemberArray = []int{}
				fmt.Println(this.MemberArray)
			}
		}
	}
}

func (this *Coalition) BeforeInsert() {
	data, _ := json.Marshal(this.MemberArray)
	this.Members = string(data)
}

func (this *Coalition) BeforeUpdate() {
	data, _ := json.Marshal(this.MemberArray)
	this.Members = string(data)
}

func (this*Coalition) Cnt() int{
	return len(this.MemberArray)
}

func (this *Coalition) SyncExecute() {
	dbCoalitionMgr.push(this)
}


type CoalitionApply struct {
	Id      int       `xorm:"id pk autoincr"`
	UnionId int       `xorm:"union_id"`
	RId     int       `xorm:"rid"`
	State   int8      `xorm:"state"`
	Ctime   time.Time `xorm:"ctime"`
}

func (this *CoalitionApply) TableName() string {
	return "tb_coalition_apply" + fmt.Sprintf("_%d", ServerId)
}

/* 推送同步 begin */
func (this *CoalitionApply) IsCellView() bool{
	return false
}

func (this *CoalitionApply) IsCanView(rid, x, y int) bool{
	return false
}

func (this *CoalitionApply) BelongToRId() []int{
	r := GetMainMembers(this.UnionId)
	return append(r, this.RId)
}

func (this *CoalitionApply) PushMsgName() string{
	return "unionApply.push"
}

func (this *CoalitionApply) Position() (int, int){
	return -1, -1
}

func (this *CoalitionApply) TPosition() (int, int){
	return -1, -1
}

func (this *CoalitionApply) ToProto() interface{}{
	p := proto.ApplyItem{}
	p.RId = this.RId
	p.Id = this.Id
	p.NickName = GetRoleNickName(this.RId)
	return p
}

func (this *CoalitionApply) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *CoalitionApply) SyncExecute() {
	this.Push()
}