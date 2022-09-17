package war

import (
	"math/rand"

	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/static_conf/skill"
)

//真正的战斗属性
type realBattleAttr struct {
	force    int //武力
	strategy int //策略
	defense  int //防御
	speed    int //速度
	destroy  int //破坏
}

//战斗位置的属性
type armyPosition struct {
	general  *model.General
	soldiers int //兵力
	force    int //武力
	strategy int //策略
	defense  int //防御
	speed    int //速度
	destroy  int //破坏
	arms     int //兵种
	position int //位置

	skills []*attachSkill
}

//攻击前触发技能
func (this *armyPosition) hitBefore() []*attachSkill {
	ret := make([]*attachSkill, 0)

	skills := this.general.SkillsArray
	for _, s := range skills {
		if s == nil {
			continue
		}

		skillCfg, ok := skill.Skill.GetCfg(s.CfgId)
		if ok {
			if !skillCfg.IsHitBefore() {
				continue
			}
			if global.IsDev() {
				as := newAttachSkill(skillCfg, s.Id, s.Lv)
				ret = append(ret, as)
			} else {
				l := skillCfg.Levels[s.Lv-1]
				b := rand.Intn(100)
				if b >= 100-l.Probability {
					as := newAttachSkill(skillCfg, s.Id, s.Lv)
					ret = append(ret, as)
				}
			}
		}
	}
	return ret
}

//攻击后触发技能
func (this *armyPosition) hitAfter() []*attachSkill {
	ret := make([]*attachSkill, 0)

	skills := this.general.SkillsArray
	for _, s := range skills {
		if s == nil {
			continue
		}

		skillCfg, ok := skill.Skill.GetCfg(s.CfgId)
		if ok {
			if !skillCfg.IsHitAfter() {
				continue
			}
			if global.IsDev() {
				as := newAttachSkill(skillCfg, s.Id, s.Lv)
				ret = append(ret, as)
			} else {
				l := skillCfg.Levels[s.Lv-1]
				b := rand.Intn(100)
				if b >= 100-l.Probability {
					as := newAttachSkill(skillCfg, s.Id, s.Lv)
					ret = append(ret, as)
				}
			}
		}
	}
	return ret
}

func (this *armyPosition) acceptSkill(s *attachSkill) {
	if this.skills == nil {
		this.skills = make([]*attachSkill, 0)
	}

	this.skills = append(this.skills, s)
}

func (this *armyPosition) checkHit() {
	skills := make([]*attachSkill, 0)
	for _, skill := range this.skills {
		if skill.duration > 0 {
			//瞬时技能，当前攻击完成后移除
			skills = append(skills, skill)
		}
	}
	this.skills = skills
}

func (this *armyPosition) checkNextRound() {
	skills := make([]*attachSkill, 0)
	for _, skill := range this.skills {
		skill.duration -= 1
		if skill.duration <= 0 {
			//持续技能，当前回合结束后持续到期移除
			skills = append(skills, skill)
		}
	}
	this.skills = skills
}

//计算真正的战斗属性，包含了技能
func (this *armyPosition) calRealBattleAttr() realBattleAttr {
	attr := realBattleAttr{}
	attr.defense = this.defense
	attr.force = this.force
	attr.destroy = this.destroy
	attr.speed = this.speed
	attr.strategy = this.strategy

	for _, s := range this.skills {
		lvData := s.cfg.Levels[s.lv-1]
		effects := s.cfg.IncludeEffect

		for i, effect := range effects {
			v := lvData.EffectValue[i]
			switch skill.EffectType(effect) {
			case skill.HurtRate:
				break
			case skill.Force:
				attr.force += v
				break
			case skill.Defense:
				attr.defense += v
				break
			case skill.Strategy:
				attr.strategy += v
				break
			case skill.Speed:
				attr.speed += v
				break
			case skill.Destroy:
				attr.destroy += v
				break
			}
		}
	}
	return attr
}
