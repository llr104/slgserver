package model

import (
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/slgserver/conn"
)

/*******db 操作begin********/
var dbRAttrMgr *roleAttrDBMgr
func init() {
	dbRAttrMgr = &roleAttrDBMgr{rattr: make(chan *RoleAttribute, 100)}
	go dbRAttrMgr.running()
}

type roleAttrDBMgr struct {
	rattr chan *RoleAttribute
}

func (this *roleAttrDBMgr) running()  {
	for true {
		select {
		case attr := <- this.rattr:
			if attr.Id >0 {
				_, err := db.MasterDB.Table(attr).ID(attr.Id).Cols(
					"parent_id").Update(attr)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
			}else{
				log.DefaultLog.Warn("update role attr fail, because id <= 0")
			}
		}
	}
}

func (this *roleAttrDBMgr) push(attr *RoleAttribute)  {
	this.rattr <- attr
}
/*******db 操作end********/

type RoleAttribute struct {
	Id			int		`xorm:"id pk autoincr"`
	RId			int		`xorm:"rid"`
	UnionId 	int		`xorm:"-"`			//联盟id
	ParentId	int		`xorm:"parent_id"`	//上级id（被沦陷）
}


func (this *RoleAttribute) TableName() string {
	return "tb_role_attribute" + fmt.Sprintf("_%d", ServerId)
}


/* 推送同步 begin */
func (this *RoleAttribute) IsCellView() bool{
	return false
}

func (this *RoleAttribute) IsCanView(rid, x, y int) bool{
	return false
}

func (this *RoleAttribute) BelongToRId() []int{
	return []int{this.RId}
}

func (this *RoleAttribute) PushMsgName() string{
	return "roleAttr.push"
}

func (this *RoleAttribute) ToProto() interface{}{
	return nil
}

func (this *RoleAttribute) Position() (int, int){
	return -1, -1
}

func (this *RoleAttribute) TPosition() (int, int){
	return -1, -1
}

func (this *RoleAttribute) Push(){
	conn.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *RoleAttribute) SyncExecute() {
	dbRAttrMgr.push(this)
}