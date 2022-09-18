package mgr

import (
	"sync"
	"time"

	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/server/slgserver/static_conf/facility"
	"github.com/llr104/slgserver/util"
	"go.uber.org/zap"
)

type roleResMgr struct {
	mutex    sync.RWMutex
	rolesRes map[int]*model.RoleRes
}

var RResMgr = &roleResMgr{
	rolesRes: make(map[int]*model.RoleRes),
}

//获取产量
func GetYield(rid int) model.Yield {
	by := RBMgr.GetYield(rid)
	cy := RFMgr.GetYield(rid)
	var y model.Yield

	y.Gold = by.Gold + cy.Gold + static_conf.Basic.Role.GoldYield
	y.Stone = by.Stone + cy.Stone + static_conf.Basic.Role.StoneYield
	y.Iron = by.Iron + cy.Iron + static_conf.Basic.Role.IronYield
	y.Grain = by.Grain + cy.Grain + static_conf.Basic.Role.GrainYield
	y.Wood = by.Wood + cy.Wood + static_conf.Basic.Role.WoodYield

	return y
}

//获取仓库容量
func GetDepotCapacity(rid int) int {
	return RFMgr.GetDepotCapacity(rid) + static_conf.Basic.Role.DepotCapacity
}

func (this *roleResMgr) Load() {

	rr := make([]*model.RoleRes, 0)
	err := db.MasterDB.Find(&rr)
	if err != nil {
		log.DefaultLog.Error("roleResMgr load role_res table error")
	}

	for _, v := range rr {
		this.rolesRes[v.RId] = v
	}

	go this.produce()

}

func (this *roleResMgr) Get(rid int) (*model.RoleRes, bool) {

	this.mutex.RLock()
	r, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if ok {
		return r, true
	}

	m := &model.RoleRes{}
	ok, err := db.MasterDB.Table(new(model.RoleRes)).Where("rid=?", rid).Get(m)
	if ok {

		this.mutex.Lock()
		this.rolesRes[rid] = m
		this.mutex.Unlock()

		return m, true
	} else {
		if err == nil {
			log.DefaultLog.Warn("RoleRes not found", zap.Int("rid", rid))
			return nil, false
		} else {
			log.DefaultLog.Warn("db error", zap.Error(err))
			return nil, false
		}
	}
}

func (this *roleResMgr) Add(res *model.RoleRes) {

	this.mutex.Lock()
	this.rolesRes[res.RId] = res
	this.mutex.Unlock()
}

func (this *roleResMgr) TryUseNeed(rid int, need facility.NeedRes) int {

	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if ok {
		if need.Decree <= rr.Decree && need.Grain <= rr.Grain &&
			need.Stone <= rr.Stone && need.Wood <= rr.Wood &&
			need.Iron <= rr.Iron && need.Gold <= rr.Gold {
			rr.Decree -= need.Decree
			rr.Iron -= need.Iron
			rr.Wood -= need.Wood
			rr.Stone -= need.Stone
			rr.Grain -= need.Grain
			rr.Gold -= need.Gold

			rr.SyncExecute()
			return constant.OK
		} else {
			if need.Decree > rr.Decree {
				return constant.DecreeNotEnough
			} else {
				return constant.ResNotEnough
			}
		}
	} else {
		return constant.RoleNotExist
	}
}

//政令是否足够
func (this *roleResMgr) DecreeIsEnough(rid int, cost int) bool {

	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if ok {
		if rr.Decree >= cost {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (this *roleResMgr) TryUseDecree(rid int, decree int) bool {

	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if ok {
		if rr.Decree >= decree {
			rr.Decree -= decree
			rr.SyncExecute()
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

//金币是否足够
func (this *roleResMgr) GoldIsEnough(rid int, cost int) bool {

	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if ok {
		if rr.Gold >= cost {
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (this *roleResMgr) TryUseGold(rid int, gold int) bool {

	this.mutex.RLock()
	rr, ok := this.rolesRes[rid]
	this.mutex.RUnlock()

	if ok {
		if rr.Gold >= gold {
			rr.Gold -= gold
			rr.SyncExecute()
			return true
		} else {
			return false
		}
	} else {
		return false
	}
}

func (this *roleResMgr) produce() {
	index := 1
	for true {
		t := static_conf.Basic.Role.RecoveryTime
		time.Sleep(time.Duration(t) * time.Second)
		this.mutex.RLock()
		for _, v := range this.rolesRes {
			//加判断是因为爆仓了，资源不无故减少
			capacity := GetDepotCapacity(v.RId)
			yield := GetYield(v.RId)
			if v.Wood < capacity {
				v.Wood = util.MinInt(v.Wood+yield.Wood/6, capacity)
			}
			if v.Iron < capacity {
				v.Iron = util.MinInt(v.Iron+yield.Iron/6, capacity)
			}

			if v.Stone < capacity {
				v.Stone = util.MinInt(v.Stone+yield.Stone/6, capacity)
			}

			if v.Grain < capacity {
				v.Grain = util.MinInt(v.Grain+yield.Grain/6, capacity)
			}

			if v.Gold < capacity {
				v.Gold = util.MinInt(v.Gold+yield.Gold/6, capacity)
			}

			if index%6 == 0 {
				if v.Decree < static_conf.Basic.Role.DecreeLimit {
					v.Decree += 1
				}
			}
			v.SyncExecute()
		}
		index++
		this.mutex.RUnlock()
	}
}
