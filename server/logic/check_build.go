package logic

import "slgserver/server/global"

//是否能到达
func IsCanArrive(x, y, rid int) bool {
	unionId := RAttributeMgr.UnionId(rid)
	for i := x-1; i <= x+1; i++ {
		if i < 0 || i >= global.MapWith{
			continue
		}
		for j := y-1; j <=y+1 ; j++ {
			if j < 0 || j >= global.MapHeight {
				continue
			}
			if i == x && j == y {
				continue
			}

			if rc, ok := RCMgr.PositionCity(i, j); ok {
				if rc.RId == rid || (unionId != 0 && rc.UnionId == unionId) {
					return true
				}
			}

			if rb, ok := RBMgr.PositionBuild(i, j); ok {
				if rb.RId == rid || (unionId != 0 && rb.UnionId == unionId){
					return true
				}
			}
		}
	}
	return false
}

func IsCanDefend(x, y, rid int) bool{
	unionId := RAttributeMgr.UnionId(rid)
	b, ok := RBMgr.PositionBuild(x, y)
	if ok {
		if b.RId == rid{
			return true
		}else if b.UnionId > 0 {
			return b.UnionId == unionId
		}
	}
	return false
}