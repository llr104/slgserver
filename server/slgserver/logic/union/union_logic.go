package union

import (
	"sync"

	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/model"
	"go.uber.org/zap"
)

func GetUnionId(rid int) int {
	attr, ok := mgr.RAttrMgr.Get(rid)
	if ok {
		return attr.UnionId
	} else {
		return 0
	}
}

func GetUnionName(unionId int) string {
	if unionId <= 0 {
		return ""
	}

	u, ok := mgr.UnionMgr.Get(unionId)
	if ok {
		return u.Name
	} else {
		return ""
	}
}

func GetParentId(rid int) int {
	attr, ok := mgr.RAttrMgr.Get(rid)
	if ok {
		return attr.ParentId
	} else {
		return 0
	}
}

func GetMainMembers(unionId int) []int {
	u, ok := mgr.UnionMgr.Get(unionId)
	r := make([]int, 0)
	if ok {
		if u.Chairman != 0 {
			r = append(r, u.Chairman)
		}
		if u.ViceChairman != 0 {
			r = append(r, u.ViceChairman)
		}
	}
	return r
}

var _unionLogic *UnionLogic = nil

func Instance() *UnionLogic {
	if _unionLogic == nil {
		_unionLogic = newUnionLogic()
	}
	return _unionLogic
}

type UnionLogic struct {
	mutex    sync.RWMutex
	children map[int]map[int]int //key:unionId,key&value:child rid
}

func newUnionLogic() *UnionLogic {
	c := &UnionLogic{
		children: make(map[int]map[int]int),
	}
	c.init()
	return c
}

func (this *UnionLogic) init() {
	//初始化下属玩家
	attrs := mgr.RAttrMgr.List()
	for _, attr := range attrs {
		if attr.ParentId != 0 {
			this.PutChild(attr.ParentId, attr.RId)
		}
	}
}

func (this *UnionLogic) MemberEnter(rid, unionId int) {

	attr, ok := mgr.RAttrMgr.TryCreate(rid)
	if ok {
		attr.UnionId = unionId
		if attr.ParentId == unionId {
			this.DelChild(unionId, attr.RId)
		}
	} else {
		log.DefaultLog.Warn("EnterUnion not found roleAttribute", zap.Int("rid", rid))
	}

	if rcs, ok := mgr.RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.SyncExecute()
		}
	}
}

func (this *UnionLogic) MemberExit(rid int) {

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
func (this *UnionLogic) Dismiss(unionId int) {
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

func (this *UnionLogic) PutChild(unionId, rid int) {
	this.mutex.Lock()
	_, ok := this.children[unionId]
	if ok == false {
		this.children[unionId] = make(map[int]int)
	}
	this.children[unionId][rid] = rid
	this.mutex.Unlock()
}

func (this *UnionLogic) DelChild(unionId, rid int) {
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

func (this *UnionLogic) DelUnionAllChild(unionId int) {
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
