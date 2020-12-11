package model

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/conn"
	"slgserver/server/proto"
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
					"cur_durable", "max_durable").Update(b)
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
	RNick		string		`xorm:"-"`
	UnionId		int			`xorm:"-"`	//联盟id
	Type  		int8   		`xorm:"type"`
	Level 		int8   		`xorm:"level"`
	X     		int    		`xorm:"x"`
	Y     		int    		`xorm:"y"`
	Name  		string 		`xorm:"name"`
	Wood  		int    		`xorm:"Wood"`
	Iron  		int    		`xorm:"iron"`
	Stone 		int    		`xorm:"stone"`
	Grain		int			`xorm:"grain"`
	CurDurable	int			`xorm:"cur_durable"`
	MaxDurable	int			`xorm:"max_durable"`
	Defender	int			`xorm:"defender"`
}

func (this *MapRoleBuild) TableName() string {
	return "map_role_build"
}


/* 推送同步 begin */
func (this *MapRoleBuild) IsCellView() bool{
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
	p.RNick = this.RNick
	p.UnionId = this.UnionId
	p.X = this.X
	p.Y = this.Y
	p.Type = this.Type
	p.CurDurable = this.CurDurable
	p.MaxDurable = this.MaxDurable
	p.Level = this.Level
	p.RId = this.RId
	p.Name = this.Name
	p.Defender = this.Defender
	return p
}

func (this *MapRoleBuild) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *MapRoleBuild) SyncExecute() {
	dbRBMgr.push(this)
	this.Push()
}