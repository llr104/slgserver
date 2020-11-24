package logic

import (
	"encoding/json"
	"go.uber.org/zap"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/server"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
	"slgserver/util"
	"time"
)


var ArmyLogic *armyLogic

func init() {
	ArmyLogic = &armyLogic{armys: make(chan *model.Army, 100),
		posArmys: make(map[int]map[int]*model.Army)}
	go ArmyLogic.running()
}

type armyLogic struct {
	armys    chan *model.Army
	posArmys map[int]map[int]*model.Army	//key:posId,armyId
}

func (this *armyLogic) running(){
	for {
		select {
		case army := <-this.armys:
			cur_t := time.Now().Unix()
			diff := army.End.Unix() - army.Start.Unix()
			if army.Cmd == model.ArmyCmdAttack {

				if cur_t >= 2*diff + army.Start.Unix() {
					//两倍路程
					army.Cmd = model.ArmyCmdIdle
					army.State = model.ArmyStop
					this.battle(army)
					AMgr.PushAction(army)
				}else if cur_t >= 1*diff + army.Start.Unix(){
					//一倍路程
					army.Cmd = model.ArmyCmdBack
					army.State = model.ArmyRunning
					army.Start = army.End
					army.End = army.Start.Add(time.Duration(diff)*time.Second)

					this.battle(army)
					AMgr.PushAction(army)
				}
			}else if army.Cmd == model.ArmyCmdDefend {
				//呆在哪里不动
				posId := ToPosition(army.ToX, army.ToY)
				if _, ok := this.posArmys[posId]; ok == false {
					this.posArmys[posId] = make(map[int]*model.Army)
				}
				this.posArmys[posId][army.Id] = army
				army.State = model.ArmyStop

			} else if army.Cmd == model.ArmyCmdBack {
				 if cur_t >= 1*diff + army.Start.Unix(){
					//一倍路程
					army.Cmd = model.ArmyCmdIdle
					army.State = model.ArmyStop
				}else{
					army.State = model.ArmyRunning
				}

				//如果该队伍在驻守，需要移除
				posId := ToPosition(army.ToX, army.ToY)
				if _, ok := this.posArmys[posId]; ok {
					delete(this.posArmys[posId], army.Id)
				}
			}

			army.DB.Sync()
			ap := &proto.ArmyStatePush{}
			ap.CityId = army.CityId
			model_to_proto.Army(army, &ap.Army)
			//通知部队变化了
			server.DefaultConnMgr.PushByRoleId(army.RId, "general.armyState", ap)
		}
	}
}

func (this *armyLogic) Arrive(army* model.Army) {
	this.armys <- army
}

//简单战斗
func (this *armyLogic) battle(army* model.Army) {
	_, ok := RCMgr.PositionCity(army.ToX, army.ToY)
	if ok {
		//打城池
		return
	}

	_, ok = RBMgr.PositionBuild(army.ToX, army.ToY)
	if ok {
		//打玩家占领的领地
		army.SoldierArray[0] = util.MaxInt(0, army.SoldierArray[0]-50)
		army.SoldierArray[1] = util.MaxInt(0, army.SoldierArray[1]-50)
		army.SoldierArray[2] = util.MaxInt(0, army.SoldierArray[2]-50)
		this.OccupyRoleBuild(army.RId, army.ToX, army.ToY)
	}else{

		enemys := SysArmy.GetArmy(army.ToX, army.ToY)
		warReports := make([]*model.WarReport, 0)
		for _, enemy := range enemys {
			//战报处理
			begArmy1, _ := json.Marshal(army)
			begArmy2, _ := json.Marshal(enemy)

			winCnt := 0
			for i, s1 := range army.SoldierArray {
				s2 := enemy.SoldierArray[i]
				enemy.SoldierArray[i] = util.MaxInt(0, enemy.SoldierArray[i]-s1)
				army.SoldierArray[i] = util.MaxInt(0, army.SoldierArray[i]-s2)
				if army.SoldierArray[i] > 0{
					winCnt+=1
				}
			}
			endArmy1, _ := json.Marshal(army)
			endArmy2, _ := json.Marshal(enemy)

			wr := &model.WarReport{X: army.ToX, Y: army.ToY, AttackRid: army.RId,
				AttackIsRead: false, DefenseIsRead: false, DefenseRid: enemy.RId,
				BegAttackArmy: string(begArmy1), BegDefenseArmy: string(begArmy2),
				EndAttackArmy: string(endArmy1), EndDefenseArmy: string(endArmy2),
				AttackIsWin: winCnt>=2,Time: time.Now(),
			}
			warReports = append(warReports, wr)
		}

		//三盘两胜
		isWinCnt := 0
		for _, s := range army.SoldierArray {
			if s>0{
				isWinCnt+=1
			}
		}
		if isWinCnt>=2 {
			//占领系统领地
			this.OccupySystemBuild(army.RId, army.ToX, army.ToY)
			wr := warReports[len(warReports)-1]
			wr.DestroyDurable = 100
			wr.Occupy = 1
		}

		push := &proto.WarReportPush{}
		push.List = make([]proto.WarReport, len(warReports))
		for i, wr := range warReports {
			db.MasterDB.InsertOne(wr)
			model_to_proto.WarReport(wr, &push.List[i])
			server.DefaultConnMgr.PushByRoleId(army.RId, "war.reportPush", push)
		}
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
			oldRole.DB.Sync()
		}
		//占领的增加产量
		if newRole, ok := RResMgr.Get(newId); ok{
			newRole.WoodYield += b.Wood
			newRole.GrainYield += b.Grain
			newRole.StoneYield += b.Stone
			newRole.IronYield += b.Iron
			newRole.DB.Sync()
		}
		b.DB.Sync()
		b.RId = rid

		push := &proto.RoleBuildStatePush{}
		model_to_proto.MRBuild(b, &push.MRBuild)
		server.DefaultConnMgr.PushByRoleId(oldId, "role.roleBuildState", push)
		server.DefaultConnMgr.PushByRoleId(newId, "role.roleBuildState", push)
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
				newRole.DB.Sync()
			}
			push := &proto.RoleBuildStatePush{}
			model_to_proto.MRBuild(rb, &push.MRBuild)
			server.DefaultConnMgr.PushByRoleId(newId, "role.roleBuildState", push)
		}
	}
}