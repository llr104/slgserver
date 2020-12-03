package logic

import (
	"slgserver/server/model"
	"slgserver/server/static_conf"
	"sync"
)


func NewSysArmy() *sysArmyLogic {
	return &sysArmyLogic{
		sysArmys: make(map[int][]*model.Army),
	}
}

type sysArmyLogic struct {
	mutex 		sync.Mutex
	sysArmys    map[int][]*model.Army //key:posId 系统建筑军队
}

func (this * sysArmyLogic) GetArmy(x, y int) []*model.Army {
	posId := ToPosition(x, y)
	this.mutex.Lock()
	a, ok := this.sysArmys[posId]
	this.mutex.Unlock()
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
				soilder := 100*int(cfg.Level)
				npc, ok1 := static_conf.Basic.GetNPC(cfg.Level)
				if ok1 {
					soilder = npc.Soilders
				}

				scnt := []int{soilder, soilder, soilder}
				army := &model.Army{RId: 0, Order: 0, CityId: 0,
					GeneralArray: gsId, SoldierArray: scnt}
				army.ToGeneral()
				army.ToSoldier()

				armys = append(armys, army)
				posId := ToPosition(x, y)
				this.sysArmys[posId] = armys

				return armys
			}else{
				return armys
			}
		}else{
			return armys
		}
	}
}

func (this * sysArmyLogic) DelArmy(x, y int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	posId := ToPosition(x, y)
	delete(this.sysArmys, posId)
}


