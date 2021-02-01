package logic

import (
	"slgserver/server/slgserver/global"
	"slgserver/server/slgserver/logic/mgr"
	"slgserver/util"
)

//是否和玩家领地相连
func hasNearByRoleBuild(x, y, rid, unionId int) bool {
	if _, ok := mgr.RCMgr.PositionCity(x, y); ok {
		for i := util.MaxInt(x-2, 0); i <= util.MinInt(x+2, global.MapWith); i++ {
			for j := util.MaxInt(y-2, 0); j <= util.MinInt(y+2, global.MapHeight) ; j++ {
				if i == x && j == y {
					continue
				}
				if rb, ok := mgr.RBMgr.PositionBuild(i, j); ok {
					tUnionId := getUnionId(rb.RId)
					if rb.RId == rid || (unionId != 0 && tUnionId == unionId){
						return true
					}

					tParentId := getParentId(rb.RId)
					if tParentId != 0 && tParentId == unionId{
						return true
					}
				}
			}
		}
	}else{
		for i := util.MaxInt(x-1, 0); i <= util.MinInt(x+1, global.MapWith); i++ {
			for j := util.MaxInt(y-1, 0); j <= util.MinInt(y+1, global.MapHeight) ; j++ {
				if i == x && j == y {
					continue
				}

				if rb, ok := mgr.RBMgr.PositionBuild(i, j); ok {
					tUnionId := getUnionId(rb.RId)
					if rb.RId == rid || (unionId != 0 && tUnionId == unionId){
						return true
					}

					tParentId := getParentId(rb.RId)
					if tParentId != 0 && tParentId == unionId{
						return true
					}
				}
			}
		}
	}
	return false
}

func hasNearByRoleCity(x, y, rid, unionId int) bool {

	for i := x-2; i <= x+2; i++ {
		for j := y-2; j <= y+2; j++ {
			if rc, ok := mgr.RCMgr.PositionCity(i, j); ok {
				tUnionId := getUnionId(rc.RId)
				if rc.RId == rid || (unionId != 0 && tUnionId == unionId) {
					return true
				}

				tParentId := getParentId(rc.RId)
				if tParentId != 0 && tParentId == unionId {
					return true
				}
			}
		}
	}
	return false
}


//是否能到达
func IsCanArrive(x, y, rid int) bool {

	unionId := getUnionId(rid)
	//是否和玩家领地相连
	ok := hasNearByRoleBuild(x, y, rid, unionId)
	if ok {
		return true
	}

	//是否和玩家城池相连
	ok = hasNearByRoleCity(x, y, rid, unionId)
	if ok {
		return true
	}

	//再判断是否是系统城池相连
	if rb, ok := mgr.RBMgr.PositionBuild(x, y); ok {
		if rb.IsSysCity() {
			for tx := x-rb.CellRadius(); tx <= x+rb.CellRadius(); tx++ {
				for ty := y-rb.CellRadius(); ty <= y+rb.CellRadius(); ty++ {
					ok = hasNearByRoleBuild(tx, ty, rid, unionId)
					if ok {
						return true
					}
				}
			}
		}
	}

	return false
}

func IsCanDefend(x, y, rid int) bool{
	unionId := getUnionId(rid)

	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok {
		tUnionId := getUnionId(b.RId)
		tParentId := getParentId(b.RId)
		if b.RId == rid{
			return true
		}else if tUnionId > 0 {
			return tUnionId == unionId
		}else if tParentId > 0 {
			return tParentId == unionId
		}
	}

	c, ok := mgr.RCMgr.PositionCity(x, y)
	if ok {
		tUnionId := getUnionId(c.RId)
		tParentId := getParentId(c.RId)
		if c.RId == rid{
			return true
		}else if tUnionId > 0 {
			return tUnionId == unionId
		}else if tParentId > 0 {
			return tParentId == unionId
		}
	}
	return false
}

//是否是免战
func IsWarFree(x, y int) bool{
	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok {
		return b.IsWarFree()
	}

	c, ok := mgr.RCMgr.PositionCity(x, y)
	if ok && getParentId(c.RId) > 0 {
		return c.IsWarFree()
	}
	return false
}