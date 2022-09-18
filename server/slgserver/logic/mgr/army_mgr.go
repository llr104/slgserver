package mgr

import (
	"sync"

	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/server/slgserver/static_conf/facility"
	"go.uber.org/zap"
)

type armyMgr struct {
	mutex        sync.RWMutex
	armyById     map[int]*model.Army   //key:armyId
	armyByCityId map[int][]*model.Army //key:cityId
	armyByRId    map[int][]*model.Army //key:rid
}

var AMgr = &armyMgr{
	armyById:     make(map[int]*model.Army),
	armyByCityId: make(map[int][]*model.Army),
	armyByRId:    make(map[int][]*model.Army),
}

func (this *armyMgr) Load() {

	db.MasterDB.Table(model.Army{}).Find(this.armyById)

	for _, army := range this.armyById {
		//处理征兵
		army.CheckConscript()
		cid := army.CityId
		c, ok := this.armyByCityId[cid]
		if ok {
			this.armyByCityId[cid] = append(c, army)
		} else {
			this.armyByCityId[cid] = make([]*model.Army, 0)
			this.armyByCityId[cid] = append(this.armyByCityId[cid], army)
		}

		//rid
		if _, ok := this.armyByRId[army.RId]; ok == false {
			this.armyByRId[army.RId] = make([]*model.Army, 0)
		}
		this.armyByRId[army.RId] = append(this.armyByRId[army.RId], army)
		this.updateGenerals(army)
	}

}

func (this *armyMgr) insertOne(army *model.Army) {

	aid := army.Id
	cid := army.CityId

	this.armyById[aid] = army
	if _, r := this.armyByCityId[cid]; r == false {
		this.armyByCityId[cid] = make([]*model.Army, 0)
	}
	this.armyByCityId[cid] = append(this.armyByCityId[cid], army)

	if _, ok := this.armyByRId[army.RId]; ok == false {
		this.armyByRId[army.RId] = make([]*model.Army, 0)
	}
	this.armyByRId[army.RId] = append(this.armyByRId[army.RId], army)

	this.updateGenerals(army)

}

func (this *armyMgr) insertMutil(armys []*model.Army) {
	for _, v := range armys {
		this.insertOne(v)
	}
}

func (this *armyMgr) Get(aid int) (*model.Army, bool) {
	this.mutex.RLock()
	a, ok := this.armyById[aid]
	this.mutex.RUnlock()
	if ok {
		a.CheckConscript()
		return a, true
	} else {
		army := &model.Army{}
		ok, err := db.MasterDB.Table(model.Army{}).Where("id=?", aid).Get(army)
		if ok {
			this.mutex.Lock()
			this.insertOne(army)
			this.mutex.Unlock()
			return army, true
		} else {
			if err == nil {
				log.DefaultLog.Warn("armyMgr GetByRId armyId db not found",
					zap.Int("armyId", aid))
				return nil, false
			} else {
				log.DefaultLog.Warn("armyMgr GetByRId db error", zap.Int("armyId", aid))
				return nil, false
			}
		}
	}
}

func (this *armyMgr) GetByCity(cid int) ([]*model.Army, bool) {
	this.mutex.RLock()
	as, ok := this.armyByCityId[cid]
	this.mutex.RUnlock()
	if ok {
		for _, a := range as {
			a.CheckConscript()
		}
		return as, true
	} else {
		m := make([]*model.Army, 0)
		err := db.MasterDB.Table(model.Army{}).Where("cityId=?", cid).Find(&m)
		if err != nil {
			log.DefaultLog.Warn("armyMgr GetByCity db error", zap.Int("cityId", cid))
			return m, false
		} else {
			this.mutex.Lock()
			this.insertMutil(m)
			this.mutex.Unlock()
			return m, true
		}
	}
}

func (this *armyMgr) GetByCityOrder(cid int, order int8) (*model.Army, bool) {
	rs, ok := this.GetByCity(cid)
	if ok {
		for _, r := range rs {
			if r.Order == order {
				return r, true
			}
		}
	} else {
		return nil, false
	}
	return nil, false
}

func (this *armyMgr) GetByRId(rid int) ([]*model.Army, bool) {
	this.mutex.RLock()
	as, ok := this.armyByRId[rid]
	this.mutex.RUnlock()

	if ok {
		for _, a := range as {
			a.CheckConscript()
		}
	}
	return as, ok
}

//归属于该位置的军队数量
func (this *armyMgr) BelongPosArmyCnt(rid int, x, y int) int {
	cnt := 0
	armys, ok := this.GetByRId(rid)
	if ok {
		for _, army := range armys {
			if army.FromX == x && army.FromY == y {
				cnt += 1
			} else if army.Cmd == model.ArmyCmdTransfer && army.ToX == x && army.ToY == y {
				cnt += 1
			}
		}
	}

	return cnt
}

func (this *armyMgr) GetOrCreate(rid int, cid int, order int8) (*model.Army, error) {

	this.mutex.RLock()
	armys, ok := this.armyByCityId[cid]
	this.mutex.RUnlock()

	if ok {
		for _, v := range armys {
			if v.Order == order {
				return v, nil
			}
		}
	}

	//需要创建
	army := &model.Army{RId: rid,
		Order:              order,
		CityId:             cid,
		Generals:           `[0,0,0]`,
		Soldiers:           `[0,0,0]`,
		GeneralArray:       [static_conf.ArmyGCnt]int{0, 0, 0},
		SoldierArray:       [static_conf.ArmyGCnt]int{0, 0, 0},
		ConscriptCnts:      `[0,0,0]`,
		ConscriptTimes:     `[0,0,0]`,
		ConscriptCntArray:  [static_conf.ArmyGCnt]int{0, 0, 0},
		ConscriptTimeArray: [static_conf.ArmyGCnt]int64{0, 0, 0},
	}

	city, ok := RCMgr.Get(cid)
	if ok {
		army.FromX = city.X
		army.FromY = city.Y
		army.ToX = city.X
		army.ToY = city.Y
	}

	_, err := db.MasterDB.Insert(army)
	if err == nil {
		this.mutex.Lock()
		this.insertOne(army)
		this.mutex.Unlock()
		return army, nil
	} else {
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, err
	}
}

func (this *armyMgr) GetSpeed(army *model.Army) int {
	speed := 100000
	for _, g := range army.Gens {
		if g != nil {
			s := g.GetSpeed()
			if s < speed {
				speed = s
			}
		}
	}

	//阵营加成
	camp := army.GetCamp()
	campAdds := []int{0}
	if camp > 0 {
		campAdds = RFMgr.GetAdditions(army.CityId, facility.TypeHanAddition-1+camp)
	}
	return speed + campAdds[0]
}

//能否已经重复上阵了
func (this *armyMgr) IsRepeat(rid int, cfgId int) bool {
	armys, ok := this.GetByRId(rid)
	if ok == false {
		return true
	}

	for _, army := range armys {
		for _, g := range army.Gens {
			if g != nil {
				if g.CfgId == cfgId && g.CityId != 0 {
					return false
				}
			}
		}
	}
	return true
}

func (this *armyMgr) updateGenerals(armys ...*model.Army) {
	for _, army := range armys {
		for i, gid := range army.GeneralArray {
			if gid != 0 {
				g, _ := GMgr.GetByGId(gid)
				army.Gens[i] = g
			}
		}

	}
}

func (this *armyMgr) All() []*model.Army {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	armys := make([]*model.Army, 0)
	for _, army := range this.armyById {
		armys = append(armys, army)
	}
	return armys
}
