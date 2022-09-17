package army

import (
	"sync"

	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/server/slgserver/static_conf/npc"
)

func NewSysArmy() *sysArmyLogic {
	return &sysArmyLogic{
		sysArmys: make(map[int][]*model.Army),
	}
}

type sysArmyLogic struct {
	mutex    sync.Mutex
	sysArmys map[int][]*model.Army //key:posId 系统建筑军队
}

func (this *sysArmyLogic) GetArmy(x, y int) []*model.Army {
	posId := global.ToPosition(x, y)
	this.mutex.Lock()
	a, ok := this.sysArmys[posId]
	this.mutex.Unlock()
	if ok {
		return a
	} else {
		armys := make([]*model.Army, 0)
		if mapBuild, ok := mgr.NMMgr.PositionBuild(x, y); ok {
			cfg, ok := static_conf.MapBuildConf.BuildConfig(mapBuild.Type, mapBuild.Level)
			if ok {
				soldiers := npc.Cfg.NPCSoilder(cfg.Level)
				ok, armyCfg := npc.Cfg.RandomOne(cfg.Level)
				out, ok := mgr.GMgr.GetNPCGenerals(armyCfg.CfgIds, armyCfg.Lvs)
				if ok {

					gsId := [static_conf.ArmyGCnt]int{}
					gs := [static_conf.ArmyGCnt]*model.General{}

					for i := 0; i < len(out) && i < static_conf.ArmyGCnt; i++ {
						gs[i] = &out[i]
						gsId[i] = out[i].Id
					}

					scnt := [static_conf.ArmyGCnt]int{soldiers, soldiers, soldiers}
					army := &model.Army{RId: 0, Order: 0, CityId: 0,
						GeneralArray: gsId, Gens: gs, SoldierArray: scnt}
					army.ToGeneral()
					army.ToSoldier()

					armys = append(armys, army)
					posId := global.ToPosition(x, y)
					this.sysArmys[posId] = armys
				}
			}
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
