package logic

import "slgserver/server/logic/mgr"


type coalitionLogic struct {

}

func getUnionId(rid int) int {
	attr, ok := mgr.RAttrMgr.Get(rid)
	if ok {
		return attr.UnionId
	}else{
		return 0
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

func (this* coalitionLogic) MemberEnter(rid, unionId int)  {
	mgr.RAttrMgr.EnterUnion(rid, unionId)

	if ra, ok := mgr.RAttrMgr.Get(rid); ok {
		ra.UnionId = unionId
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
