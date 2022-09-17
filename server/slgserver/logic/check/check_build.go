package check

import (
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/logic/union"
	"github.com/llr104/slgserver/util"
)

//是否能到达
func IsCanArrive(x, y, rid int) bool {
	var radius = 0
	unionId := union.GetUnionId(rid)
	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok {
		radius = b.CellRadius()
	}

	c, ok := mgr.RCMgr.PositionCity(x, y)
	if ok {
		radius = c.CellRadius()
	}

	//查找10格半径
	for tx := x - 10; tx <= x+10; tx++ {
		for ty := y - 10; ty <= y+10; ty++ {
			b1, ok := mgr.RBMgr.PositionBuild(tx, ty)
			if ok {
				absX := util.AbsInt(x - tx)
				absY := util.AbsInt(y - ty)
				if absX <= radius+b1.CellRadius()+1 && absY <= radius+b1.CellRadius()+1 {
					unionId1 := union.GetUnionId(b1.RId)
					parentId := union.GetParentId(b1.RId)
					if b1.RId == rid || (unionId != 0 && (unionId == unionId1 || unionId == parentId)) {
						return true
					}
				}
			}

			c1, ok := mgr.RCMgr.PositionCity(tx, ty)
			if ok {
				absX := util.AbsInt(x - tx)
				absY := util.AbsInt(y - ty)
				if absX <= radius+c1.CellRadius()+1 && absY <= radius+c1.CellRadius()+1 {
					unionId1 := union.GetUnionId(c1.RId)
					parentId := union.GetParentId(c1.RId)
					if c1.RId == rid || (unionId != 0 && (unionId == unionId1 || unionId == parentId)) {
						return true
					}
				}
			}
		}
	}

	return false
}

func IsCanDefend(x, y, rid int) bool {
	unionId := union.GetUnionId(rid)

	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok {
		tUnionId := union.GetUnionId(b.RId)
		tParentId := union.GetParentId(b.RId)
		if b.RId == rid {
			return true
		} else if tUnionId > 0 {
			return tUnionId == unionId
		} else if tParentId > 0 {
			return tParentId == unionId
		}
	}

	c, ok := mgr.RCMgr.PositionCity(x, y)
	if ok {
		tUnionId := union.GetUnionId(c.RId)
		tParentId := union.GetParentId(c.RId)
		if c.RId == rid {
			return true
		} else if tUnionId > 0 {
			return tUnionId == unionId
		} else if tParentId > 0 {
			return tParentId == unionId
		}
	}
	return false
}

//是否是免战
func IsWarFree(x, y int) bool {
	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok {
		return b.IsWarFree()
	}

	c, ok := mgr.RCMgr.PositionCity(x, y)
	if ok && union.GetParentId(c.RId) > 0 {
		return c.IsWarFree()
	}
	return false
}
