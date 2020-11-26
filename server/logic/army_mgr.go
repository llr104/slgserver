package logic

import (
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/server/static_conf"
	"sync"
	"time"
)

type ArmyMgr struct {
	mutex        	sync.RWMutex
	armyById     	map[int]*model.Army   	//key:armyId
	armyByCityId 	map[int][]*model.Army 	//key:cityId
	armyByEndTime	map[int64][]*model.Army	//key:到达时间
	armyByRId		map[int][]*model.Army	//key:rid
}

var AMgr = &ArmyMgr{
	armyById:     make(map[int]*model.Army),
	armyByCityId: make(map[int][]*model.Army),
	armyByEndTime: make(map[int64][]*model.Army),
	armyByRId: make(map[int][]*model.Army),
}

func (this* ArmyMgr) Load() {
	this.mutex.Lock()
	db.MasterDB.Table(model.Army{}).Find(this.armyById)

	for _, v := range this.armyById {
		cid := v.CityId
		c,ok:= this.armyByCityId[cid]
		if ok {
			this.armyByCityId[cid] = append(c, v)
		}else{
			this.armyByCityId[cid] = make([]*model.Army, 0)
			this.armyByCityId[cid] = append(this.armyByCityId[cid], v)
		}

		//rid
		if _, ok := this.armyByRId[v.RId]; ok == false{
			this.armyByRId[v.RId] = make([]*model.Army, 0)
		}
		this.armyByRId[v.RId] = append(this.armyByRId[v.RId], v)

		//恢复已经执行行动的军队
		if v.Cmd != model.ArmyCmdIdle {
			e := v.End.Unix()
			_, ok := this.armyByEndTime[e]
			if ok == false{
				this.armyByEndTime[e] = make([]*model.Army, 0)
			}
			this.armyByEndTime[e] = append(this.armyByEndTime[e], v)
		}
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
				a.DB.Sync()
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
	go this.toDatabase()
}

func (this* ArmyMgr) insertOne(army* model.Army)  {

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

}

func (this* ArmyMgr) insertMutil(armys[] *model.Army)  {
	for _, v := range armys {
		this.insertOne(v)
	}
}

//把行动丢进来
func (this* ArmyMgr) PushAction(army *model.Army)  {
	this.mutex.Lock()
	defer this.mutex.Unlock()

	if army.Cmd == model.ArmyCmdAttack || army.Cmd == model.ArmyCmdDefend{
		t := army.End.Unix()
		_, ok := this.armyByEndTime[t]
		if ok == false {
			this.armyByEndTime[t] = make([]*model.Army, 0)
		}
		this.armyByEndTime[t] = append(this.armyByEndTime[t], army)

	}else if army.Cmd == model.ArmyCmdReclamation{
		costTime := static_conf.Basic.General.ReclamationTime
		t := army.End.Unix()+int64(costTime)

		_, ok := this.armyByEndTime[t]
		if ok == false {
			this.armyByEndTime[t] = make([]*model.Army, 0)
		}
		this.armyByEndTime[t] = append(this.armyByEndTime[t], army)

	}else if army.Cmd == model.ArmyCmdBack{
		cur := time.Now()
		diff := army.End.Unix()-army.Start.Unix()
		if cur.Unix() < army.End.Unix(){
			diff = cur.Unix()-army.Start.Unix()
		}
		army.Start = cur
		army.End = cur.Add(time.Duration(diff) * time.Second)
		army.State = model.ArmyRunning
		army.Cmd = model.ArmyCmdBack
		army.DB.Sync()
	}

	ArmyLogic.Update(army)

}

func (this* ArmyMgr) ArmyBack(army *model.Army)  {
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

func (this* ArmyMgr) Reclamation(army *model.Army)  {
	army.State = model.ArmyStop
	army.Cmd = model.ArmyCmdReclamation
	this.PushAction(army)
}

func (this* ArmyMgr) running() {
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

func (this* ArmyMgr) toDatabase() {
	for true {
		time.Sleep(2*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.armyById {
			if v.DB.NeedSync() {
				v.DB.BeginSync()
				cnt+=1
				_, err := db.MasterDB.Table(model.Army{}).ID(v.Id).Cols("soldiers",
					"generals", "cmd", "from_x", "from_y", "to_x", "to_y", "start", "end").Update(v)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}
				v.DB.EndSync()
			}

			//一次最多更新20个
			if cnt>20{
				break
			}
		}

		this.mutex.RUnlock()
	}
}

func (this* ArmyMgr) Get(aid int) (*model.Army, bool){
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
				log.DefaultLog.Warn("ArmyMgr GetByRId armyId db not found",
					zap.Int("armyId", aid))
				return nil, false
			}else{
				log.DefaultLog.Warn("ArmyMgr GetByRId db error", zap.Int("armyId", aid))
				return nil, false
			}
		}
	}
}

func (this* ArmyMgr) GetByCity(cid int) ([]*model.Army, bool){
	this.mutex.RLock()
	a,ok := this.armyByCityId[cid]
	this.mutex.RUnlock()
	if ok {
		return a, true
	}else{
		m := make([]*model.Army, 0)
		err := db.MasterDB.Table(model.Army{}).Where("cityId=?", cid).Find(&m)
		if err!=nil{
			log.DefaultLog.Warn("ArmyMgr GetByCity db error", zap.Int("cityId", cid))
			return m, false
		}else{
			this.mutex.Lock()
			this.insertMutil(m)
			this.mutex.Unlock()
			return m, true
		}
	}
}

func (this* ArmyMgr) GetByRId(rid int) ([]*model.Army, bool){
	this.mutex.RLock()
	a,ok := this.armyByRId[rid]
	this.mutex.RUnlock()
	return a, ok
}

func (this* ArmyMgr) GetOrCreate(rid int, cid int, order int8) (*model.Army, error){

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


