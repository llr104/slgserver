package model

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/conn"
	"slgserver/server/proto"
	"time"
)


/*******db 操作begin********/
var dbWarReportMgr *warReportDBMgr
func init() {
	dbWarReportMgr = &warReportDBMgr{reports: make(chan *WarReport, 100)}
	go dbWarReportMgr.running()
}

type warReportDBMgr struct {
	reports    chan *WarReport
}

func (this* warReportDBMgr) running()  {
	for true {
		select {
		case r := <- this.reports:
			if r.Id ==0 {
				_, err := db.MasterDB.InsertOne(r)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}
		}
	}
}

func (this* warReportDBMgr) push(r *WarReport)  {
	this.reports <- r
}
/*******db 操作end********/

type WarReport struct {
	Id                int    	`xorm:"id pk autoincr"`
	AttackRid         int    	`xorm:"attack_rid"`
	DefenseRid        int    	`xorm:"defense_rid"`
	BegAttackArmy     string 	`xorm:"beg_attack_army"`
	BegDefenseArmy    string 	`xorm:"beg_defense_army"`
	EndAttackArmy     string 	`xorm:"end_attack_army"`
	EndDefenseArmy    string 	`xorm:"end_defense_army"`
	BegAttackGeneral  string 	`xorm:"beg_attack_general"`
	BegDefenseGeneral string 	`xorm:"beg_defense_general"`
	EndAttackGeneral  string 	`xorm:"end_attack_general"`
	EndDefenseGeneral string    `xorm:"end_defense_general"`
	AttackIsWin       bool      `xorm:"attack_is_win"`
	AttackIsRead      bool      `xorm:"attack_is_read"`
	DefenseIsRead     bool      `xorm:"defense_is_read"`
	DestroyDurable    int       `xorm:"destroy_durable"`
	Occupy            int       `xorm:"occupy"`
	X                 int       `xorm:"x"`
	Y                 int       `xorm:"y"`
	CTime             time.Time `xorm:"ctime"`
}

func (this *WarReport) TableName() string {
	return "war_report"
}

/* 推送同步 begin */
func (this *WarReport) IsCellView() bool{
	return false
}

func (this *WarReport) BelongToRId() []int{
	return []int{this.AttackRid, this.DefenseRid}
}

func (this *WarReport) PushMsgName() string{
	return "warReport.push"
}

func (this *WarReport) Position() (int, int){
	return this.X, this.Y
}

func (this *WarReport) ToProto() interface{}{
	p := proto.WarReport{}
	p.CTime = this.CTime.UnixNano()/1e6
	p.Id = this.Id
	p.AttackRid = this.AttackRid
	p.DefenseRid = this.DefenseRid
	p.BegAttackArmy = this.BegAttackArmy
	p.BegDefenseArmy = this.BegDefenseArmy
	p.EndAttackArmy = this.EndAttackArmy
	p.EndDefenseArmy = this.EndDefenseArmy
	p.BegAttackGeneral = this.BegAttackGeneral
	p.BegDefenseGeneral = this.BegDefenseGeneral
	p.EndAttackGeneral = this.EndAttackGeneral
	p.EndDefenseGeneral = this.EndDefenseGeneral
	p.AttackIsWin = this.AttackIsWin
	p.AttackIsRead = this.AttackIsRead
	p.DefenseIsRead = this.DefenseIsRead
	p.DestroyDurable = this.DestroyDurable
	p.Occupy = this.Occupy
	p.X = this.X
	p.X = this.X
	return p
}

func (this *WarReport) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *WarReport) SyncExecute() {
	dbWarReportMgr.push(this)
	this.Push()
}
