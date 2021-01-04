package model

import (
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/net"
	"slgserver/server/slgserver/proto"
	"time"
)

/*******db 操作begin********/
var dbRBMgr *rbDBMgr
func init() {
	dbRBMgr = &rbDBMgr{builds: make(chan *MapRoleBuild, 100)}
	go dbRBMgr.running()
}

type rbDBMgr struct {
	builds   chan *MapRoleBuild
}

func (this *rbDBMgr) running()  {
	for true {
		select {
		case b := <- this.builds:
			if b.Id >0 {
				_, err := db.MasterDB.Table(b).ID(b.Id).Cols("rid",
					"cur_durable", "max_durable", "occupy_time", "giveUp_time").Update(b)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update role build fail, because id <= 0")
			}
		}
	}
}

func (this *rbDBMgr) push(b *MapRoleBuild)  {
	this.builds <- b
}
/*******db 操作end********/

type MapRoleBuild struct {
	Id    		int    		`xorm:"id pk autoincr"`
	RId   		int    		`xorm:"rid"`
	Type  		int8   		`xorm:"type"`
	Level		int8   		`xorm:"level"`
	X          	int       	`xorm:"x"`
	Y          	int       	`xorm:"y"`
	Name       	string    	`xorm:"name"`
	Wood       	int       	`xorm:"Wood"`
	Iron       	int       	`xorm:"iron"`
	Stone      	int       	`xorm:"stone"`
	Grain      	int       	`xorm:"grain"`
	CurDurable 	int       	`xorm:"cur_durable"`
	MaxDurable 	int       	`xorm:"max_durable"`
	Defender   	int       	`xorm:"defender"`
	OccupyTime 	time.Time 	`xorm:"occupy_time"`
	GiveUpTime 	int64 		`xorm:"giveUp_time"`
}

func (this *MapRoleBuild) TableName() string {
	return "tb_map_role_build" + fmt.Sprintf("_%d", ServerId)
}



/* 推送同步 begin */
func (this *MapRoleBuild) IsCellView() bool{
	return true
}

func (this *MapRoleBuild) IsCanView(rid, x, y int) bool{
	return true
}


func (this *MapRoleBuild) BelongToRId() []int{
	return []int{this.RId}
}

func (this *MapRoleBuild) PushMsgName() string{
	return "roleBuild.push"
}

func (this *MapRoleBuild) Position() (int, int){
	return this.X, this.Y
}

func (this *MapRoleBuild) TPosition() (int, int){
	return -1, -1
}

func (this *MapRoleBuild) ToProto() interface{}{

	p := proto.MapRoleBuild{}
	p.RNick = GetRoleNickName(this.RId)
	p.UnionId = GetUnionId(this.RId)
	p.UnionName = GetUnionName(p.UnionId)
	p.ParentId = GetParentId(this.RId)
	p.X = this.X
	p.Y = this.Y
	p.Type = this.Type
	p.CurDurable = this.CurDurable
	p.MaxDurable = this.MaxDurable
	p.Level = this.Level
	p.RId = this.RId
	p.Name = this.Name
	p.Defender = this.Defender
	p.OccupyTime = this.OccupyTime.UnixNano()/1e6
	p.GiveUpTime = this.GiveUpTime

	return p
}

func (this *MapRoleBuild) Push(){
	net.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *MapRoleBuild) SyncExecute() {
	dbRBMgr.push(this)
	this.Push()
}