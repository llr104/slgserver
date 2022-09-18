package war

import (
	"encoding/json"
	"time"

	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/server/slgserver/ILogic"
	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/logic/union"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/server/slgserver/static_conf/general"
	"github.com/llr104/slgserver/util"
	"go.uber.org/zap"
)

type hit struct {
	AId          int         `json:"a_id"`   //本回合发起攻击的武将id
	DId          int         `json:"d_id"`   //本回合防御方的武将id
	DLoss        int         `json:"d_loss"` //本回合防守方损失的兵力
	ABeforeSkill []*skillHit `json:"a_bs"`   //攻击方攻击前技能
	AAfterSkill  []*skillHit `json:"a_as"`   //攻击方攻击后技能
	BAfterSkill  []*skillHit `json:"d_as"`   //防守方被攻击后触发技能
}

type skillHit struct {
	FromId  int   `json:"f_id"` //发起的id
	ToId    []int `json:"t_id"` //作用目标id
	CfgId   int   `json:"c_id"` //技能配置id
	Lv      int   `json:"lv"`   //技能等级
	IEffect []int `json:"i_e"`  //技能包括的效果
	EValue  []int `json:"e_v"`  //效果值
	ERound  []int `json:"e_r"`  //效果持续回合数
	Kill    []int `json:"kill"` //技能杀死数量
}

type warRound struct {
	Battle []hit `json:"b"`
}

func NewWar(attack *model.Army, defense *model.Army) *warResult {

	c := newCamp(attack, defense)

	result := newWarResult(c)
	result.battle()

	return result
}

func NewEmptyWar(attack *model.Army) *model.WarReport {
	//战报处理
	pArmy := attack.ToProto().(proto.Army)
	begArmy, _ := json.Marshal(pArmy)

	//武将战斗前
	begGeneral := make([][]int, 0)
	for _, g := range attack.Gens {
		if g != nil {
			pg := g.ToProto().(proto.General)
			begGeneral = append(begGeneral, pg.ToArray())
		}
	}
	begGeneralData, _ := json.Marshal(begGeneral)

	wr := &model.WarReport{X: attack.ToX, Y: attack.ToY, AttackRid: attack.RId,
		AttackIsRead: false, DefenseIsRead: true, DefenseRid: 0,
		BegAttackArmy: string(begArmy), BegDefenseArmy: "",
		EndAttackArmy: string(begArmy), EndDefenseArmy: "",
		BegAttackGeneral:  string(begGeneralData),
		EndAttackGeneral:  string(begGeneralData),
		BegDefenseGeneral: "",
		EndDefenseGeneral: "",
		Rounds:            "",
		Result:            0,
		CTime:             time.Now(),
	}
	return wr
}

//简单战斗
func NewBattle(attackArmy *model.Army, armyLogic ILogic.IArmyLogic) *Battle {
	battle := Battle{
		attackArmy: attackArmy,
		IArmyLogic: armyLogic,
	}
	battle.run()
	return &battle
}

type Battle struct {
	IArmyLogic ILogic.IArmyLogic
	attackArmy *model.Army
}

func (this *Battle) run() {
	city, ok := mgr.RCMgr.PositionCity(this.attackArmy.ToX, this.attackArmy.ToY)
	if ok {

		//驻守队伍被打
		posId := global.ToPosition(this.attackArmy.ToX, this.attackArmy.ToY)
		enemys := this.IArmyLogic.GetStopArmys(posId)

		//城内空闲的队伍被打
		if armys, ok := mgr.AMgr.GetByCity(city.CityId); ok {
			for _, enemy := range armys {
				if enemy.IsCanOutWar() {
					enemys = append(enemys, enemy)
				}
			}
		}

		if len(enemys) == 0 {
			//没有队伍
			destory := mgr.GMgr.GetDestroy(this.attackArmy)
			city.DurableChange(-destory)
			city.SyncExecute()

			wr := NewEmptyWar(this.attackArmy)
			wr.Result = 2
			wr.DefenseRid = city.RId
			wr.DefenseIsRead = false
			this.checkCityOccupy(wr, this.attackArmy, city)
			wr.SyncExecute()
		} else {
			lastWar, warReports := this.trigger(this.attackArmy, enemys, true)
			if lastWar.result > 1 {
				wr := warReports[len(warReports)-1]
				this.checkCityOccupy(wr, this.attackArmy, city)
			}
			for _, wr := range warReports {
				wr.SyncExecute()
			}
		}
	} else {
		//打建筑
		this.executeBuild(this.attackArmy)
	}
}

func (this *Battle) checkCityOccupy(wr *model.WarReport, attackArmy *model.Army, city *model.MapRoleCity) {
	destory := mgr.GMgr.GetDestroy(attackArmy)
	wr.DestroyDurable = util.MinInt(destory, city.CurDurable)
	city.DurableChange(-destory)

	if city.CurDurable == 0 {
		aAttr, _ := mgr.RAttrMgr.Get(attackArmy.RId)
		if aAttr.UnionId != 0 {
			//有联盟才能俘虏玩家
			wr.Occupy = 1
			dAttr, _ := mgr.RAttrMgr.Get(city.RId)
			dAttr.ParentId = aAttr.UnionId
			union.Instance().PutChild(aAttr.UnionId, city.RId)
			dAttr.SyncExecute()
			city.OccupyTime = time.Now()
		} else {
			wr.Occupy = 0
		}
	} else {
		wr.Occupy = 0
	}
	city.SyncExecute()
}

func (this *Battle) trigger(army *model.Army, enemys []*model.Army, isRoleEnemy bool) (*warResult, []*model.WarReport) {

	posId := global.ToPosition(army.ToX, army.ToY)
	warReports := make([]*model.WarReport, 0)
	var lastWar *warResult = nil

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

		rounds, _ := json.Marshal(lastWar.getRounds())
		wr := &model.WarReport{X: army.ToX, Y: army.ToY, AttackRid: army.RId,
			AttackIsRead: false, DefenseIsRead: false, DefenseRid: enemy.RId,
			BegAttackArmy: string(begArmy1), BegDefenseArmy: string(begArmy2),
			EndAttackArmy: string(endArmy1), EndDefenseArmy: string(endArmy2),
			BegAttackGeneral:  string(begGeneralData1),
			BegDefenseGeneral: string(begGeneralData2),
			EndAttackGeneral:  string(endGeneralData1),
			EndDefenseGeneral: string(endGeneralData2),
			Rounds:            string(rounds),
			Result:            lastWar.result,
			CTime:             time.Now(),
		}

		warReports = append(warReports, wr)
		enemy.ToSoldier()
		enemy.ToGeneral()

		if isRoleEnemy {
			if lastWar.result > 1 {
				if isRoleEnemy {
					this.IArmyLogic.DeleteStopArmy(posId)
				}
				this.IArmyLogic.ArmyBack(enemy)
			}
			enemy.SyncExecute()
		} else {
			wr.DefenseIsRead = true
		}
	}
	army.SyncExecute()
	return lastWar, warReports
}

func (this *Battle) executeBuild(army *model.Army) {
	roleBuild, _ := mgr.RBMgr.PositionBuild(army.ToX, army.ToY)

	posId := global.ToPosition(army.ToX, army.ToY)
	posArmys := this.IArmyLogic.GetStopArmys(posId)
	isRoleEnemy := len(posArmys) != 0
	var enemys []*model.Army
	if isRoleEnemy == false {
		enemys = this.IArmyLogic.GetSysArmy(army.ToX, army.ToY)
	} else {
		for _, v := range posArmys {
			enemys = append(enemys, v)
		}
	}

	lastWar, warReports := this.trigger(army, enemys, isRoleEnemy)

	if lastWar.result > 1 {
		if roleBuild != nil && roleBuild.RId > 0 {
			destory := mgr.GMgr.GetDestroy(army)
			wr := warReports[len(warReports)-1]
			wr.DestroyDurable = util.MinInt(destory, roleBuild.CurDurable)
			roleBuild.CurDurable = util.MaxInt(0, roleBuild.CurDurable-destory)
			if roleBuild.CurDurable == 0 {
				//攻占了玩家的领地
				bLimit := static_conf.Basic.Role.BuildLimit
				if bLimit > mgr.RBMgr.BuildCnt(army.RId) {
					wr.Occupy = 1
					mgr.RBMgr.RemoveFromRole(roleBuild)
					mgr.RBMgr.AddBuild(army.RId, army.ToX, army.ToY)
					this.OccupyRoleBuild(army.RId, army.ToX, army.ToY)
				} else {
					wr.Occupy = 0
				}
			} else {
				wr.Occupy = 0
			}

		} else {
			//占领系统领地
			wr := warReports[len(warReports)-1]
			blimit := static_conf.Basic.Role.BuildLimit
			if blimit > mgr.RBMgr.BuildCnt(army.RId) {
				this.OccupySystemBuild(army.RId, army.ToX, army.ToY)
				wr.DestroyDurable = 10000
				wr.Occupy = 1
			} else {
				wr.Occupy = 0
			}
			this.IArmyLogic.DelSysArmy(army.ToX, army.ToY)
		}
	}

	//领地发生变化
	if newRoleBuild, ok := mgr.RBMgr.PositionBuild(army.ToX, army.ToY); ok {
		newRoleBuild.SyncExecute()
	}

	for _, wr := range warReports {
		wr.SyncExecute()
	}
}

func (this *Battle) OccupyRoleBuild(rid, x, y int) {
	newId := rid

	if b, ok := mgr.RBMgr.PositionBuild(x, y); ok {

		b.CurDurable = b.MaxDurable
		b.OccupyTime = time.Now()

		oldId := b.RId
		log.DefaultLog.Info("hit in role build",
			zap.Int("oldRId", oldId),
			zap.Int("newRId", newId))
		b.RId = rid
	}
}

func (this *Battle) OccupySystemBuild(rid, x, y int) {

	if r, _ := mgr.RBMgr.PositionBuild(x, y); r != nil && r.RId > 0 {
		return
	}

	if mgr.NMMgr.IsCanBuild(x, y) {
		rb, ok := mgr.RBMgr.AddBuild(rid, x, y)
		if ok {
			rb.OccupyTime = time.Now()
		}
	}
}
