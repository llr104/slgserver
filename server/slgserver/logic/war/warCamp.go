package war

import (
	"math/rand"
	"slgserver/server/slgserver/logic/mgr"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/static_conf/facility"
	"slgserver/server/slgserver/static_conf/skill"
)

type attachSkill struct {
	cfg      skill.Conf
	id       int
	lv       int
	duration int //剩余轮数
}

func newAttachSkill(cfg skill.Conf, id int, lv int) *attachSkill {
	return &attachSkill{cfg: cfg, id: id, lv: lv, duration: cfg.Duration}
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
		skillCfg, ok := skill.Skill.GetCfg(s.CfgId)
		if ok {
			if !skillCfg.IsHitBefore() {
				continue
			}
			l := skillCfg.Levels[s.Lv-1]
			b := rand.Intn(100)
			if b >= 100-l.Probability {
				as := newAttachSkill(skillCfg, s.Id, s.Lv)
				ret = append(ret, as)
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
		skillCfg, ok := skill.Skill.GetCfg(s.CfgId)
		if ok {
			if !skillCfg.IsHitAfter() {
				continue
			}
			l := skillCfg.Levels[s.Lv-1]
			b := rand.Intn(100)
			if b >= 100-l.Probability {
				as := newAttachSkill(skillCfg, s.Id, s.Lv)
				ret = append(ret, as)
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

type warCamp struct {
	attack     *model.Army
	defense    *model.Army
	attackPos  []*armyPosition
	defensePos []*armyPosition
}

func newCamp(attack *model.Army, defense *model.Army) *warCamp {
	c := &warCamp{attack: attack, defense: defense}
	c.init()
	return c
}

//初始化军队和武将属性、兵种、加成等
func (this *warCamp) init() {

	//城内设施加成
	attackAdds := []int{0, 0, 0, 0}
	if this.attack.CityId > 0 {
		attackAdds = mgr.RFMgr.GetAdditions(this.attack.CityId,
			facility.TypeForce,
			facility.TypeDefense,
			facility.TypeSpeed,
			facility.TypeStrategy)
	}

	defenseAdds := []int{0, 0, 0, 0}
	if this.defense.CityId > 0 {
		defenseAdds = mgr.RFMgr.GetAdditions(this.defense.CityId,
			facility.TypeForce,
			facility.TypeDefense,
			facility.TypeSpeed,
			facility.TypeStrategy)
	}

	//阵营加成
	aCampAdds := []int{0}
	aCamp := this.attack.GetCamp()
	if aCamp > 0 {
		aCampAdds = mgr.RFMgr.GetAdditions(this.attack.CityId, facility.TypeHanAddition-1+aCamp)
	}

	dCampAdds := []int{0}
	dCamp := this.attack.GetCamp()
	if dCamp > 0 {
		dCampAdds = mgr.RFMgr.GetAdditions(this.defense.CityId, facility.TypeHanAddition-1+aCamp)
	}

	this.attackPos = make([]*armyPosition, 0)
	this.defensePos = make([]*armyPosition, 0)

	for i, g := range this.attack.Gens {
		if g == nil {
			this.attackPos = append(this.attackPos, nil)
		} else {
			pos := &armyPosition{
				general:  g,
				soldiers: this.attack.SoldierArray[i],
				force:    g.GetForce() + attackAdds[0] + aCampAdds[0],
				defense:  g.GetDefense() + attackAdds[1] + aCampAdds[0],
				speed:    g.GetSpeed() + attackAdds[2] + aCampAdds[0],
				strategy: g.GetStrategy() + attackAdds[3] + aCampAdds[0],
				destroy:  g.GetDestroy() + aCampAdds[0],
				arms:     g.CurArms,
				position: i,
			}
			this.attackPos = append(this.attackPos, pos)
		}
	}

	for i, g := range this.defense.Gens {
		if g == nil {
			this.defensePos = append(this.defensePos, nil)
		} else {
			pos := &armyPosition{
				general:  g,
				soldiers: this.defense.SoldierArray[i],
				force:    g.GetForce() + defenseAdds[0] + dCampAdds[0],
				defense:  g.GetDefense() + defenseAdds[1] + dCampAdds[0],
				speed:    g.GetSpeed() + defenseAdds[2] + dCampAdds[0],
				strategy: g.GetStrategy() + defenseAdds[3] + dCampAdds[0],
				destroy:  g.GetDestroy() + dCampAdds[0],
				arms:     g.CurArms,
				position: i,
			}
			this.defensePos = append(this.defensePos, pos)
		}
	}

}

//随机一个目标位置
func (this *warCamp) randArmyPosition(pos []*armyPosition) (*armyPosition, int) {
	isEmpty := true
	for _, v := range pos {
		if v != nil && v.soldiers != 0 {
			isEmpty = false
			break
		}
	}

	if isEmpty {
		return nil, -1
	}

	for true {
		r := rand.Intn(100)
		index := r % len(pos)
		if pos[index] != nil && pos[index].soldiers != 0 {
			return pos[index], index
		}
	}

	return nil, -1
}
