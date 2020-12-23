package model

import (
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/conn"
	"slgserver/server/slgserver/proto"
)

/*******db 操作begin********/
var dbRResMgr *roleResDBMgr
func init() {
	dbRResMgr = &roleResDBMgr{ress: make(chan *RoleRes, 100)}
	go dbRResMgr.running()
}

type roleResDBMgr struct {
	ress   chan *RoleRes
}

func (this *roleResDBMgr) running()  {
	for true {
		select {
		case res := <- this.ress:
			if res.Id >0 {
				_, err := db.MasterDB.Table(res).ID(res.Id).Cols(
					"wood", "iron", "stone",
					"grain", "gold", "decree", "wood_yield",
					"iron_yield", "stone_yield", "gold_yield",
					"gold_yield", "depot_capacity").Update(res)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update role build fail, because id <= 0")
			}
		}
	}
}

func (this *roleResDBMgr) push(res *RoleRes)  {
	this.ress <- res
}
/*******db 操作end********/


type RoleRes struct {
	Id        		int    		`xorm:"id pk autoincr"`
	RId       		int    		`xorm:"rid"`
	Wood      		int    		`xorm:"wood"`
	Iron      		int    		`xorm:"iron"`
	Stone     		int    		`xorm:"stone"`
	Grain     		int    		`xorm:"grain"`
	Gold      		int    		`xorm:"gold"`
	Decree    		int    		`xorm:"decree"`	//令牌
	WoodYield 		int    		`xorm:"wood_yield"`
	IronYield 		int    		`xorm:"iron_yield"`
	StoneYield		int			`xorm:"stone_yield"`
	GrainYield		int			`xorm:"grain_yield"`
	GoldYield		int			`xorm:"gold_yield"`
	DepotCapacity	int			`xorm:"depot_capacity"`	//仓库容量
}

func (this *RoleRes) TableName() string {
	return "tb_role_res" + fmt.Sprintf("_%d", ServerId)
}


/* 推送同步 begin */
func (this *RoleRes) IsCellView() bool{
	return false
}

func (this *RoleRes) IsCanView(rid, x, y int) bool{
	return false
}

func (this *RoleRes) BelongToRId() []int{
	return []int{this.RId}
}

func (this *RoleRes) PushMsgName() string{
	return "roleRes.push"
}

func (this *RoleRes) ToProto() interface{}{
	p := proto.RoleRes{}
	p.Gold = this.Gold
	p.Grain = this.Grain
	p.Stone = this.Stone
	p.Iron = this.Iron
	p.Wood = this.Wood
	p.Decree = this.Decree
	p.GoldYield = this.GoldYield
	p.GrainYield = this.GrainYield
	p.StoneYield = this.StoneYield
	p.IronYield = this.IronYield
	p.WoodYield = this.WoodYield
	p.DepotCapacity = this.DepotCapacity
	return p
}

func (this *RoleRes) Position() (int, int){
	return -1, -1
}

func (this *RoleRes) TPosition() (int, int){
	return -1, -1
}

func (this *RoleRes) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *RoleRes) SyncExecute() {
	dbRResMgr.push(this)
	this.Push()
}