package logic

import (
	"slgserver/server/global"
	"slgserver/server/logic/mgr"
	"slgserver/util"
)

var ViewWidth = 5
var ViewHeight = 5

//是否在视野范围内
func armyIsInView(rid, x, y int) bool {

	attr, ok := mgr.RAttrMgr.Get(rid)
	if ok == false{
		return false
	}

	unionId := attr.UnionId
	for i := util.MaxInt(x-ViewWidth, 0); i < util.MinInt(x+ViewWidth, global.MapWith) ; i++ {
		for j := util.MaxInt(y-ViewHeight, 0); j < util.MinInt(y+ViewHeight, global.MapHeight) ; j++ {
			build, ok := mgr.RBMgr.PositionBuild(i, j)
			if ok {
				if (build.UnionId != 0 && unionId == build.UnionId) || build.RId == rid{
					return true
				}
			}

			city, ok := mgr.RCMgr.PositionCity(i, j)
			if ok {
				if (city.UnionId != 0 && unionId == city.UnionId) || city.RId == rid{
					return true
				}
			}
		}
	}

	return false
}