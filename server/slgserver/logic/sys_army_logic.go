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

func (this* sysArmyLogic) getArmyCfg(x, y int) (star int8, lv int8, soilders int) {
	defender := 1
	star = 3
	lv = 5
	soilders = 100

	if mapBuild, ok := mgr.NMMgr.PositionBuild(x, y); ok{
		cfg, ok := static_conf.MapBuildConf.BuildConfig(mapBuild.Type, mapBuild.Level)
		if ok {
			defender = cfg.Defender
			if npc, ok := static_conf.Basic.GetNPC(cfg.Level); ok {
				soilders = npc.Soilders
			}
		}
	}

	if defender == 1{
		star = 3
		lv = 5
	}else if defender == 2{
		star = 4
		lv = 10
	}else {
		star = 5
		lv = 20
	}

	return star, lv, soilders
}

func (this *sysArmyLogic) GetArmy(x, y int) []*model.Army {
	posId := global.ToPosition(x, y)
	this.mutex.Lock()
	a, ok := this.sysArmys[posId]
	this.mutex.Unlock()
	if ok {
		return a
	}else{
		armys := make([]*model.Army, 0)

		star, lv, soilders := this.getArmyCfg(x, y)
		out, ok := mgr.GMgr.GetNPCGenerals(3, star, lv)
		if ok {
			gsId := make([]int, 0)
			gs := make([]*model.General, 3)

			for i := 0; i < len(out) ; i++ {
				gs[i] = &out[i]
			}

			scnt := []int{soilders, soilders, soilders}
			army := &model.Army{RId: 0, Order: 0, CityId: 0,
				GeneralArray: gsId, Gens: gs, SoldierArray: scnt}
			army.ToGeneral()
			army.ToSoldier()

			armys = append(armys, army)
			posId := global.ToPosition(x, y)
			this.sysArmys[posId] = armys
		}

		return armys
	}
}

func (this *sysArmyLogic) DelArmy(x, y int) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	posId := global.ToPosition(x, y)
	delete(this.sysArmys, posId)
}


