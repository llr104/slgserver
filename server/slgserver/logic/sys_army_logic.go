package logic

import (
	"slgserver/server/slgserver/global"
	"slgserver/server/slgserver/logic/mgr"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/static_conf"
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

func (this *sysArmyLogic) GetArmy(x, y int) []*model.Army {
	posId := global.ToPosition(x, y)
	this.mutex.Lock()
	a, ok := this.sysArmys[posId]
	this.mutex.Unlock()
	if ok {
		return a
	}else{
		out, ok := mgr.GMgr.GetNPCGenerals(3)
		gsId := make([]int, 0)
		gs := make([]*model.General, 3)

		for i := 0; i < len(out) ; i++ {
			gs[i] = &out[i]
		}

		armys := make([]*model.Army, 0)
		if ok {
			if cfg, ok := mgr.NMMgr.PositionBuild(x, y); ok{
				soilder := 100*int(cfg.Level)
				npc, ok1 := static_conf.Basic.GetNPC(cfg.Level)
				if ok1 {
					soilder = npc.Soilders
				}

				scnt := []int{soilder, soilder, soilder}
				army := &model.Army{RId: 0, Order: 0, CityId: 0,
					GeneralArray: gsId, Gens: gs, SoldierArray: scnt}
				army.ToGeneral()
				army.ToSoldier()

				armys = append(armys, army)
				posId := global.ToPosition(x, y)
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

func (this *sysArmyLogic) DelArmy(x, y int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	posId := global.ToPosition(x, y)
	delete(this.sysArmys, posId)
}


