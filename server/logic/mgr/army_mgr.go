package mgr

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/model"
	"slgserver/server/static_conf/facility"
	"sync"
)


type armyMgr struct {
	mutex        	sync.RWMutex
	armyById     	map[int]*model.Army      //key:armyId
	armyByCityId 	map[int][]*model.Army    //key:cityId
	armyByRId		map[int][]*model.Army   //key:rid
}

var AMgr = &armyMgr{
	armyById:    	make(map[int]*model.Army),
	armyByCityId: 	make(map[int][]*model.Army),
	armyByRId: 		make(map[int][]*model.Army),
}

func (this*armyMgr) Load() {
	this.mutex.Lock()
	db.MasterDB.Table(model.Army{}).Find(this.armyById)

	for _, army := range this.armyById {
		cid := army.CityId
		c,ok:= this.armyByCityId[cid]
		if ok {
			this.armyByCityId[cid] = append(c, army)
		}else{
			this.armyByCityId[cid] = make([]*model.Army, 0)
			this.armyByCityId[cid] = append(this.armyByCityId[cid], army)
		}

		//rid
		if _, ok := this.armyByRId[army.RId]; ok == false{
			this.armyByRId[army.RId] = make([]*model.Army, 0)
		}
		this.armyByRId[army.RId] = append(this.armyByRId[army.RId], army)
		this.updateGenerals(army)
	}

	this.mutex.Unlock()
}

func (this*armyMgr) insertOne(army *model.Army)  {

	aid := army.Id
	cid := army.CityId

	this.armyById[aid] = army
	if _, r:= this.armyByCityId[cid]; r == false{
		this.armyByCityId[cid] = make([]*model.Army, 0)
	}
	this.armyByCityId[cid] = append(this.armyByCityId[cid], army)

	if _, ok := this.armyByRId[army.RId]; ok == false{
		this.armyByRId[army.RId] = make([]*model.Army, 0)
	}
	this.armyByRId[army.RId] = append(this.armyByRId[army.RId], army)

	this.updateGenerals(army)

}

func (this*armyMgr) insertMutil(armys []*model.Army)  {
	for _, v := range armys {
		this.insertOne(v)
	}
}


func (this*armyMgr) Get(aid int) (*model.Army, bool){
	this.mutex.RLock()
	a, ok := this.armyById[aid]
	this.mutex.RUnlock()
	if ok {
		return a, true
	}else{
		army := &model.Army{}
		ok, err := db.MasterDB.Table(model.Army{}).Where("id=?", aid).Get(army)
		if ok {
			this.mutex.Lock()
			this.insertOne(army)
			this.mutex.Unlock()
			return army, true
		}else{
			if err == nil{
				log.DefaultLog.Warn("armyMgr GetByRId armyId db not found",
					zap.Int("armyId", aid))
				return nil, false
			}else{
				log.DefaultLog.Warn("armyMgr GetByRId db error", zap.Int("armyId", aid))
				return nil, false
			}
		}
	}
}

func (this*armyMgr) GetByCity(cid int) ([]*model.Army, bool){
	this.mutex.RLock()
	a,ok := this.armyByCityId[cid]
	this.mutex.RUnlock()
	if ok {
		return a, true
	}else{
		m := make([]*model.Army, 0)
		err := db.MasterDB.Table(model.Army{}).Where("cityId=?", cid).Find(&m)
		if err!=nil{
			log.DefaultLog.Warn("armyMgr GetByCity db error", zap.Int("cityId", cid))
			return m, false
		}else{
			this.mutex.Lock()
			this.insertMutil(m)
			this.mutex.Unlock()
			return m, true
		}
	}
}

func (this*armyMgr) GetByRId(rid int) ([]*model.Army, bool){
	this.mutex.RLock()
	a,ok := this.armyByRId[rid]
	this.mutex.RUnlock()
	return a, ok
}

func (this*armyMgr) GetOrCreate(rid int, cid int, order int8) (*model.Army, error){

	this.mutex.RLock()
	armys, ok := this.armyByCityId[cid]
	this.mutex.RUnlock()

	if ok {
		for _, v := range armys {
			if v.Order == order{
				return v, nil
			}
		}
	}

	//需要创建
	army := &model.Army{RId: rid, Order: order,
		CityId: cid, Generals: `[0,0,0]`, Soldiers: `[0,0,0]`,
		GeneralArray: []int{0,0,0}, SoldierArray: []int{0,0,0}}

	_, err := db.MasterDB.Insert(army)
	if err == nil{
		this.mutex.Lock()
		this.insertOne(army)
		this.mutex.Unlock()
		return army, nil
	}else{
		log.DefaultLog.Warn("db error", zap.Error(err))
		return nil, err
	}
}

func (this*armyMgr) GetSpeed(army* model.Army) int{
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
	if camp > 0{
		campAdds = RFMgr.GetAdditions(army.CityId, facility.TypeHanAddition-1+camp)
	}
	return speed + campAdds[0]
}

//能否上阵
func (this*armyMgr) IsCanDispose(rid int, cfgId int) bool{
	armys, ok := this.GetByRId(rid)
	if ok == false{
		return true
	}

	for _, army := range armys {
		for _, g := range army.Gens {
			if g != nil {
				if g.CfgId == cfgId && g.CityId != 0{
					return false
				}
			}
		}
	}
	return true
}

func (this*armyMgr) updateGenerals(armys... *model.Army) {
	for _, army := range armys {
		army.Gens = make([]*model.General, 0)
		for _, gid := range army.GeneralArray {
			if gid == 0{
				army.Gens = append(army.Gens, nil)
			}else{
				g, _ := GMgr.GetByGId(gid)
				army.Gens = append(army.Gens, g)
			}
		}
	}
}

func (this* armyMgr) All()[]*model.Army {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	armys := make([]*model.Army, 0)
	for _, army := range this.armyById {
		armys = append(armys, army)
	}
	return armys
}

