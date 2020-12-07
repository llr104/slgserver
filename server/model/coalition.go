package model

import (
	"encoding/json"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
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

func (this* coalitionDBMgr) running()  {
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

func (this* coalitionDBMgr) push(coalition *Coalition)  {
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


func (this *Coalition) TableName() string {
	return "coalition"
}

func (this *Coalition) AfterSet(name string, cell xorm.Cell){
	if name == "members"{
		if cell != nil{
			ss, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(ss, &this.MemberArray)
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

func (this* Coalition) Cnt() int{
	return len(this.MemberArray)
}

func (this *Coalition) SyncExecute() {
	dbCoalitionMgr.push(this)
}


type CoalitionApply struct {
	Id          int       `xorm:"id pk autoincr"`
	CoalitionId int       `xorm:"coalition_id"`
	RId         int       `xorm:"rid"`
	State       int8      `xorm:"state"`
	Ctime       time.Time `xorm:"ctime"`
}

func (this *CoalitionApply) TableName() string {
	return "coalition_apply"
}


