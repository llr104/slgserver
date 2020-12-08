package logic

import (
	"encoding/json"
	"go.uber.org/zap"
	"slgserver/log"
	"slgserver/server/global"
	"slgserver/server/model"
	"slgserver/server/proto"
	"slgserver/server/static_conf/general"
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
	passby			sync.Mutex
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
		if IsCanArrive(army.ToX, army.ToY, army.RId){
			this.battle(army)
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

	this.passby.Lock()
	defer this.passby.Unlock()

	maxX := util.MinInt(global.MapWith, x+length-1)
	maxY := util.MinInt(global.MapHeight, y+length-1)

	out := make([]*model.Army, 0)
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

//简单战斗
func (this *armyLogic) battle(army *model.Army) {
	_, ok := RCMgr.PositionCity(army.ToX, army.ToY)
	if ok {
		//打城池
		AMgr.ArmyBack(army)
		return
	}

	this.executeBuild(army)
}

func (this* armyLogic) executeBuild(army *model.Army)  {
	roleBuid, _ := RBMgr.PositionBuild(army.ToX, army.ToY)

	posId := ToPosition(army.ToX, army.ToY)
	posArmys, isRoleEnemy := this.stopInPosArmys[posId]
	var enemys []*model.Army
	if isRoleEnemy == false {
		enemys = this.sys.GetArmy(army.ToX, army.ToY)
	}else{
		for _, v := range posArmys {
			enemys = append(enemys, v)
		}
	}

	warReports := make([]*model.WarReport, 0)
	var lastWar *WarResult = nil

	for _, enemy := range enemys {
		//战报处理
		pArmy := army.ToProto().(proto.Army)
		pEnemy := enemy.ToProto().(proto.Army)

		begArmy1, _ := json.Marshal(pArmy)
		begArmy2, _ := json.Marshal(pEnemy)

		//武将战斗前
		begGeneral1 := make([][]int, 0)
		for _, g := range army.Gens {
			if g != nil {
				pg := g.ToProto().(proto.General)
				begGeneral1 = append(begGeneral1, pg.ToArray())
			}
		}
		begGeneralData1, _ := json.Marshal(begGeneral1)

		begGeneral2 := make([][]int, 0)
		for _, g := range enemy.Gens {
			if g != nil {
				pg := g.ToProto().(proto.General)
				begGeneral2 = append(begGeneral2, pg.ToArray())
			}
		}
		begGeneralData2, _ := json.Marshal(begGeneral2)

		lastWar = NewWar(army, enemy)

		//武将战斗后
		endGeneral1 := make([][]int, 0)
		for _, g := range army.Gens {
			if g != nil {
				pg := g.ToProto().(proto.General)
				endGeneral1 = append(endGeneral1, pg.ToArray())
				level, exp := general.GenBasic.ExpToLevel(g.Exp)
				g.Level = level
				g.Exp = exp
				g.SyncExecute()
			}
		}
		endGeneralData1, _ := json.Marshal(endGeneral1)

		endGeneral2 := make([][]int, 0)
		for _, g := range enemy.Gens {
			if g != nil {
				pg := g.ToProto().(proto.General)
				endGeneral2 = append(endGeneral2, pg.ToArray())
				level, exp := general.GenBasic.ExpToLevel(g.Exp)
				g.Level = level
				g.Exp = exp
				g.SyncExecute()
			}
		}
		endGeneralData2, _ := json.Marshal(endGeneral2)

		pArmy = army.ToProto().(proto.Army)
		pEnemy = enemy.ToProto().(proto.Army)
		endArmy1, _ := json.Marshal(pArmy)
		endArmy2, _ := json.Marshal(pEnemy)

		rounds, _ := json.Marshal(lastWar.round)
		wr := &model.WarReport{X: army.ToX, Y: army.ToY, AttackRid: army.RId,
			AttackIsRead: false, DefenseIsRead: false, DefenseRid: enemy.RId,
			BegAttackArmy: string(begArmy1), BegDefenseArmy: string(begArmy2),
			EndAttackArmy: string(endArmy1), EndDefenseArmy: string(endArmy2),
			BegAttackGeneral: string(begGeneralData1),
			BegDefenseGeneral: string(begGeneralData2),
			EndAttackGeneral: string(endGeneralData1),
			EndDefenseGeneral: string(endGeneralData2),
			Rounds: string(rounds),
			Result: lastWar.result,
			CTime: time.Now(),
		}

		warReports = append(warReports, wr)
		enemy.ToSoldier()
		enemy.ToGeneral()

		if isRoleEnemy {
			if lastWar.result > 1 {
				if isRoleEnemy {
					delete(this.stopInPosArmys, posId)
				}
				AMgr.ArmyBack(enemy)
			}
			enemy.SyncExecute()
		}

	}
	army.SyncExecute()

	if lastWar.result > 1 {
		if roleBuid != nil {
			destory := GMgr.GetDestroy(army)
			wr := warReports[len(warReports)-1]
			wr.DestroyDurable = util.MinInt(destory, roleBuid.CurDurable)
			roleBuid.CurDurable = util.MaxInt(0, roleBuid.CurDurable - destory)
			if roleBuid.CurDurable == 0{
				//攻占了玩家的领地
				wr.Occupy = 1
				RBMgr.RemoveFromRole(roleBuid)
				RBMgr.AddBuild(army.RId, army.ToX, army.ToY)
				roleBuid.CurDurable = roleBuid.MaxDurable
				this.OccupyRoleBuild(army.RId, army.ToX, army.ToY)
			}else{
				wr.Occupy = 0
			}

		}else{
			//占领系统领地
			this.OccupySystemBuild(army.RId, army.ToX, army.ToY)
			wr := warReports[len(warReports)-1]
			wr.DestroyDurable = 100
			wr.Occupy = 1
			this.sys.DelArmy(army.ToX, army.ToY)
		}
	}

	//领地发生变化
	if newRoleBuild, ok := RBMgr.PositionBuild(army.ToX, army.ToY); ok {
		RoleBuildExtra(newRoleBuild)
		newRoleBuild.SyncExecute()
	}

	for _, wr := range warReports {
		wr.SyncExecute()
	}

}

func (this* armyLogic) OccupyRoleBuild(rid, x, y int)  {
	newId := rid

	if b, ok := RBMgr.PositionBuild(x, y); ok {

		oldId := b.RId
		log.DefaultLog.Info("battle in role build",
			zap.Int("oldRId", oldId),
			zap.Int("newRId", newId))

		//被占领的减产
		if oldRole, ok := RResMgr.Get(oldId); ok{
			oldRole.WoodYield -= b.Wood
			oldRole.GrainYield -= b.Grain
			oldRole.StoneYield -= b.Stone
			oldRole.IronYield -= b.Iron
			oldRole.SyncExecute()
		}
		//占领的增加产量
		if newRole, ok := RResMgr.Get(newId); ok{
			newRole.WoodYield += b.Wood
			newRole.GrainYield += b.Grain
			newRole.StoneYield += b.Stone
			newRole.IronYield += b.Iron
			newRole.SyncExecute()
		}
		b.RId = rid
	}
}

func (this* armyLogic) OccupySystemBuild(rid, x, y int)  {
	newId := rid

	if _, ok := RBMgr.PositionBuild(x, y); ok {
		return
	}

	if NMMgr.IsCanBuild(x, y){
		rb, ok := RBMgr.AddBuild(rid, x, y)
		if ok {
			//占领的增加产量
			if newRole, ok := RResMgr.Get(newId); ok{
				newRole.WoodYield += rb.Wood
				newRole.GrainYield += rb.Grain
				newRole.StoneYield += rb.Stone
				newRole.IronYield += rb.Iron
				newRole.SyncExecute()
			}
		}
	}
}