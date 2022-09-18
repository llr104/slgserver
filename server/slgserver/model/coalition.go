package model

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"go.uber.org/zap"
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
	net.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *CoalitionApply) SyncExecute() {
	this.Push()
}

const (
	UnionOpCreate		= 0 //创建
	UnionOpDismiss		= 1 //解散
	UnionOpJoin			= 2 //加入
	UnionOpExit			= 3 //退出
	UnionOpKick			= 4 //踢出
	UnionOpAppoint		= 5 //任命
	UnionOpAbdicate		= 6 //禅让
	UnionOpModNotice	= 7 //修改公告
)

type CoalitionLog struct {
	Id      	int       	`xorm:"id pk autoincr"`
	UnionId 	int       	`xorm:"union_id"`
	OPRId   	int       	`xorm:"op_rid"`
	TargetId   	int       	`xorm:"target_id"`
	State   	int8      	`xorm:"state"`
	Des			string		`xorm:"des"`
	Ctime   	time.Time 	`xorm:"ctime"`
}

func (this *CoalitionLog) TableName() string {
	return "tb_coalition_log" + fmt.Sprintf("_%d", ServerId)
}

func (this *CoalitionLog) ToProto() interface{}{
	p := proto.UnionLog{}
	p.OPRId = this.OPRId
	p.TargetId = this.TargetId
	p.Des = this.Des
	p.State = this.State
	p.Ctime = this.Ctime.UnixNano()/1e6
	return p
}


func NewCreate(opNickName string, unionId int, opRId int)  {
	ulog := &CoalitionLog{
		UnionId: unionId,
		OPRId: opRId,
		TargetId: 0,
		State: UnionOpCreate,
		Des: opNickName + " 创建了联盟",
		Ctime: time.Now(),
	}

	db.MasterDB.InsertOne(ulog)
}

func NewDismiss(opNickName string, unionId int, opRId int) {
	ulog := &CoalitionLog{
		UnionId: unionId,
		OPRId: opRId,
		TargetId: 0,
		State: UnionOpDismiss,
		Des: opNickName + " 解散了联盟",
		Ctime: time.Now(),
	}
	db.MasterDB.InsertOne(ulog)
}

func NewJoin(targetNickName string, unionId int, opRId int, targetId int) {
	ulog := &CoalitionLog{
		UnionId: unionId,
		OPRId: opRId,
		TargetId: targetId,
		State: UnionOpJoin,
		Des: targetNickName + " 加入了联盟",
		Ctime: time.Now(),
	}
	db.MasterDB.InsertOne(ulog)
}

func NewExit(opNickName string, unionId int, opRId int) {
	ulog := &CoalitionLog{
		UnionId: unionId,
		OPRId: opRId,
		TargetId: opRId,
		State: UnionOpExit,
		Des: opNickName + " 退出了联盟",
		Ctime: time.Now(),
	}
	db.MasterDB.InsertOne(ulog)
}

func NewKick(opNickName string, targetNickName string, unionId int, opRId int, targetId int) {
	ulog := &CoalitionLog{
		UnionId: unionId,
		OPRId: opRId,
		TargetId: targetId,
		State: UnionOpKick,
		Des: opNickName + " 将 " + targetNickName + " 踢出了联盟",
		Ctime: time.Now(),
	}
	db.MasterDB.InsertOne(ulog)
}

func NewAppoint(opNickName string, targetNickName string,
	unionId int, opRId int, targetId int, memberType int) {

	title := ""
	if memberType == proto.UnionChairman{
		title = "盟主"
	}else if memberType == proto.UnionViceChairman{
		title = "副盟主"
	}else{
		title = "普通成员"
	}

	ulog := &CoalitionLog{
		UnionId: unionId,
		OPRId: opRId,
		TargetId: targetId,
		State: UnionOpAppoint,
		Des: opNickName + " 将 " + targetNickName + " 任命为 " + title,
		Ctime: time.Now(),
	}
	db.MasterDB.InsertOne(ulog)
}

func NewAbdicate(opNickName string, targetNickName string,
	unionId int, opRId int, targetId int, memberType int) {

	title := ""
	if memberType == proto.UnionChairman{
		title = "盟主"
	}else if memberType == proto.UnionViceChairman{
		title = "副盟主"
	}else{
		title = "普通成员"
	}

	ulog := &CoalitionLog{
		UnionId: unionId,
		OPRId: opRId,
		TargetId: targetId,
		State: UnionOpAbdicate,
		Des: opNickName + " 将 " + title + " 禅让给 "  + targetNickName,
		Ctime: time.Now(),
	}
	db.MasterDB.InsertOne(ulog)
}

func NewModNotice(opNickName string, unionId int, opRId int) {
	ulog := &CoalitionLog{
		UnionId: unionId,
		OPRId: opRId,
		TargetId: 0,
		State: UnionOpModNotice,
		Des: opNickName + " 修改了公告",
		Ctime: time.Now(),
	}
	db.MasterDB.InsertOne(ulog)
}