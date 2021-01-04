package model

import (
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/net"
	"slgserver/server/slgserver/proto"
	"slgserver/server/slgserver/static_conf"
	"slgserver/util"
	"sync"
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

func (this*rcDBMgr) running()  {
	for true {
		select {
		case b := <- this.builds:
			if b.CityId >0 {
				_, err := db.MasterDB.Table(b).ID(b.CityId).Cols("cur_durable", "occupy_time").Update(b)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update role city build fail, because CityId <= 0")
			}
		}
	}
}

func (this*rcDBMgr) push(b *MapRoleCity)  {
	this.builds <- b
}
/*******db 操作end********/

type MapRoleCity struct {
	mutex		sync.Mutex	`xorm:"-"`
	CityId		int			`xorm:"cityId pk autoincr"`
	RId			int			`xorm:"rid"`
	Name		string		`xorm:"name" validate:"min=4,max=20,regexp=^[a-zA-Z0-9_]*$"`
	X			int			`xorm:"x"`
	Y			int			`xorm:"y"`
	IsMain		int8		`xorm:"is_main"`
	CurDurable	int			`xorm:"cur_durable"`
	CreatedAt	time.Time	`xorm:"created_at"`
	OccupyTime	time.Time 	`xorm:"occupy_time"`
}

func (this* MapRoleCity) IsWarFree() bool  {
	curTime := time.Now().Unix()
	if curTime - this.OccupyTime.Unix() < static_conf.Basic.Build.WarFree{
		return true
	}else{
		return false
	}
}

func (this*MapRoleCity) DurableChange(change int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	t := this.CurDurable + change
	if t < 0{
		this.CurDurable = 0
	}else{
		this.CurDurable = util.MinInt(GetMaxDurable(this.CityId), t)
	}
}

func (this *MapRoleCity) Level() int8 {
	return GetCityLv(this.CityId)
}

func (this *MapRoleCity) TableName() string {
	return "tb_map_role_city" + fmt.Sprintf("_%d", ServerId)
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
	p.UnionId = GetUnionId(this.RId)
	p.UnionName = GetUnionName(p.UnionId)
	p.ParentId = GetParentId(this.RId)
	p.MaxDurable = GetMaxDurable(this.RId)
	p.CurDurable = this.CurDurable
	p.Level = this.Level()
	p.RId = this.RId
	p.Name = this.Name
	p.IsMain = this.IsMain == 1
	p.OccupyTime = this.OccupyTime.UnixNano()/1e6
	return p
}

func (this *MapRoleCity) Push(){
	net.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *MapRoleCity) SyncExecute() {
	dbRCMgr.push(this)
	this.Push()
}