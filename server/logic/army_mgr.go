package logic

import (
	"errors"
	"fmt"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
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
		if v.State != model.ArmyIdle {
			e := v.End.Unix()
			_, ok := this.armyByEndTime[e]
			if ok == false{
				this.armyByEndTime[e] = make([]*model.Army, 0)
			}
			this.armyByEndTime[e] = append(this.armyByEndTime[e], v)
		}
	}

	cur_t := time.Now().Unix()
	for k_t, armys := range this.armyByEndTime {
		if k_t <= cur_t {
			for _, a := range armys {
				if a.State == model.ArmyAttack{
					ArmyLogic.Arrive(a)
				}else if a.State == model.ArmyDefend{

				}else if a.State == model.ArmyBack {
					if cur_t >= a.End.Unix() {
						a.ToX = a.FromX
						a.ToY = a.FromY
						a.State = model.ArmyIdle
					}
				}
				a.NeedUpdate = true
			}
			delete(this.armyByEndTime, k_t)
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

	t := army.End.Unix()
	_, ok := this.armyByEndTime[t]
	if ok == false {
		this.armyByEndTime[t] = make([]*model.Army, 0)
	}
	this.armyByEndTime[t] = append(this.armyByEndTime[t], army)

}


func (this* ArmyMgr) running() {
	for true {
		t := time.Now().Unix()
		time.Sleep(1*time.Second)

		this.mutex.Lock()
		//往前5秒找，以防有些占用太久没有执行到
		for i := t-5; i <= t ; i++ {
			arr, ok := this.armyByEndTime[i]
			if ok {
				for _, army := range arr {
					ArmyLogic.Arrive(army)
				}
			}
			delete(this.armyByEndTime, i)
		}
		this.mutex.Unlock()
	}
}

func (this* ArmyMgr) toDatabase() {
	for true {
		time.Sleep(5*time.Second)
		this.mutex.RLock()
		cnt :=0
		for _, v := range this.armyById {
			if v.NeedUpdate {
				cnt+=1
				_, err := db.MasterDB.Table(model.Army{}).Cols("firstId",
					"secondId", "thirdId", "first_soldier_cnt",
					"second_soldier_cnt", "third_soldier_cnt", "state",
					"from_x", "from_y", "to_x", "to_y", "start", "end").Update(v)
				if err != nil{
					log.DefaultLog.Warn("db error", zap.Error(err))
				}else{
					v.NeedUpdate = false
				}
			}

			//一次最多更新20个
			if cnt>20{
				break
			}
		}

		this.mutex.RUnlock()
	}
}

func (this* ArmyMgr) Get(aid int) (*model.Army, error){
	this.mutex.RLock()
	a, ok := this.armyById[aid]
	this.mutex.RUnlock()
	if ok {
		return a, nil
	}else{
		army := &model.Army{}
		ok, err := db.MasterDB.Table(model.Army{}).Where("id=?", aid).Get(army)
		if ok {
			this.mutex.Lock()
			this.insertOne(army)
			this.mutex.Unlock()
			return army, nil
		}else{
			if err == nil{
				str := fmt.Sprintf("ArmyMgr Get armyId:%d db not found", aid)
				log.DefaultLog.Warn(str)
				return nil, errors.New(str)
			}else{
				log.DefaultLog.Warn("ArmyMgr Get db error", zap.Int("armyId", aid))
				return nil, err
			}
		}
	}
}

func (this* ArmyMgr) GetByCity(cid int) ([]*model.Army, error){
	this.mutex.RLock()
	a,ok := this.armyByCityId[cid]
	this.mutex.RUnlock()
	if ok {
		return a, nil
	}else{
		m := make([]*model.Army, 0)
		err := db.MasterDB.Table(model.Army{}).Where("cityId=?", cid).Find(&m)
		if err!=nil{
			log.DefaultLog.Warn("ArmyMgr GetByCity db error", zap.Int("cityId", cid))
			return m, err
		}else{
			this.mutex.Lock()
			this.insertMutil(m)
			this.mutex.Unlock()
			return m, nil
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
	army := &model.Army{RId: rid, Order: order, CityId: cid,
		FirstId: 0, SecondId: 0, ThirdId: 0,
		FirstSoldierCnt: 0, SecondSoldierCnt: 0, ThirdSoldierCnt: 0}
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


