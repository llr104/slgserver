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
var dbRCMgr *rcDBMgr
func init() {
	dbRCMgr = &rcDBMgr{builds: make(chan *MapRoleCity, 100)}
	go dbRCMgr.running()
}

type rcDBMgr struct {
	builds   chan *MapRoleCity
}

func (this* rcDBMgr) running()  {
	for true {
		select {
		case b := <- this.builds:
			if b.CityId >0 {
				_, err := db.MasterDB.Table(b).ID(b.CityId).Cols("level",
					"cur_durable", "max_durable", "cost").Update(b)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update role city build fail, because CityId <= 0")
			}
		}
	}
}

func (this* rcDBMgr) push(b *MapRoleCity)  {
	this.builds <- b
}
/*******db 操作end********/

type MapRoleCity struct {
	CityId		int			`xorm:"cityId pk autoincr"`
	RId			int			`xorm:"rid"`
	Name		string		`xorm:"name" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	UnionId		int			`xorm:"-"`	//联盟id
	X			int			`xorm:"x"`
	Y			int			`xorm:"y"`
	IsMain		int8		`xorm:"is_main"`
	Level		int8		`xorm:"level"`
	CurDurable	int			`xorm:"cur_durable"`
	MaxDurable	int			`xorm:"max_durable"`
	Cost       	int8   		`xorm:"cost"`
	CreatedAt	time.Time	`xorm:"created_at"`
}

func (this *MapRoleCity) TableName() string {
	return "map_role_city"
}

/* 推送同步 begin */
func (this *MapRoleCity) IsCellView() bool{
	return true
}

func (this *MapRoleCity) IsCanView(rid, x, y int) bool{
	return true
}

func (this *MapRoleCity) BelongToRId() []int{
	return []int{this.RId}
}

func (this *MapRoleCity) PushMsgName() string{
	return "roleCity.push"
}

func (this *MapRoleCity) Position() (int, int){
	return this.X, this.Y
}

func (this *MapRoleCity) TPosition() (int, int){
	return -1, -1
}

func (this *MapRoleCity) ToProto() interface{}{
	p := proto.MapRoleCity{}
	p.X = this.X
	p.Y = this.Y
	p.CityId = this.CityId
	p.UnionId = this.UnionId
	p.CurDurable = this.CurDurable
	p.MaxDurable = this.MaxDurable
	p.Level = this.Level
	p.RId = this.RId
	p.Name = this.Name
	p.IsMain = this.IsMain == 1
	p.Cost = this.Cost
	return p
}

func (this *MapRoleCity) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *MapRoleCity) SyncExecute() {
	dbRCMgr.push(this)
	this.Push()
}