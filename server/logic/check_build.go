package logic

import (
	"slgserver/server/global"
	"slgserver/server/logic/mgr"
	"slgserver/util"
)

func hasRoleBuildNearBy(x, y, rid, unionId int) bool {
	for i := util.MaxInt(x-1, 0); i <= util.MinInt(x+1, global.MapWith); i++ {
		for j := util.MaxInt(y-1, 0); j <= util.MinInt(y+1, global.MapHeight) ; j++ {
			if i == x && j == y {
				continue
			}
			if rb, ok := mgr.RBMgr.PositionBuild(i, j); ok {
				if rb.RId == rid || (unionId != 0 && rb.UnionId == unionId){
					return true
				}
			}
		}
	}
	return false
}

//是否能到达
func IsCanArrive(x, y, rid int) bool {
	unionId := mgr.RAttrMgr.UnionId(rid)

	//目标位置是城池
	if _, ok := mgr.RCMgr.PositionCity(x, y); ok {
		//城的四周是否有地相连
		//上
		ok := hasRoleBuildNearBy(x, y+2, rid, unionId)
		if ok {
			return ok
		}

		//下
		ok = hasRoleBuildNearBy(x, y-2, rid, unionId)
		if ok {
			return ok
		}

		//左
		ok = hasRoleBuildNearBy(x-2, y, rid, unionId)
		if ok {
			return ok
		}

		ok = hasRoleBuildNearBy(x+2, y, rid, unionId)
		if ok {
			return ok
		}
	}else{
		//普通领地
		ok := hasRoleBuildNearBy(x, y, rid, unionId)
		if ok {
			return true
		}

		//再判断是否和城市相连， 因为城池占了9格，所以该格子附近两个格子范围内有城池，则该地方是城池
		for i := x-2; i <= x+2; i++ {
			for j := y-2; j <= y+2; j++ {
				if rc, ok := mgr.RCMgr.PositionCity(i, j); ok {
					if rc.RId == rid || (unionId != 0 && rc.UnionId == unionId) {
						return true
					}
				}
			}
		}
	}

	return false
}

func IsCanDefend(x, y, rid int) bool{
	unionId := mgr.RAttrMgr.UnionId(rid)
	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok {
		if b.RId == rid{
			return true
		}else if b.UnionId > 0 {
			return b.UnionId == unionId
		}
	}

	c, ok := mgr.RCMgr.PositionCity(x, y)
	if ok {
		if c.RId == rid{
			return true
		}else if c.UnionId > 0 {
			return c.UnionId == unionId
		}
	}
	return false
}