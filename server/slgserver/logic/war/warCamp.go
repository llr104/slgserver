package war

import (
	"math/rand"
	"slgserver/server/slgserver/logic/mgr"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/static_conf/facility"
)

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
