package logic

import (
	"slgserver/server/slgserver/global"
	"slgserver/server/slgserver/logic/mgr"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/static_conf"
	"slgserver/util"
	"sync"
	"time"
)

type armyLogic struct {
	passby			sync.RWMutex
	timeMutex		sync.RWMutex

	sys            	*sysArmyLogic
	giveUpId       	chan int
	arriveArmys    	chan *model.Army
	updateArmys    	chan *model.Army
	outArmys		map[int]*model.Army

	armyByEndTime	map[int64][]*model.Army      //key:到达时间
	stopInPosArmys 	map[int]map[int]*model.Army //玩家停留位置的军队 key:posId,armyId
	passbyPosArmys 	map[int]map[int]*model.Army //玩家路过位置的军队 key:posId,armyId
}

func (this *armyLogic) init(){

	armys := mgr.AMgr.All()
	for _, army := range armys {
		//恢复已经执行行动的军队
		if army.Cmd != model.ArmyCmdIdle {
			e := army.End.Unix()
			_, ok := this.armyByEndTime[e]
			if ok == false{
				this.armyByEndTime[e] = make([]*model.Army, 0)
			}
			this.armyByEndTime[e] = append(this.armyByEndTime[e], army)
		}
	}

	curTime := time.Now().Unix()
	for kT, armys := range this.armyByEndTime {
		if kT <= curTime {
			for _, a := range armys {
				if a.Cmd == model.ArmyCmdAttack {
					this.Arrive(a)
				}else if a.Cmd == model.ArmyCmdDefend {
					this.Arrive(a)
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
}

func (this*armyLogic) check() {
	for true {
		t := time.Now().Unix()
		time.Sleep(100*time.Millisecond)

		this.timeMutex.Lock()
		for k, armies := range this.armyByEndTime {
			if k <= t{
				for _, army := range armies {
					this.Arrive(army)
				}
				delete(this.armyByEndTime, k)
			}
		}
		this.timeMutex.Unlock()
	}
}

func (this *armyLogic) running(){
	passbyTimer := time.NewTicker(10 * time.Second)
	for {
		select {
			case <-passbyTimer.C:{
				this.passby.Lock()
				this.passbyPosArmys = make(map[int]map[int]*model.Army)
				for _, army := range this.outArmys {
					if army.State == model.ArmyRunning {
						x, y := army.Position()
						posId := global.ToPosition(x, y)
						if _, ok := this.passbyPosArmys[posId]; ok == false {
							this.passbyPosArmys[posId] = make(map[int]*model.Army)
						}
						this.passbyPosArmys[posId][army.Id] = army
						army.CheckSyncCell()
					}
				}

				for posId, armys := range this.stopInPosArmys {
					for _, army := range armys {
						if _, ok := this.passbyPosArmys[posId]; ok == false {
							this.passbyPosArmys[posId] = make(map[int]*model.Army)
						}
						this.passbyPosArmys[posId][army.Id] = army
					}
				}

				this.passby.Unlock()
			}
			case army := <-this.updateArmys:{
				this.exeUpdate(army)
			}
			case army := <-this.arriveArmys:{
				this.exeArrive(army)
			}
			case giveUpId := <- this.giveUpId:{
				armys, ok := this.stopInPosArmys[giveUpId]
				if ok {
					for _, army := range armys {
						this.ArmyBack(army)
					}
					delete(this.stopInPosArmys, giveUpId)
				}
			}
		}
	}
}

func (this *armyLogic) exeUpdate(army *model.Army) {
	army.SyncExecute()
	if army.Cmd == model.ArmyCmdBack {
		this.deleteArmy(army.ToX, army.ToY)
	}

	if army.Cmd != model.ArmyCmdIdle {
		this.outArmys[army.Id] = army
	}else{
		delete(this.outArmys, army.RId)
	}
}

func (this *armyLogic) exeArrive(army *model.Army) {
	if army.Cmd == model.ArmyCmdAttack {
		if IsCanArrive(army.ToX, army.ToY, army.RId) &&
			IsWarFree(army.ToX, army.ToY) == false &&
			IsCanDefend(army.ToX, army.ToY, army.RId) == false{
			newBattle(army)
		} else{
			war := NewEmptyWar(army)
			war.SyncExecute()
		}
		this.ArmyBack(army)
	}else if army.Cmd == model.ArmyCmdDefend {
		//呆在哪里不动
		ok := IsCanDefend(army.ToX, army.ToY, army.RId)
		if ok {
			//目前是自己的领地才能驻守
			army.State = model.ArmyStop
			this.addArmy(army)
			this.Update(army)
		}else{
			war := NewEmptyWar(army)
			war.SyncExecute()
			this.ArmyBack(army)
		}

	}else if army.Cmd == model.ArmyCmdReclamation {
		if army.State == model.ArmyRunning {

			ok := mgr.RBMgr.BuildIsRId(army.ToX, army.ToY, army.RId)
			if ok  {
				//目前是自己的领地才能屯田
				this.addArmy(army)
				this.Reclamation(army)
			}else{
				war := NewEmptyWar(army)
				war.SyncExecute()
				this.ArmyBack(army)
			}

		}else {
			this.ArmyBack(army)
			//增加场量
			rr, ok := mgr.RResMgr.Get(army.RId)
			if ok {
				b, ok1 := mgr.RBMgr.PositionBuild(army.ToX, army.ToY)
				if ok1 {
					rr.Stone += b.Stone
					rr.Iron += b.Iron
					rr.Wood += b.Wood
					rr.Gold += rr.Gold
					rr.Grain += rr.Grain
					rr.SyncExecute()
				}
			}
		}
	}else if army.Cmd == model.ArmyCmdBack {
		army.State = model.ArmyStop
		army.Cmd = model.ArmyCmdIdle
		army.ToX = army.FromX
		army.ToY = army.FromY

		this.Update(army)
	}else if army.Cmd == model.ArmyCmdTransfer {
		//调动到位置了
		if army.State == model.ArmyRunning{
			army.State = model.ArmyStop
			army.Cmd = model.ArmyCmdIdle
			x := army.ToX
			y := army.ToY
			army.FromX = x
			army.FromY = y
			army.ToX = x
			army.ToY = y
			this.Update(army)
		}
	}
}

func (this *armyLogic) ScanBlock(rid, x, y, length int) []*model.Army {

	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}

	maxX := util.MinInt(global.MapWith, x+length-1)
	maxY := util.MinInt(global.MapHeight, y+length-1)
	out := make([]*model.Army, 0)
	this.passby.RLock()
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {

			posId := global.ToPosition(i, j)
			armys, ok := this.passbyPosArmys[posId]
			if ok {
				is := armyIsInView(rid, i, j)
				if is == false{
					continue
				}
				for _, army := range armys {
					out = append(out, army)
				}
			}
		}
	}
	this.passby.RUnlock()
	return out
}

func (this *armyLogic) Arrive(army *model.Army) {
	this.arriveArmys <- army
}

func (this *armyLogic) Update(army *model.Army) {
	this.updateArmys <- army
}

func (this *armyLogic) GiveUp(posId int) {
	this.giveUpId <- posId
}

func (this *armyLogic) deleteArmy(x, y int) {
	posId := global.ToPosition(x, y)
	delete(this.stopInPosArmys, posId)
}

func (this*armyLogic) addArmy(army *model.Army)  {
	posId := global.ToPosition(army.ToX, army.ToY)

	if _, ok := this.stopInPosArmys[posId]; ok == false {
		this.stopInPosArmys[posId] = make(map[int]*model.Army)
	}
	this.stopInPosArmys[posId][army.Id] = army
}


func (this*armyLogic) addAction(t int64, army *model.Army)  {
	_, ok := this.armyByEndTime[t]
	if ok == false {
		this.armyByEndTime[t] = make([]*model.Army, 0)
	}
	this.armyByEndTime[t] = append(this.armyByEndTime[t], army)
}

//把行动丢进来
func (this*armyLogic) PushAction(army *model.Army)  {
	this.timeMutex.Lock()
	defer this.timeMutex.Unlock()

	if  army.Cmd == model.ArmyCmdAttack ||
		army.Cmd == model.ArmyCmdDefend ||
		army.Cmd == model.ArmyCmdTransfer{
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

		if army.FromX == army.ToX && army.FromY == army.ToY {
			//处理调动到其他地方待命的情况，会归属的城池
			city, ok := mgr.RCMgr.Get(army.CityId)
			if ok {
				army.FromX = city.X
				army.FromY = city.Y

				//计算回去的时间
				//speed := mgr.AMgr.GetSpeed(army)
				//t := mgr.TravelTime(speed, army.FromX, army.FromY, army.ToX, army.ToY)
				army.Start = time.Now()
				//army.End = time.Now().Add(time.Duration(t) * time.Millisecond)
				army.End = time.Now().Add(40*time.Second)
			}

		}else{
			cur := time.Now()
			diff := army.End.Unix()-army.Start.Unix()
			if cur.Unix() < army.End.Unix(){
				diff = cur.Unix()-army.Start.Unix()
			}
			army.Start = cur
			army.End = cur.Add(time.Duration(diff) * time.Second)

		}
		army.Cmd = model.ArmyCmdBack
		this.addAction(army.End.Unix(), army)
	}

	this.Update(army)

}

func (this*armyLogic) ArmyBack(army *model.Army)  {
	army.State = model.ArmyRunning
	army.Cmd = model.ArmyCmdBack

	this.timeMutex.Lock()
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
	this.timeMutex.Unlock()
	this.PushAction(army)
}

func (this*armyLogic) Reclamation(army *model.Army)  {
	army.State = model.ArmyStop
	army.Cmd = model.ArmyCmdReclamation
	this.PushAction(army)
}



