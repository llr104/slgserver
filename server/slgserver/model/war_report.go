package model

import (
	"fmt"
	"time"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"go.uber.org/zap"
)

type WarReport struct {
	Id                int       `xorm:"id pk autoincr"`
	AttackRid         int       `xorm:"a_rid"`
	DefenseRid        int       `xorm:"d_rid"`
	BegAttackArmy     string    `xorm:"b_a_army"`
	BegDefenseArmy    string    `xorm:"b_d_army"`
	EndAttackArmy     string    `xorm:"e_a_army"`
	EndDefenseArmy    string    `xorm:"e_d_army"`
	BegAttackGeneral  string    `xorm:"b_a_general"`
	BegDefenseGeneral string    `xorm:"b_d_general"`
	EndAttackGeneral  string    `xorm:"e_a_general"`
	EndDefenseGeneral string    `xorm:"e_d_general"`
	Result            int       `xorm:"result"` //0失败，1打平，2胜利
	Rounds            string    `xorm:"rounds"` //回合
	AttackIsRead      bool      `xorm:"a_is_read"`
	DefenseIsRead     bool      `xorm:"d_is_read"`
	DestroyDurable    int       `xorm:"destroy"`
	Occupy            int       `xorm:"occupy"`
	X                 int       `xorm:"x"`
	Y                 int       `xorm:"y"`
	CTime             time.Time `xorm:"ctime"`
}

func (this *WarReport) TableName() string {
	return "tb_war_report" + fmt.Sprintf("_%d", ServerId)
}

/* 推送同步 begin */
func (this *WarReport) IsCellView() bool {
	return false
}

func (this *WarReport) IsCanView(rid, x, y int) bool {
	return false
}

func (this *WarReport) BelongToRId() []int {
	return []int{this.AttackRid, this.DefenseRid}
}

func (this *WarReport) PushMsgName() string {
	return "warReport.push"
}

func (this *WarReport) Position() (int, int) {
	return this.X, this.Y
}

func (this *WarReport) TPosition() (int, int) {
	return -1, -1
}

func (this *WarReport) ToProto() interface{} {
	p := proto.WarReport{}
	p.CTime = int(this.CTime.UnixNano() / 1e6)
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
	p.Result = this.Result
	p.Rounds = this.Rounds
	p.AttackIsRead = this.AttackIsRead
	p.DefenseIsRead = this.DefenseIsRead
	p.DestroyDurable = this.DestroyDurable
	p.Occupy = this.Occupy
	p.X = this.X
	p.Y = this.Y
	return p
}

func (this *WarReport) Push() {
	net.ConnMgr.Push(this)
}

/* 推送同步 end */

func (this *WarReport) SyncExecute() {
	_, err := db.MasterDB.InsertOne(this)
	if err != nil {
		log.DefaultLog.Warn("db error", zap.Error(err))
	}
	this.Push()
}
