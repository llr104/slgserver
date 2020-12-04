package logic

import (
	"fmt"
	"math/rand"
	"slgserver/server/model"
	"slgserver/server/static_conf/facility"
	"slgserver/server/static_conf/general"
	"slgserver/util"
)

//战斗位置的属性
type armyPosition struct {
	GId			int		//武将id
	Soldiers	int		//兵力
	Force    	int		//武力
	Strategy	int		//策略
	Defense  	int		//防御
	Speed    	int		//速度
	Destroy  	int		//破坏
	Arms		int		//兵种
	Position	int		//位置
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
	if w.attackPos[0].Soldiers == 0{
		result.result = 0
	}else if w.defensePos[0].Soldiers != 0{
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

	this.attackPos = make([]*armyPosition, 0)
	this.defensePos = make([]*armyPosition, 0)

	for i, gid := range this.attack.GeneralArray {
		if gid == 0 {
			this.attackPos = append(this.attackPos, nil)
		}else{
			if g, ok := GMgr.GetByGId(gid); ok {
				cfg := general.General.GMap[g.CfgId]
				pos := &armyPosition{
					GId: g.Id,
					Soldiers: this.attack.SoldierArray[i],
					Force: cfg.Force + attackAdds[0],
					Defense: cfg.Defense + attackAdds[1],
					Speed: cfg.Speed + attackAdds[2],
					Strategy: cfg.Strategy + attackAdds[3],
					Destroy: cfg.Destroy,
					Arms: g.CurArms,
					Position: i,
				}
				this.attackPos = append(this.attackPos, pos)
			}else{
				this.attackPos = append(this.attackPos, nil)
			}
		}
	}

	for i, gid := range this.defense.GeneralArray {
		if g, ok := GMgr.GetByGId(gid); ok {
			cfg := general.General.GMap[g.CfgId]
			pos := &armyPosition{
				GId: g.Id,
				Soldiers: this.attack.SoldierArray[i],
				Force: cfg.Force + defenseAdds[0],
				Defense: cfg.Defense + defenseAdds[1],
				Speed: cfg.Speed + defenseAdds[2],
				Strategy: cfg.Strategy + defenseAdds[3],
				Destroy: cfg.Destroy,
				Arms: g.CurArms,
				Position: i,
			}
			this.defensePos = append(this.defensePos, pos)
		}else{
			this.defensePos = append(this.defensePos, nil)
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
		if posAttack == nil || posAttack.Soldiers == 0{
			continue
		}
		//计算
		posDefense, index := this.randArmyPosition(defense)
		if posDefense == nil{
			continue
		}

		hurm := posAttack.Soldiers*posAttack.Force/1000
		def := posDefense.Soldiers*posDefense.Defense/1000

		kill := hurm-def
		if kill > 0{
			kill = util.MinInt(kill, posDefense.Soldiers)
			posDefense.Soldiers -= kill
			defenseArmy.SoldierArray[index] -= kill
		}else{
			kill = 0
		}

		b := battle{AId: posAttack.GId, ALoss: 0, DId: posDefense.GId, DLoss: kill}
		war.Battle = append(war.Battle, b.to())

		//大营干死了，直接结束
		if posDefense.Position == 0 && posDefense.Soldiers == 0 {
			isEnd = true
			goto end
		}
	}

	//防守方回合
	for _, posAttack := range defense {
		if posAttack == nil || posAttack.Soldiers == 0{
			continue
		}

		//计算
		posDefense, index := this.randArmyPosition(attack)
		hurm := posAttack.Soldiers*posAttack.Force/10000
		def := posDefense.Soldiers*posDefense.Defense/10000

		if posDefense == nil{
			continue
		}

		kill := hurm-def
		if kill > 0{
			kill = util.MinInt(kill, posDefense.Soldiers)
			posDefense.Soldiers -= kill
			attackArmy.SoldierArray[index] -= kill
		}else{
			kill = 0
		}

		b := battle{AId: posAttack.GId, ALoss: 0, DId: posDefense.GId, DLoss: kill}
		war.Battle = append(war.Battle, b.to())

		//大营干死了，直接结束
		if posDefense.Position == 0 && posDefense.Soldiers == 0 {
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
		if pos[index] != nil && pos[index].Soldiers != 0{
			return pos[index], index
		}
	}

	return nil, -1
}
