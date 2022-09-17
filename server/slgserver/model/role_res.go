package model

import (
	"fmt"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"go.uber.org/zap"
)

type Yield struct {
	Wood  int
	Iron  int
	Stone int
	Grain int
	Gold  int
}

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
					"grain", "gold", "decree").Update(res)
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

	y := GetYield(this.RId)
	p.GoldYield = y.Gold
	p.GrainYield = y.Grain
	p.StoneYield = y.Stone
	p.IronYield = y.Iron
	p.WoodYield = y.Wood
	p.DepotCapacity = GetDepotCapacity(this.RId)
	return p
}

func (this *RoleRes) Position() (int, int){
	return -1, -1
}

func (this *RoleRes) TPosition() (int, int){
	return -1, -1
}

func (this *RoleRes) Push(){
	net.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *RoleRes) SyncExecute() {
	dbRResMgr.push(this)
	this.Push()
}