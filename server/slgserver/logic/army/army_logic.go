package army

import (
	"sync"
	"time"

	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/logic/check"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/logic/war"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/util"
)

var _armyLogic *ArmyLogic = nil

func Instance() *ArmyLogic {
	if _armyLogic == nil {
		_armyLogic = newArmyLogic()
	}
	return _armyLogic
}

func newArmyLogic() *ArmyLogic {
	a := &ArmyLogic{
		arriveArmys:    make(chan *model.Army, 100),
		interruptId:    make(chan int, 100),
		giveUpId:       make(chan int, 100),
		updateArmys:    make(chan *model.Army, 100),
		outArmys:       make(map[int]*model.Army),
		endTimeArmys:   make(map[int64][]*model.Army),
		stopInPosArmys: make(map[int]map[int]*model.Army),
		passByPosArmys: make(map[int]map[int]*model.Army),
		sys:            NewSysArmy(),
	}
	a.init()

	go a.check()
	go a.running()

	return a
}

type ArmyLogic struct {
	sys     *sysArmyLogic
	passBy  sync.RWMutex
	stop    sync.RWMutex
	out     sync.RWMutex
	endTime sync.RWMutex

	interruptId chan int
	giveUpId    chan int
	arriveArmys chan *model.Army
	updateArmys chan *model.Army

	outArmys       map[int]*model.Army         //城外的军队
	endTimeArmys   map[int64][]*model.Army     //key:到达时间
	stopInPosArmys map[int]map[int]*model.Army //玩家停留位置的军队 key:posId,armyId
	passByPosArmys map[int]map[int]*model.Army //玩家路过位置的军队 key:posId,armyId
}

func (this *ArmyLogic) init() {

	armys := mgr.AMgr.All()
	for _, army := range armys {
		//恢复已经执行行动的军队
		if army.Cmd != model.ArmyCmdIdle {
			e := army.End.Unix()
			_, ok := this.endTimeArmys[e]
			if ok == false {
				this.endTimeArmys[e] = make([]*model.Army, 0)
			}
			this.endTimeArmys[e] = append(this.endTimeArmys[e], army)
		}
	}

	curTime := time.Now().Unix()
	for kT, armys := range this.endTimeArmys {
		if kT <= curTime {
			for _, a := range armys {
				if a.Cmd == model.ArmyCmdAttack {
					this.Arrive(a)
				} else if a.Cmd == model.ArmyCmdDefend {
					this.Arrive(a)
				} else if a.Cmd == model.ArmyCmdBack {
					if curTime >= a.End.Unix() {
						a.ToX = a.FromX
						a.ToY = a.FromY
						a.Cmd = model.ArmyCmdIdle
						a.State = model.ArmyStop
					}
				}
				a.SyncExecute()
			}
			delete(this.endTimeArmys, kT)
		} else {
			for _, a := range armys {
				a.State = model.ArmyRunning
			}
		}
	}
}

func (this *ArmyLogic) check() {
	for true {
		t := time.Now().Unix()
		time.Sleep(100 * time.Millisecond)

		this.endTime.Lock()
		for k, armies := range this.endTimeArmys {
			if k <= t {
				for _, army := range armies {
					this.Arrive(army)
				}
				delete(this.endTimeArmys, k)
			}
		}
		this.endTime.Unlock()
	}
}

func (this *ArmyLogic) running() {
	passbyTimer := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-passbyTimer.C:
			{
				this.updatePassBy()
			}
		case army := <-this.updateArmys:
			{
				this.exeUpdate(army)
			}
		case army := <-this.arriveArmys:
			{
				this.exeArrive(army)
			}
		case giveId := <-this.giveUpId:
			{
				//在该位置驻守、调动的都需要返回
				this.stop.RLock()
				armys, ok := this.stopInPosArmys[giveId]
				this.stop.RUnlock()

				if ok {
					for _, army := range armys {
						this.ArmyBack(army)
					}
					this.DeleteStopArmy(giveId)
				}
			}
		case interruptId := <-this.interruptId:
			{
				//只有调动到该位置的军队需要返回
				var targets []*model.Army
				this.stop.Lock()
				armys, ok := this.stopInPosArmys[interruptId]
				if ok {
					for key, army := range armys {
						if army.FromX == army.ToX && army.FromY == army.ToY {
							targets = append(targets, army)
							delete(armys, key)
						}
					}
				}
				this.stop.Unlock()

				for _, target := range targets {
					this.ArmyBack(target)
				}
			}
		}
	}
}

func (this *ArmyLogic) updatePassBy() {

	temp := make(map[int]map[int]*model.Army)
	this.out.RLock()
	for _, army := range this.outArmys {
		if army.State == model.ArmyRunning {
			x, y := army.Position()
			posId := global.ToPosition(x, y)
			if _, ok := temp[posId]; ok == false {
				temp[posId] = make(map[int]*model.Army)
			}
			temp[posId][army.Id] = army
			army.CheckSyncCell()
		}
	}
	this.out.RUnlock()

	this.stop.RLock()
	for posId, armys := range this.stopInPosArmys {
		for _, army := range armys {
			if _, ok := temp[posId]; ok == false {
				temp[posId] = make(map[int]*model.Army)
			}
			temp[posId][army.Id] = army
		}
	}
	this.stop.RUnlock()

	this.passBy.Lock()
	this.passByPosArmys = temp
	this.passBy.Unlock()
}

func (this *ArmyLogic) exeUpdate(army *model.Army) {
	army.SyncExecute()
	if army.Cmd == model.ArmyCmdBack {
		this.stop.Lock()
		posId := global.ToPosition(army.ToX, army.ToY)
		armys, ok := this.stopInPosArmys[posId]
		if ok {
			delete(armys, army.Id)
			this.stopInPosArmys[posId] = armys
		}
		this.stop.Unlock()
	}

	this.out.Lock()
	if army.Cmd != model.ArmyCmdIdle {
		this.outArmys[army.Id] = army
	} else {
		delete(this.outArmys, army.RId)
	}
	this.out.Unlock()
}

func (this *ArmyLogic) exeArrive(army *model.Army) {
	if army.Cmd == model.ArmyCmdAttack {
		if check.IsCanArrive(army.ToX, army.ToY, army.RId) &&
			check.IsWarFree(army.ToX, army.ToY) == false &&
			check.IsCanDefend(army.ToX, army.ToY, army.RId) == false {
			war.NewBattle(army, this)
		} else {
			emptyWar := war.NewEmptyWar(army)
			emptyWar.SyncExecute()
		}
		this.ArmyBack(army)
	} else if army.Cmd == model.ArmyCmdDefend {
		//呆在哪里不动
		ok := check.IsCanDefend(army.ToX, army.ToY, army.RId)
		if ok {
			//目前是自己的领地才能驻守
			army.State = model.ArmyStop
			this.addStopArmy(army)
			this.Update(army)
		} else {
			emptyWar := war.NewEmptyWar(army)
			emptyWar.SyncExecute()
			this.ArmyBack(army)
		}

	} else if army.Cmd == model.ArmyCmdReclamation {
		if army.State == model.ArmyRunning {

			ok := mgr.RBMgr.BuildIsRId(army.ToX, army.ToY, army.RId)
			if ok {
				//目前是自己的领地才能屯田
				this.addStopArmy(army)
				this.Reclamation(army)
			} else {
				emptyWar := war.NewEmptyWar(army)
				emptyWar.SyncExecute()
				this.ArmyBack(army)
			}

		} else {
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
	} else if army.Cmd == model.ArmyCmdBack {
		army.State = model.ArmyStop
		army.Cmd = model.ArmyCmdIdle
		army.ToX = army.FromX
		army.ToY = army.FromY

		this.Update(army)
	} else if army.Cmd == model.ArmyCmdTransfer {
		//调动到位置了
		if army.State == model.ArmyRunning {

			ok := mgr.RBMgr.BuildIsRId(army.ToX, army.ToY, army.RId)
			if ok == false {
				this.ArmyBack(army)
			} else {
				b, _ := mgr.RBMgr.PositionBuild(army.ToX, army.ToY)
				if b.IsHasTransferAuth() {
					army.State = model.ArmyStop
					army.Cmd = model.ArmyCmdIdle
					x := army.ToX
					y := army.ToY
					army.FromX = x
					army.FromY = y
					army.ToX = x
					army.ToY = y
					this.addStopArmy(army)
					this.Update(army)
				} else {
					this.ArmyBack(army)
				}
			}
		}
	}
}

func (this *ArmyLogic) ScanBlock(rid, x, y, length int) []*model.Army {

	if x < 0 || x >= global.MapWith || y < 0 || y >= global.MapHeight {
		return nil
	}

	maxX := util.MinInt(global.MapWith, x+length-1)
	maxY := util.MinInt(global.MapHeight, y+length-1)
	out := make([]*model.Army, 0)

	this.passBy.RLock()
	for i := x; i <= maxX; i++ {
		for j := y; j <= maxY; j++ {

			posId := global.ToPosition(i, j)
			armys, ok := this.passByPosArmys[posId]
			if ok {
				is := ArmyIsInView(rid, i, j)
				if is == false {
					continue
				}
				for _, army := range armys {
					out = append(out, army)
				}
			}
		}
	}
	this.passBy.RUnlock()
	return out
}

func (this *ArmyLogic) Arrive(army *model.Army) {
	this.arriveArmys <- army
}

func (this *ArmyLogic) Update(army *model.Army) {
	this.updateArmys <- army
}

func (this *ArmyLogic) Interrupt(posId int) {
	this.interruptId <- posId
}

func (this *ArmyLogic) GiveUp(posId int) {
	this.giveUpId <- posId
}

func (this *ArmyLogic) GetStopArmys(posId int) []*model.Army {
	ret := make([]*model.Army, 0)
	this.stop.RLock()
	armys, ok := this.stopInPosArmys[posId]
	if ok {
		for _, army := range armys {
			ret = append(ret, army)
		}
	}
	this.stop.RUnlock()
	return ret
}

func (this *ArmyLogic) DeleteStopArmy(posId int) {
	this.stop.Lock()
	delete(this.stopInPosArmys, posId)
	this.stop.Unlock()
}

func (this *ArmyLogic) addStopArmy(army *model.Army) {
	posId := global.ToPosition(army.ToX, army.ToY)

	this.stop.Lock()
	if _, ok := this.stopInPosArmys[posId]; ok == false {
		this.stopInPosArmys[posId] = make(map[int]*model.Army)
	}
	this.stopInPosArmys[posId][army.Id] = army
	this.stop.Unlock()
}

func (this *ArmyLogic) addAction(t int64, army *model.Army) {
	this.endTime.Lock()
	defer this.endTime.Unlock()
	_, ok := this.endTimeArmys[t]
	if ok == false {
		this.endTimeArmys[t] = make([]*model.Army, 0)
	}
	this.endTimeArmys[t] = append(this.endTimeArmys[t], army)
}

//把行动丢进来
func (this *ArmyLogic) PushAction(army *model.Army) {

	if army.Cmd == model.ArmyCmdAttack ||
		army.Cmd == model.ArmyCmdDefend ||
		army.Cmd == model.ArmyCmdTransfer {
		t := army.End.Unix()
		this.addAction(t, army)

	} else if army.Cmd == model.ArmyCmdReclamation {
		if army.State == model.ArmyRunning {
			t := army.End.Unix()
			this.addAction(t, army)
		} else {
			costTime := static_conf.Basic.General.ReclamationTime
			t := army.End.Unix() + int64(costTime)
			this.addAction(t, army)
		}
	} else if army.Cmd == model.ArmyCmdBack {

		if army.FromX == army.ToX && army.FromY == army.ToY {
			//处理调动到其他地方待命的情况，会归属的城池
			city, ok := mgr.RCMgr.Get(army.CityId)
			if ok {
				army.FromX = city.X
				army.FromY = city.Y

				//计算回去的时间
				if global.IsDev() {
					army.Start = time.Now()
					army.End = time.Now().Add(40 * time.Second)
				} else {
					speed := mgr.AMgr.GetSpeed(army)
					t := mgr.TravelTime(speed, army.FromX, army.FromY, army.ToX, army.ToY)
					army.Start = time.Now()
					army.End = time.Now().Add(time.Duration(t) * time.Millisecond)
				}
			}

		} else {
			cur := time.Now()
			diff := army.End.Unix() - army.Start.Unix()
			if cur.Unix() < army.End.Unix() {
				diff = cur.Unix() - army.Start.Unix()
			}
			army.Start = cur
			army.End = cur.Add(time.Duration(diff) * time.Second)

		}
		army.Cmd = model.ArmyCmdBack
		this.addAction(army.End.Unix(), army)
	}

	this.Update(army)

}

func (this *ArmyLogic) ArmyBack(army *model.Army) {
	army.ClearConscript()

	army.State = model.ArmyRunning
	army.Cmd = model.ArmyCmdBack

	this.endTime.Lock()
	t := army.End.Unix()
	if actions, ok := this.endTimeArmys[t]; ok {
		for i, v := range actions {
			if v.Id == army.Id {
				actions = append(actions[:i], actions[i+1:]...)
				this.endTimeArmys[t] = actions
				break
			}
		}
	}
	this.endTime.Unlock()
	this.PushAction(army)
}

func (this *ArmyLogic) Reclamation(army *model.Army) {
	army.State = model.ArmyStop
	army.Cmd = model.ArmyCmdReclamation
	this.PushAction(army)
}

func (this *ArmyLogic) GetSysArmy(x, y int) []*model.Army {
	return this.sys.GetArmy(x, y)
}

func (this *ArmyLogic) DelSysArmy(x, y int) {
	this.sys.DelArmy(x, y)
}
