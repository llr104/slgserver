package war

import (
	"slgserver/server/slgserver/static_conf"
	"slgserver/server/slgserver/static_conf/general"
	"slgserver/server/slgserver/static_conf/skill"
	"slgserver/util"
)

const maxRound = 10

type warResult struct {
	camp     *warCamp
	rounds   []*warRound
	curRound *warRound
	result   int //0失败，1平，2胜利
}

func newWarResult(camp *warCamp) *warResult {
	return &warResult{camp: camp}
}

func (this *warResult) battle() {
	cur := 0
	for true {
		r, isEnd := this.runRound()
		this.rounds = append(this.rounds, r)
		cur += 1
		if cur >= maxRound || isEnd {
			break
		}
	}

	for i := 0; i < static_conf.ArmyGCnt; i++ {
		if this.camp.attackPos[i] != nil {
			this.camp.attack.SoldierArray[i] = this.camp.attackPos[i].soldiers
		}
		if this.camp.defensePos[i] != nil {
			this.camp.defense.SoldierArray[i] = this.camp.defensePos[i].soldiers
		}
	}

	if this.camp.attackPos[0].soldiers == 0 {
		this.result = 0
	} else if this.camp.defensePos[0] != nil && this.camp.defensePos[0].soldiers != 0 {
		this.result = 1
	} else {
		this.result = 2
	}
}

//打击前技能
func (this *warResult) beforeSkill(att *armyPosition, our []*armyPosition, enemy []*armyPosition) []skillHit {
	beforeSkills := att.hitBefore()
	return this.acceptSkill(beforeSkills, att, our, enemy)
}

//打击后技能
func (this *warResult) afterSkill(att *armyPosition, our []*armyPosition, enemy []*armyPosition) []skillHit {
	afterSkills := att.hitAfter()
	return this.acceptSkill(afterSkills, att, our, enemy)
}

func (this *warResult) acceptSkill(skills []*attachSkill, att *armyPosition, our []*armyPosition, enemy []*armyPosition) []skillHit {
	ret := make([]skillHit, 0)
	for _, bs := range skills {

		cfg := bs.cfg
		sh := skillHit{Lv: bs.lv, CfgId: cfg.CfgId, FromId: att.general.Id}
		sh.IEffect = cfg.IncludeEffect
		sh.EValue = cfg.Levels[bs.lv-1].EffectValue
		sh.ERound = cfg.Levels[bs.lv-1].EffectRound
		
		switch skill.TargetType(bs.cfg.Target) {
		case skill.MySelf:
		case skill.OurSingle:
		case skill.OurMostTwo:
		case skill.OurMostThree:
		case skill.OurAll:
			att.acceptSkill(bs)
			sh.ToId = append(sh.ToId, att.general.Id)
			break
		case skill.EnemySingle:
		case skill.EnemyMostTwo:
		case skill.EnemyMostThree:
		case skill.EnemyAll:
			o, _ := this.camp.randArmyPosition(enemy)
			o.acceptSkill(bs)
			sh.ToId = append(sh.ToId, o.general.Id)
			break
		}
		ret = append(ret, sh)
	}
	return ret
}

//回合
func (this *warResult) runRound() (*warRound, bool) {

	this.curRound = &warRound{}
	attacks := this.camp.attackPos
	defenses := this.camp.defensePos

	//随机先手
	//n := rand.Intn(10)
	//if n%2 == 0 {
	//	attacks = this.camp.defensePos
	//	defenses = this.camp.attackPos
	//}

	for _, hitA := range attacks {

		////////攻击方begin//////////
		if hitA == nil || hitA.soldiers == 0 {
			continue
		}
		hitB, _ := this.camp.randArmyPosition(defenses)
		if hitB == nil {
			goto end
		}
		if this.hit(hitA, hitB, attacks, defenses) {
			goto end
		}
		////////攻击方end//////////

		////////防守方begin//////////
		if hitB.soldiers == 0 || hitA.soldiers == 0 {
			continue
		}
		if this.hit(hitB, hitA, defenses, attacks) {
			goto end
		}
		////////防守方end//////////
	}
	//清理过期的技能功能效果
	for _, attack := range attacks {
		attack.checkNextRound()
	}

	for _, defense := range defenses {
		defense.checkNextRound()
	}

	return this.curRound, false

end:
	return this.curRound, true
}

func (this *warResult) hit(hitA *armyPosition, hitB *armyPosition, attacks []*armyPosition, defenses []*armyPosition) bool {
	//释放技能
	h := hit{}
	h.ABeforeSkill = this.beforeSkill(hitA, attacks, defenses)
	realA := hitA.calRealBattleAttr()
	realB := hitB.calRealBattleAttr()

	attHarmRatio := general.GenArms.GetHarmRatio(hitA.arms, hitB.arms)
	attHarm := float64(util.AbsInt(realA.force-realB.defense)*hitA.soldiers) * attHarmRatio * 0.0005
	attKill := int(attHarm)
	attKill = util.MinInt(attKill, hitB.soldiers)
	hitB.soldiers -= attKill
	hitA.general.Exp += attKill * 5

	//清理瞬时技能
	hitA.checkHit()
	hitB.checkHit()

	//战报
	h.AId = hitA.general.Id
	h.DId = hitB.general.Id
	h.DLoss = attKill

	//大营干死了，直接结束
	if hitB.position == 0 && hitB.soldiers == 0 {
		this.curRound.Battle = append(this.curRound.Battle, h)
		return true
	} else {
		h.AAfterSkill = this.afterSkill(hitB, defenses, attacks)
		h.BAfterSkill = this.afterSkill(hitA, attacks, defenses)
		this.curRound.Battle = append(this.curRound.Battle, h)
		return false
	}
}

func (this *warResult) getRounds() []*warRound {
	return this.rounds
}
