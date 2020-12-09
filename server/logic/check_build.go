package logic

import "slgserver/server/global"

//是否能到达
func IsCanArrive(x, y, rid int) bool {
	unionId := RAttributeMgr.UnionId(rid)

	//先判断上下左右是否有普通地相连
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

			if rb, ok := RBMgr.PositionBuild(i, j); ok {
				if rb.RId == rid || (unionId != 0 && rb.UnionId == unionId){
					return true
				}
			}
		}
	}

	//再判断是否和城市相连， 因为城池占了9格，所以该格子附近两个格子范围内有城池，则该地方是城池
	for i := x-2; i <= x+2; i++ {
		for j := y-2; j <= y+2; j++ {
			if rc, ok := RCMgr.PositionCity(i, j); ok {
				if rc.RId == rid || (unionId != 0 && rc.UnionId == unionId) {
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

	c, ok := RCMgr.PositionCity(x, y)
	if ok {
		if c.RId == rid{
			return true
		}else if c.UnionId > 0 {
			return c.UnionId == unionId
		}
	}
	return false
}