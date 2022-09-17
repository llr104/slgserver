package mgr

import (
	"sync"
	"time"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/server/slgserver/static_conf/general"
	"github.com/llr104/slgserver/server/slgserver/static_conf/npc"
	"github.com/llr104/slgserver/util"
	"go.uber.org/zap"
)

type generalMgr struct {
	mutex     sync.RWMutex
	genByRole map[int][]*model.General
	genByGId  map[int]*model.General
}

var GMgr = &generalMgr{
	genByRole: make(map[int][]*model.General),
	genByGId:  make(map[int]*model.General),
}

func (this *generalMgr) updatePhysicalPower() {
	limit := static_conf.Basic.General.PhysicalPowerLimit
	recoverCnt := static_conf.Basic.General.RecoveryPhysicalPower
	for true {
		time.Sleep(1 * time.Hour)
		this.mutex.RLock()
		for _, g := range this.genByGId {
			if g.PhysicalPower < limit {
				g.PhysicalPower = util.MinInt(limit, g.PhysicalPower+recoverCnt)
				g.SyncExecute()
			}
		}
		this.mutex.RUnlock()
	}
}

//创建npc
func (this *generalMgr) createNPC() ([]*model.General, bool) {

	gs := make([]*model.General, 0)

	for _, armys := range npc.Cfg.Armys {
		for _, cfgs := range armys.ArmyCfg {
			for i, cfgId := range cfgs.CfgIds {
				r, ok := this.NewGeneral(cfgId, 0, cfgs.Lvs[i])
				if ok == false {
					return nil, false
				}
				gs = append(gs, r)
			}
		}

	}

	return gs, true
}

func (this *generalMgr) add(g *model.General) {
	this.mutex.Lock()

	if _, ok := this.genByRole[g.RId]; ok == false {
		this.genByRole[g.RId] = make([]*model.General, 0)
	}
	this.genByRole[g.RId] = append(this.genByRole[g.RId], g)
	this.genByGId[g.Id] = g

	this.mutex.Unlock()
}

func (this *generalMgr) Load() {

	err := db.MasterDB.Table(model.General{}).Where("state=?",
		model.GeneralNormal).Find(this.genByGId)

	if err != nil {
		log.DefaultLog.Warn("db error", zap.Error(err))
		return
	}

	for _, v := range this.genByGId {
		if _, ok := this.genByRole[v.RId]; ok == false {
			this.genByRole[v.RId] = make([]*model.General, 0)
		}
		this.genByRole[v.RId] = append(this.genByRole[v.RId], v)
	}

	if len(this.genByGId) == 0 {
		this.createNPC()
	}

	go this.updatePhysicalPower()
}

func (this *generalMgr) GetByRId(rid int) ([]*model.General, bool) {
	this.mutex.Lock()
	r, ok := this.genByRole[rid]
	this.mutex.Unlock()

	if ok {
		out := make([]*model.General, 0)
		for _, g := range r {
			if g.IsActive() {
				out = append(out, g)
			}
		}
		return out, true
	}

	gs := make([]*model.General, 0)
	err := db.MasterDB.Table(new(model.General)).Where(
		"rid=? and state=?", rid, model.GeneralNormal).Find(&gs)

	if err == nil {
		if len(gs) > 0 {
			for _, g := range gs {
				this.add(g)
			}
			return gs, true
		} else {
			log.DefaultLog.Warn("general not fount", zap.Int("rid", rid))
			return nil, false
		}
	} else {
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, false
	}
}

//查找将领
func (this *generalMgr) GetByGId(gid int) (*model.General, bool) {
	this.mutex.RLock()
	g, ok := this.genByGId[gid]
	this.mutex.RUnlock()
	if ok {
		if g.IsActive() {
			return g, true
		} else {
			return nil, false
		}
	} else {

		g := &model.General{}
		r, err := db.MasterDB.Table(new(model.General)).Where(
			"id=? and state=?", gid, model.GeneralNormal).Get(g)

		if err == nil {
			if r {
				this.add(g)
				return g, true
			} else {
				log.DefaultLog.Warn("general gid not found", zap.Int("gid", gid))
				return nil, false
			}

		} else {
			log.DefaultLog.Warn("general gid not found", zap.Int("gid", gid))
			return nil, false
		}
	}
}

//这个角色是否有这个武将
func (this *generalMgr) HasGeneral(rid int, gid int) (*model.General, bool) {
	r, ok := this.GetByRId(rid)
	if ok {
		for _, v := range r {
			t := v
			if t.Id == gid {
				return t, true
			}
		}
	}
	return nil, false
}

func (this *generalMgr) HasGenerals(rid int, gIds []int) ([]*model.General, bool) {
	gs := make([]*model.General, 0)
	for i := 0; i < len(gIds); i++ {
		g, ok := this.HasGeneral(rid, gIds[i])
		if ok {
			gs = append(gs, g)
		} else {
			return gs, false
		}
	}
	return gs, true
}

func (this *generalMgr) Count(rid int) int {
	gs, ok := this.GetByRId(rid)
	if ok {
		return len(gs)
	} else {
		return 0
	}
}

func (this *generalMgr) NewGeneral(cfgId int, rid int, level int8) (*model.General, bool) {
	g, ok := model.NewGeneral(cfgId, rid, level)
	if ok {
		this.add(g)
	}
	return g, ok
}

/*
如果不存在则去创建
*/
func (this *generalMgr) GetOrCreateByRId(rid int) ([]*model.General, bool) {
	r, ok := this.GetByRId(rid)
	if ok {
		return r, true
	} else {
		//创建
		gs, ok := this.RandCreateGeneral(rid, 3)
		if ok == false {
			return nil, false
		}
		return gs, true
	}
}

/*
随机创建一个
*/
func (this *generalMgr) RandCreateGeneral(rid int, nums int) ([]*model.General, bool) {
	gs := make([]*model.General, 0)

	for i := 0; i < nums; i++ {
		cfgId := general.General.Draw()
		g, ok := this.NewGeneral(cfgId, rid, 1)
		if ok == false {
			return nil, false
		}
		gs = append(gs, g)
	}

	return gs, true
}

//获取npc武将
func (this *generalMgr) GetNPCGenerals(cfgIds []int, levels []int8) ([]model.General, bool) {
	if len(cfgIds) != len(levels) {
		return nil, false
	}
	gs, ok := this.GetByRId(0)
	if ok == false {
		return nil, false
	} else {
		target := make([]model.General, 0)
		for i := 0; i < len(cfgIds); i++ {
			for _, g := range gs {
				if g.Level == levels[i] && g.CfgId == cfgIds[i] {
					target = append(target, *g)
					break
				}
			}
		}

		return target, true
	}
}

func (this *generalMgr) GetDestroy(army *model.Army) int {
	destroy := 0
	for _, g := range army.Gens {
		if g != nil {
			destroy += g.GetDestroy()
		}
	}
	return destroy
}

//体力是否足够
func (this *generalMgr) PhysicalPowerIsEnough(army *model.Army, cost int) bool {
	for _, g := range army.Gens {
		if g == nil {
			continue
		}
		if g.PhysicalPower < cost {
			return false
		}
	}
	return true
}

//尝试使用体力
func (this *generalMgr) TryUsePhysicalPower(army *model.Army, cost int) bool {

	if this.PhysicalPowerIsEnough(army, cost) == false {
		return false
	}

	for _, g := range army.Gens {
		if g == nil {
			continue
		}
		g.PhysicalPower -= cost
		g.SyncExecute()
	}

	return true
}
