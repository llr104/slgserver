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
func (this *warResult) beforeSkill(att *armyPosition, our []*armyPosition, enemy []*armyPosition) []*skillHit {
	beforeSkills := att.hitBefore()
	return this.acceptSkill(beforeSkills, att, our, enemy)
}

//打击后技能
func (this *warResult) afterSkill(att *armyPosition, our []*armyPosition, enemy []*armyPosition) []*skillHit {
	afterSkills := att.hitAfter()
	return this.acceptSkill(afterSkills, att, our, enemy)
}

func (this *warResult) acceptSkill(skills []*attachSkill, att *armyPosition, our []*armyPosition, enemy []*armyPosition) []*skillHit {
	ret := make([]*skillHit, 0)
	for _, bs := range skills {

		cfg := bs.cfg
		sh := &skillHit{Lv: bs.lv, CfgId: cfg.CfgId, FromId: att.general.Id}
		sh.IEffect = cfg.IncludeEffect
		sh.EValue = cfg.Levels[bs.lv-1].EffectValue
		sh.ERound = cfg.Levels[bs.lv-1].EffectRound

		switch skill.TargetType(bs.cfg.Target) {
		case skill.MySelf:
			bs.isEnemy = false
			ps := []*armyPosition{att}
			this._acceptSkill_(ps, bs, sh)
			break
		case skill.OurSingle:
			bs.isEnemy = false
			s, _ := this.camp.randArmyPosition(our)
			ps := []*armyPosition{s}
			this._acceptSkill_(ps, bs, sh)
			break
		case skill.OurMostTwo:
			bs.isEnemy = false
			ps, _ := this.camp.randMostTwoArmyPosition(our)
			this._acceptSkill_(ps, bs, sh)
		case skill.OurMostThree:
			bs.isEnemy = false
			ps, _ := this.camp.randMostTwoArmyPosition(our)
			this._acceptSkill_(ps, bs, sh)
		case skill.OurAll:
			bs.isEnemy = false
			ps, _ := this.camp.allArmyPosition(our)
			this._acceptSkill_(ps, bs, sh)
			break
		case skill.EnemySingle:
			bs.isEnemy = true
			s, _ := this.camp.randArmyPosition(enemy)
			ps := []*armyPosition{s}
			this._acceptSkill_(ps, bs, sh)
		case skill.EnemyMostTwo:
			bs.isEnemy = true
			ps, _ := this.camp.randMostTwoArmyPosition(enemy)
			this._acceptSkill_(ps, bs, sh)
		case skill.EnemyMostThree:
			bs.isEnemy = true
			ps, _ := this.camp.randMostThreeArmyPosition(enemy)
			this._acceptSkill_(ps, bs, sh)
		case skill.EnemyAll:
			bs.isEnemy = true
			ps, _ := this.camp.allArmyPosition(enemy)
			this._acceptSkill_(ps, bs, sh)
			break
		}
		ret = append(ret, sh)
	}
	return ret
}

func (this *warResult) _acceptSkill_(ps []*armyPosition, skill *attachSkill, sh *skillHit) {
	if ps == nil {
		return
	}
	for _, p := range ps {
		p.acceptSkill(skill)
		sh.ToId = append(sh.ToId, p.general.Id)
	}
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

	//伤害技能
	for _, s := range h.ABeforeSkill {
		s.Kill = make([]int, len(s.ToId))
		for i, e := range s.IEffect {
			if skill.EffectType(e) == skill.HurtRate {
				v := s.EValue[i]
				for j, to := range s.ToId {
					hitB := this.camp.findByGiId(defenses, to)
					if hitB != nil && hitB.soldiers > 0 {
						realB := hitB.calRealBattleAttr()
						force := realA.force * v / 100
						attKill := this.kill(hitA, hitB, force, realB.defense)
						s.Kill[j] += attKill
					}
				}
			}
		}
	}

	//战报
	if hitB.soldiers > 0 {
		attKill := this.kill(hitA, hitB, realA.force, realB.defense)
		h.AId = hitA.general.Id
		h.DId = hitB.general.Id
		h.DLoss = attKill
	}

	//清理瞬时技能
	hitA.checkHit()
	hitB.checkHit()

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

func (this *warResult) kill(hitA *armyPosition, hitB *armyPosition, aForce int, bDefense int) int {
	attHarmRatio := general.GenArms.GetHarmRatio(hitA.arms, hitB.arms)
	attHarm := float64(util.AbsInt(aForce-bDefense)*hitA.soldiers) * attHarmRatio * 0.0005
	attKill := int(attHarm)
	attKill = util.MinInt(attKill, hitB.soldiers)
	hitB.soldiers -= attKill
	hitA.general.Exp += attKill * 5

	return attKill
}

func (this *warResult) getRounds() []*warRound {
	return this.rounds
}
