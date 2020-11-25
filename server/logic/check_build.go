package logic

//是否能到达
func IsCanArrive(x, y, rid int) bool {
	for i := x-1; i <= x+1; i++ {
		if i < 0 || i >= MapWith{
			continue
		}
		for j := y-1; j <=y+1 ; j++ {
			if j < 0 || j >= MapHeight {
				continue
			}
			if i == x && j == y {
				continue
			}

			if rc, ok := RCMgr.PositionCity(i, j); ok {
				if rc.RId == rid{
					return true
				}
			}

			if rb, ok := RBMgr.PositionBuild(i, j); ok {
				if rb.RId == rid{
					return true
				}
			}
		}
	}
	return false
}