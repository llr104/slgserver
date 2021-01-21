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
	"time"
)

const (
	MapBuildFortress = 50		//玩家要塞
	MapBuildSysFortress = 56	//系统要塞
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
				_, err := db.MasterDB.Table(b).ID(b.Id).Cols(
					"rid", "name", "type", "level", "op_level",
					"cur_durable", "max_durable", "occupy_time",
					"giveUp_time", "end_time").Update(b)
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
	OPLevel		int8		`xorm:"op_level"`	//操作level
	X          	int       	`xorm:"x"`
	Y          	int       	`xorm:"y"`
	Name       	string    	`xorm:"name"`
	Wood       	int       	`xorm:"-"`
	Iron       	int       	`xorm:"-"`
	Stone      	int       	`xorm:"-"`
	Grain      	int       	`xorm:"-"`
	Defender   	int       	`xorm:"-"`
	CurDurable 	int       	`xorm:"cur_durable"`
	MaxDurable 	int       	`xorm:"max_durable"`
	OccupyTime 	time.Time 	`xorm:"occupy_time"`
	EndTime 	time.Time 	`xorm:"end_time"`	//建造或升级完的时间
	GiveUpTime 	int64 		`xorm:"giveUp_time"`
}

func (this *MapRoleBuild) TableName() string {
	return "tb_map_role_build" + fmt.Sprintf("_%d", ServerId)
}

func (this* MapRoleBuild) Init() {
	if cfg, _ := static_conf.MapBuildConf.BuildConfig(this.Type, this.Level); cfg != nil {
		this.Name = cfg.Name
		this.Level = cfg.Level
		this.Type = cfg.Type
		this.Wood = cfg.Wood
		this.Iron = cfg.Iron
		this.Stone = cfg.Stone
		this.Grain = cfg.Grain
		this.MaxDurable = cfg.Durable
		this.Defender = cfg.Defender
	}
}


func (this* MapRoleBuild) Reset() {
	ok, t, level := MapResTypeLevel(this.X, this.Y)
	if ok {
		if cfg, _ := static_conf.MapBuildConf.BuildConfig(t, level); cfg != nil {
			this.Name = cfg.Name
			this.Level = cfg.Level
			this.Type = cfg.Type
			this.Wood = cfg.Wood
			this.Iron = cfg.Iron
			this.Stone = cfg.Stone
			this.Grain = cfg.Grain
			this.MaxDurable = cfg.Durable
			this.Defender = cfg.Defender
		}
	}

	this.GiveUpTime = 0
	this.RId = 0
	this.EndTime = time.Time{}
	this.OPLevel = this.Level
	this.CurDurable = util.MinInt(this.MaxDurable, this.CurDurable)
}

func (this* MapRoleBuild) ConvertToRes() {
	rid := this.RId
	giveUp := this.GiveUpTime
	this.Reset()
	this.RId = rid
	this.GiveUpTime = giveUp
}

func (this* MapRoleBuild) IsInGiveUp() bool {
	return this.GiveUpTime != 0
}

func (this* MapRoleBuild) IsWarFree() bool  {
	curTime := time.Now().Unix()
	if curTime - this.OccupyTime.Unix() < static_conf.Basic.Build.WarFree{
		return true
	}else{
		return false
	}
}

func (this* MapRoleBuild) IsResBuild() bool  {
	return this.Grain > 0 || this.Stone > 0 || this.Iron > 0 || this.Wood > 0
}

//是否有修改等级权限
func (this* MapRoleBuild) IsHaveModifyLVAuth() bool  {
	return this.Type == MapBuildFortress
}

func (this* MapRoleBuild) IsBusy() bool{
	if this.Level != this.OPLevel{
		return true
	}else {
		return false
	}
}

func (this* MapRoleBuild) IsRoleFortress() bool  {
	return this.Type == MapBuildFortress
}

//是否有调兵权限
func (this* MapRoleBuild) IsHasTransferAuth() bool  {
	return this.Type == MapBuildFortress || this.Type == MapBuildSysFortress
}

func (this* MapRoleBuild) BuildOrUp(cfg static_conf.BCLevelCfg) {
	this.Type = cfg.Type
	this.Level = cfg.Level -1
	this.Name = cfg.Name
	this.OPLevel = cfg.Level
	this.GiveUpTime = 0

	this.Wood = 0
	this.Iron = 0
	this.Stone = 0
	this.Grain = 0
	this.EndTime = time.Now().Add(time.Duration(cfg.Time) * time.Second)
}

func (this* MapRoleBuild) DelBuild(cfg static_conf.BCLevelCfg) {
	this.OPLevel = 0
	this.EndTime = time.Now().Add(time.Duration(cfg.Time) * time.Second)
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
	p.RId = this.RId
	p.Name = this.Name

	p.OccupyTime = this.OccupyTime.UnixNano()/1e6
	p.GiveUpTime = this.GiveUpTime*1000
	p.EndTime = this.EndTime.UnixNano()/1e6

	if this.EndTime.IsZero() == false{
		if this.IsHasTransferAuth(){
			if time.Now().Before(this.EndTime) == false{

				if this.OPLevel == 0{
					this.ConvertToRes()
				}else{
					this.Level = this.OPLevel
					this.EndTime = time.Time{}
					cfg, ok := static_conf.MapBCConf.BuildConfig(this.Type, this.Level)
					if ok {
						this.MaxDurable = cfg.Durable
						this.CurDurable = util.MinInt(this.MaxDurable, this.CurDurable)
						this.Defender = cfg.Defender
					}
				}
			}
		}
	}

	p.CurDurable = this.CurDurable
	p.MaxDurable = this.MaxDurable
	p.Defender = this.Defender
	p.Level = this.Level
	p.OPLevel = this.OPLevel
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