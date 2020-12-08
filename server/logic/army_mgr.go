package logic

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/server/model"
	"slgserver/server/static_conf"
	"sync"
	"time"
)

func RoleArmyExtra(army* model.Army) {
	ra, ok := RAttributeMgr.Get(army.RId)
	if ok {
		army.UnionId = ra.UnionId
	}
}

type armyMgr struct {
	mutex        	sync.RWMutex
	armyById     	map[int]*model.Army      //key:armyId
	armyByCityId 	map[int][]*model.Army    //key:cityId
	armyByEndTime	map[int64][]*model.Army //key:到达时间
	armyByRId		map[int][]*model.Army   //key:rid
}

var AMgr = &armyMgr{
	armyById:     make(map[int]*model.Army),
	armyByCityId: make(map[int][]*model.Army),
	armyByEndTime: make(map[int64][]*model.Army),
	armyByRId: make(map[int][]*model.Army),
}

func (this*armyMgr) Load() {
	this.mutex.Lock()
	db.MasterDB.Table(model.Army{}).Find(this.armyById)

	for _, army := range this.armyById {
		RoleArmyExtra(army)
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

		//恢复已经执行行动的军队
		if army.Cmd != model.ArmyCmdIdle {
			e := army.End.Unix()
			_, ok := this.armyByEndTime[e]
			if ok == false{
				this.armyByEndTime[e] = make([]*model.Army, 0)
			}
			this.armyByEndTime[e] = append(this.armyByEndTime[e], army)
		}

		this.updateGenerals(army)
	}

	curTime := time.Now().Unix()
	for kT, armys := range this.armyByEndTime {
		if kT <= curTime {
			for _, a := range armys {
				if a.Cmd == model.ArmyCmdAttack {
					ArmyLogic.Arrive(a)
				}else if a.Cmd == model.ArmyCmdDefend {
					ArmyLogic.Arrive(a)
				}else if a.Cmd == model.ArmyCmdBack {
					if curTime >= a.End.Unix() {
						a.ToX = a.FromX
						a.ToY = a.FromY
						a.Cmd = model.ArmyCmdIdle
						a.State = model.ArmyStop
					}
				}
				a.SyncExecute()
			}
			delete(this.armyByEndTime, kT)
		}else{
			for _, a := range armys {
				a.State = model.ArmyRunning
			}
		}
	}
	this.mutex.Unlock()
	go this.running()
}

func (this*armyMgr) insertOne(army *model.Army)  {
	RoleArmyExtra(army)
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

func (this*armyMgr) addAction(t int64, army *model.Army)  {
	_, ok := this.armyByEndTime[t]
	if ok == false {
		this.armyByEndTime[t] = make([]*model.Army, 0)
	}
	this.armyByEndTime[t] = append(this.armyByEndTime[t], army)
}

//把行动丢进来
func (this*armyMgr) PushAction(army *model.Army)  {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if army.Cmd == model.ArmyCmdAttack || army.Cmd == model.ArmyCmdDefend {
		t := army.End.Unix()
		this.addAction(t, army)

	}else if army.Cmd == model.ArmyCmdReclamation {
		if army.State == model.ArmyRunning {
			t := army.End.Unix()
			this.addAction(t, army)
		}else{
			costTime := static_conf.Basic.General.ReclamationTime
			t := army.End.Unix()+int64(costTime)
			this.addAction(t, army)
		}
	}else if army.Cmd == model.ArmyCmdBack {
		cur := time.Now()
		diff := army.End.Unix()-army.Start.Unix()
		if cur.Unix() < army.End.Unix(){
			diff = cur.Unix()-army.Start.Unix()
		}
		army.Start = cur
		army.End = cur.Add(time.Duration(diff) * time.Second)
		army.Cmd = model.ArmyCmdBack
		this.addAction(army.End.Unix(), army)
	}

	ArmyLogic.Update(army)

}

func (this*armyMgr) ArmyBack(army *model.Army)  {
	army.State = model.ArmyRunning
	army.Cmd = model.ArmyCmdBack

	this.mutex.Lock()
	t := army.End.Unix()
	if actions, ok := this.armyByEndTime[t]; ok {
		for i, v := range actions {
			if v.Id == army.Id{
				actions = append(actions[:i], actions[i+1:]...)
				this.armyByEndTime[t] = actions
				break
			}
		}
	}
	this.mutex.Unlock()
	this.PushAction(army)
}

func (this*armyMgr) Reclamation(army *model.Army)  {
	army.State = model.ArmyStop
	army.Cmd = model.ArmyCmdReclamation
	this.PushAction(army)
}

func (this*armyMgr) running() {
	for true {
		t := time.Now().Unix()
		time.Sleep(100*time.Millisecond)

		this.mutex.Lock()
		for k, armies := range this.armyByEndTime {
			if k <= t{
				for _, army := range armies {
					ArmyLogic.Arrive(army)
				}
				delete(this.armyByEndTime, k)
			}
		}
		this.mutex.Unlock()
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
	return speed
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


