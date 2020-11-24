package logic

import (
	"slgserver/model"
	"sync"
)

var SysArmy* sysArmyLogic

func init() {
	SysArmy = &sysArmyLogic{
		sysArmys: make(map[int][]*model.Army),
	}
}

type sysArmyLogic struct {
	mutex 		sync.Mutex
	sysArmys    map[int][]*model.Army   //key:posId 系统建筑军队
}

func (this * sysArmyLogic) GetArmy(x, y int) []*model.Army{
	this.mutex.Lock()
	defer this.mutex.Unlock()
	posId := ToPosition(x, y)
	a,ok := this.sysArmys[posId]
	if ok {
		return a
	}else{
		gs, ok := GMgr.GetNPCGenerals(3)
		gsId := make([]int, 0)

		for _, v := range gs {
			gsId = append(gsId, v.Id)
		}

		armys := make([]*model.Army, 0)
		if ok {
			if cfg, ok := NMMgr.PositionBuild(x, y); ok{
				n := 100*int(cfg.Level)
				scnt := []int{n, n, n}
				army := &model.Army{RId: 0, Order: 0, CityId: 0,
					GeneralArray: gsId, SoldierArray: scnt}
				army.ToGeneral()
				army.ToSoldier()

				armys = append(armys, army)
				return armys
			}else{
				return armys
			}
		}else{
			return armys
		}
	}
}


