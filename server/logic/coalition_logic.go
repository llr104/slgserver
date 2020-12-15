package logic

import "slgserver/server/logic/mgr"


type coalitionLogic struct {

}

func (this* coalitionLogic) MemberEnter(rid, unionId int)  {
	mgr.RAttrMgr.EnterUnion(rid, unionId)

	if ra, ok := mgr.RAttrMgr.Get(rid); ok {
		ra.UnionId = unionId
	}

	if rbs, ok := mgr.RBMgr.GetRoleBuild(rid); ok {
		for _, rb := range rbs {
			rb.UnionId = unionId
		}
	}

	if rcs, ok := mgr.RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.UnionId = unionId
			rc.SyncExecute()
		}
	}

	if armys, ok := mgr.AMgr.GetByRId(rid); ok {
		for _, army := range armys {
			army.UnionId = unionId
		}
	}
}

func (this* coalitionLogic) MemberExit(rid int) {

	if ra, ok := mgr.RAttrMgr.Get(rid); ok {
		ra.UnionId = 0
	}

	if rbs, ok := mgr.RBMgr.GetRoleBuild(rid); ok {
		for _, rb := range rbs {
			rb.UnionId = 0
		}
	}

	if rcs, ok := mgr.RCMgr.GetByRId(rid); ok {
		for _, rc := range rcs {
			rc.UnionId = 0
			rc.SyncExecute()
		}
	}

	if armys, ok := mgr.AMgr.GetByRId(rid); ok {
		for _, army := range armys {
			army.UnionId = 0
		}
	}

}
