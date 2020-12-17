package logic

import (
	"go.uber.org/zap"
	"slgserver/log"
	"slgserver/server/logic/mgr"
	"slgserver/server/model"
	"sync"
)

func getUnionId(rid int) int {
	attr, ok := mgr.RAttrMgr.Get(rid)
	if ok {
		return attr.UnionId
	}else{
		return 0
	}
}

func getUnionName(unionId int) string {
	if unionId <= 0{
		return ""
	}

	u, ok := mgr.UnionMgr.Get(unionId)
	if ok {
		return u.Name
	}else{
		return ""
	}
}

func getParentId(rid int) int {
	attr, ok := mgr.RAttrMgr.Get(rid)
	if ok {
		return attr.ParentId
	}else{
		return 0
	}
}

func getMainMembers(unionId int) []int {
	u, ok := mgr.UnionMgr.Get(unionId)
	r := make([]int, 0)
	if ok {
		if u.Chairman != 0{
			r = append(r, u.Chairman)
		}
		if u.ViceChairman != 0{
			r = append(r, u.ViceChairman)
		}
	}
	return r
}


type coalitionLogic struct {
	mutex sync.RWMutex
	children map[int]map[int]int	//key:unionId,key&value:child rid
}

func NewCoalitionLogic() *coalitionLogic {
	c := &coalitionLogic{
		children: make(map[int]map[int]int),
	}
	c.init()
	return c
}

func (this* coalitionLogic) init()  {
	//初始化下属玩家
	attrs := mgr.RAttrMgr.List()
	for _, attr := range attrs {
		if attr.ParentId !=0 {
			this.PutChild(attr.ParentId, attr.RId)
		}
	}
}

func (this* coalitionLogic) MemberEnter(rid, unionId int)  {

	attr, ok := mgr.RAttrMgr.TryCreate(rid)
	if ok {
		attr.UnionId = unionId
		if attr.ParentId == unionId{
			this.DelChild(unionId, attr.RId)
		}
	}else{
		log.DefaultLog.Warn("EnterUnion not found roleAttribute", zap.Int("rid", rid))
	}

	if rcs, ok := mgr.RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.SyncExecute()
		}
	}
}

func (this* coalitionLogic) MemberExit(rid int) {

	if ra, ok := mgr.RAttrMgr.Get(rid); ok {
		ra.UnionId = 0
	}

	if rcs, ok := mgr.RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.SyncExecute()
		}
	}
}

//解散
func (this* coalitionLogic) Dismiss(unionId int) {
	u, ok := mgr.UnionMgr.Get(unionId)
	if ok {
		mgr.UnionMgr.Remove(unionId)
		for _, rid := range u.MemberArray {
			this.MemberExit(rid)
			this.DelUnionAllChild(unionId)
		}
		u.State = model.UnionDismiss
		u.MemberArray = []int{}
		u.SyncExecute()
	}
}

func (this* coalitionLogic) PutChild(unionId, rid int) {
	this.mutex.Lock()
	_, ok := this.children[unionId]
	if ok == false {
		this.children[unionId] = make(map[int]int)
	}
	this.children[unionId][rid] = rid
	this.mutex.Unlock()
}

func (this* coalitionLogic) DelChild(unionId, rid int) {
	this.mutex.Lock()
	children, ok := this.children[unionId]
	if ok {
		attr, ok := mgr.RAttrMgr.Get(rid)
		if ok {
			attr.ParentId = 0
			attr.SyncExecute()
		}
		delete(children, rid)
	}
	this.mutex.Unlock()
}

func (this* coalitionLogic) DelUnionAllChild(unionId int) {
	this.mutex.Lock()
	children, ok := this.children[unionId]
	if ok {
		for _, child := range children {
			attr, ok := mgr.RAttrMgr.Get(child)
			if ok {
				attr.ParentId = 0
				attr.SyncExecute()
			}

			city, ok := mgr.RCMgr.GetMainCity(child)
			if ok {
				city.SyncExecute()
			}
		}
		delete(this.children, unionId)
	}
	this.mutex.Unlock()
}
