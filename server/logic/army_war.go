package logic

import (
	"fmt"
	"math/rand"
	"slgserver/server/model"
	"slgserver/server/static_conf/facility"
	"slgserver/util"
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

const maxRound = 10

type armyWar struct {
	attack *model.Army
	defense *model.Army
	attackPos []*armyPosition
	defensePos []*armyPosition
}

type battle struct {
	AId   int `json:"a_id"`   //本回合发起攻击的武将id
	DId   int `json:"d_id"`   //本回合防御方的武将id
	ALoss int `json:"a_loss"` //本回合攻击方损失的兵力
	DLoss int `json:"d_loss"` //本回合防守方损失的兵力
}

func (this* battle) to() []int{
	r := make([]int, 0)
	r = append(r, this.AId)
	r = append(r, this.DId)
	r = append(r, this.ALoss)
	r = append(r, this.DLoss)
	return r
}

type warRound struct {
	Battle	[][]int	`json:"b"`
}

type WarResult struct {
	round 	[]*warRound
	result	int			//0失败，1平，2胜利
}

func NewWar(attack *model.Army, defense *model.Army) *WarResult {

	w := armyWar{attack: attack, defense: defense}
	w.init()
	wars := w.battle()

	result := &WarResult{round: wars}
	if w.attackPos[0].soldiers == 0{
		result.result = 0
	}else if w.defensePos[0].soldiers != 0{
		result.result = 1
	}else{
		result.result = 2
	}

	return result
}

//初始化军队和武将属性、兵种、加成等
func (this* armyWar) init() {

	//城内设施加成
	attackAdds := []int{0,0,0,0}
	if this.attack.CityId > 0{
		attackAdds = RFMgr.GetAdditions(this.attack.CityId,
			facility.TypeForce,
			facility.TypeDefense,
			facility.TypeSpeed,
			facility.TypeStrategy)
	}

	defenseAdds := []int{0,0,0,0}
	if this.defense.CityId > 0{
		defenseAdds = RFMgr.GetAdditions(this.defense.CityId,
			facility.TypeForce,
			facility.TypeDefense,
			facility.TypeSpeed,
			facility.TypeStrategy)
	}

	//阵营加成
	aCampAdds := []int{0}
	aCamp := this.attack.GetCamp()
	if aCamp > 0{
		aCampAdds = RFMgr.GetAdditions(this.defense.CityId, facility.TypeHanAddition-1+aCamp)
	}

	dCampAdds := []int{0}
	dCamp := this.attack.GetCamp()
	if dCamp > 0 {
		dCampAdds = RFMgr.GetAdditions(this.defense.CityId, facility.TypeHanAddition-1+aCamp)
	}

	this.attackPos = make([]*armyPosition, 0)
	this.defensePos = make([]*armyPosition, 0)

	for i, g := range this.attack.Gens {
		if g == nil {
			this.attackPos = append(this.attackPos, nil)
		}else{
			pos := &armyPosition{
				general:  g,
				soldiers: this.attack.SoldierArray[i],
				force:    g.GetForce()  + attackAdds[0] + aCampAdds[0],
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
		}else{
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



	fmt.Println(this.defensePos)
}

func (this* armyWar) battle() []*warRound{
	rounds := make([]*warRound, 0)
	cur := 0
	for true{
		r, isEnd := this.round()
		rounds = append(rounds, r)
		cur += 1
		if cur >= maxRound || isEnd{
			break
		}
	}
	return rounds
}

//回合
func (this* armyWar) round() (*warRound, bool) {

	war := &warRound{}
	n := rand.Intn(10)
	attack := this.attackPos
	defense := this.defensePos
	attackArmy := this.attack
	defenseArmy := this.defense

	isEnd := false
	//随机先手
	if n % 2 == 0{
		attack = this.defensePos
		defense = this.attackPos

		attackArmy = this.defense
		defenseArmy = this.attack
	}

	//攻击方回合
	for _, posAttack := range attack {
		if posAttack == nil || posAttack.soldiers == 0{
			continue
		}
		//计算
		posDefense, index := this.randArmyPosition(defense)
		if posDefense == nil{
			continue
		}

		hurm := posAttack.soldiers *posAttack.force /1000
		def := posDefense.soldiers *posDefense.defense /1000

		kill := hurm-def
		if kill > 0{
			kill = util.MinInt(kill, posDefense.soldiers)
			posDefense.soldiers -= kill
			defenseArmy.SoldierArray[index] -= kill
			posAttack.general.Exp += kill*5
		}else{
			kill = 0
		}

		b := battle{AId: posAttack.general.Id, ALoss: 0, DId: posDefense.general.Id, DLoss: kill}
		war.Battle = append(war.Battle, b.to())

		//大营干死了，直接结束
		if posDefense.position == 0 && posDefense.soldiers == 0 {
			isEnd = true
			goto end
		}
	}

	//防守方回合
	for _, posAttack := range defense {
		if posAttack == nil || posAttack.soldiers == 0{
			continue
		}

		//计算
		posDefense, index := this.randArmyPosition(attack)
		hurm := posAttack.soldiers *posAttack.force /10000
		def := posDefense.soldiers *posDefense.defense /10000

		if posDefense == nil{
			continue
		}

		kill := hurm-def
		if kill > 0{
			kill = util.MinInt(kill, posDefense.soldiers)
			posDefense.soldiers -= kill
			attackArmy.SoldierArray[index] -= kill
			posAttack.general.Exp += kill*10
		}else{
			kill = 0
		}

		b := battle{AId: posAttack.general.Id, ALoss: 0, DId: posDefense.general.Id, DLoss: kill}
		war.Battle = append(war.Battle, b.to())

		//大营干死了，直接结束
		if posDefense.position == 0 && posDefense.soldiers == 0 {
			isEnd = true
			goto end
		}
	}

	end:
	return war, isEnd
}

//随机一个目标队伍
func (this* armyWar) randArmyPosition(pos []*armyPosition) (*armyPosition, int){
	isEmpty := true
	for _, v := range pos {
		if v != nil {
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
		if pos[index] != nil && pos[index].soldiers != 0{
			return pos[index], index
		}
	}

	return nil, -1
}
