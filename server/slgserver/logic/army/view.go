package army

import (
	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/logic/union"
	"github.com/llr104/slgserver/util"
)

var ViewWidth = 5
var ViewHeight = 5

//是否在视野范围内
func ArmyIsInView(rid, x, y int) bool {
	unionId := union.GetUnionId(rid)
	for i := util.MaxInt(x-ViewWidth, 0); i < util.MinInt(x+ViewWidth, global.MapWith); i++ {
		for j := util.MaxInt(y-ViewHeight, 0); j < util.MinInt(y+ViewHeight, global.MapHeight); j++ {
			build, ok := mgr.RBMgr.PositionBuild(i, j)
			if ok {
				tUnionId := union.GetUnionId(build.RId)
				if (tUnionId != 0 && unionId == tUnionId) || build.RId == rid {
					return true
				}
			}

			city, ok := mgr.RCMgr.PositionCity(i, j)
			if ok {
				tUnionId := union.GetUnionId(city.RId)
				if (tUnionId != 0 && unionId == tUnionId) || city.RId == rid {
					return true
				}
			}
		}
	}

	return false
}
