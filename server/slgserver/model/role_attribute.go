package model

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/net"
	"slgserver/server/slgserver/proto"
	"time"
	"xorm.io/xorm"
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
					"parent_id", "collect_times", "last_collect_time", "pos_tag").Update(attr)
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
	Id              int       		`xorm:"id pk autoincr"`
	RId             int       		`xorm:"rid"`
	UnionId         int       		`xorm:"-"`					//联盟id
	ParentId        int       		`xorm:"parent_id"`			//上级id（被沦陷）
	CollectTimes    int8      		`xorm:"collect_times"`		//征收次数
	LastCollectTime time.Time 		`xorm:"last_collect_time"`	//最后征收的时间
	PosTags			string    		`xorm:"pos_tags"`			//位置标记
	PosTagArray		[]proto.PosTag	`xorm:"-"`
}


func (this *RoleAttribute) TableName() string {
	return "tb_role_attribute" + fmt.Sprintf("_%d", ServerId)
}

func (this *RoleAttribute) AfterSet(name string, cell xorm.Cell){
	if name == "pos_tags"{
		this.PosTagArray = make([]proto.PosTag,0)
		if cell != nil{
			data, ok := (*cell).([]uint8)
			if ok {
				json.Unmarshal(data, &this.PosTagArray)
				fmt.Println(this.PosTagArray)
			}
		}
	}
}

func (this *RoleAttribute) beforeModify()  {
	data, _ := json.Marshal(this.PosTagArray)
	this.PosTags = string(data)
}

func (this *RoleAttribute) BeforeInsert() {
	this.beforeModify()
}

func (this *RoleAttribute) BeforeUpdate() {
	this.beforeModify()
}

func (this* RoleAttribute) RemovePosTag(x, y int) {
	tags := make([]proto.PosTag, 0)
	for _, tag := range this.PosTagArray {
		if tag.X != x || tag.Y != y{
			tags = append(tags, tag)
		}
	}
	this.PosTagArray = tags
}

func (this* RoleAttribute) AddPosTag(x, y int, name string) {
	ok := true
	for _, tag := range this.PosTagArray {
		if tag.X == x && tag.Y == y{
			ok = false
			break
		}
	}
	if ok{
		this.PosTagArray = append(this.PosTagArray, proto.PosTag{X: x, Y: y, Name: name})
	}
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
	net.ConnMgr.Push(this)
}
/* 推送同步 end */

func (this *RoleAttribute) SyncExecute() {
	dbRAttrMgr.push(this)
}