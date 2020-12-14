package logic

import (
	"slgserver/server/global"
	"slgserver/server/model"
	"slgserver/util"
	"sync"
	"time"
)


var ArmyLogic *armyLogic

func init() {
	ArmyLogic = &armyLogic{
		arriveArmys: make(chan *model.Army, 100),
		giveUpId:       make(chan int, 100),
		updateArmys:    make(chan *model.Army, 100),
		outArmys:		make(map[int]*model.Army),
		stopInPosArmys: make(map[int]map[int]*model.Army),
		passbyPosArmys: make(map[int]map[int]*model.Army),
		sys:            NewSysArmy()}

	go ArmyLogic.running()
}

type armyLogic struct {
	passby			sync.RWMutex
	sys            	*sysArmyLogic
	giveUpId       	chan int
	arriveArmys    	chan *model.Army
	updateArmys    	chan *model.Army
	outArmys		map[int]*model.Army
	stopInPosArmys 	map[int]map[int]*model.Army //玩家停留位置的军队 key:posId,armyId
	passbyPosArmys 	map[int]map[int]*model.Army //玩家路过位置的军队 key:posId,armyId
}

func (this *armyLogic) running(){
	passbyTimer := time.NewTicker(10 * time.Second)
	for {
		select {
			case <-passbyTimer.C:{
				this.passby.Lock()
				this.passbyPosArmys = make(map[int]map[int]*model.Army)
				for _, army := range this.outArmys {
					if army.State == model.ArmyRunning{
						x, y := army.Position()
						posId := ToPosition(x, y)
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
						AMgr.ArmyBack(army)
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
			IsCanDefend(army.ToX, army.ToY, army.RId) == false{
			newBattle(army)
		} else{
			war := NewEmptyWar(army)
			war.SyncExecute()
		}
		AMgr.ArmyBack(army)
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
			AMgr.ArmyBack(army)
		}

	}else if army.Cmd == model.ArmyCmdReclamation {
		if army.State == model.ArmyRunning {

			ok := RBMgr.BuildIsRId(army.ToX, army.ToY, army.RId)
			if ok  {
				//目前是自己的领地才能屯田
				this.addArmy(army)
				AMgr.Reclamation(army)
			}else{
				war := NewEmptyWar(army)
				war.SyncExecute()
				AMgr.ArmyBack(army)
			}

		}else {
			AMgr.ArmyBack(army)
			//增加场量
			rr, ok := RResMgr.Get(army.RId)
			if ok {
				b, ok1 := RBMgr.PositionBuild(army.ToX, army.ToY)
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
		this.Update(army)
	}
}

func (this *armyLogic) ScanBlock(x, y, length int) []*model.Army {

	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}

	maxX := util.MinInt(global.MapWith, x+length-1)
	maxY := util.MinInt(global.MapHeight, y+length-1)
	out := make([]*model.Army, 0)
	this.passby.RLock()
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {
			posId := ToPosition(i, j)
			armys, ok := this.passbyPosArmys[posId]
			if ok {
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
	posId := ToPosition(x, y)
	delete(this.stopInPosArmys, posId)
}

func (this* armyLogic) addArmy(army *model.Army)  {
	posId := ToPosition(army.ToX, army.ToY)

	if _, ok := this.stopInPosArmys[posId]; ok == false {
		this.stopInPosArmys[posId] = make(map[int]*model.Army)
	}
	this.stopInPosArmys[posId][army.Id] = army

}

