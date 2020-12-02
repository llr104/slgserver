package logic

import (
	"math/rand"
	"slgserver/server/model"
	"slgserver/server/static_conf/general"
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

type warRound struct {
	Battle[]	battle	`json:"battle"`
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
					Force: cfg.Force,
					Strategy: cfg.Strategy,
					Defense: cfg.Defense,
					Speed: cfg.Speed,
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
				Force: cfg.Force,
				Strategy: cfg.Strategy,
				Defense: cfg.Defense,
				Speed: cfg.Speed,
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

func (this* armyWar) battle() []*warRound{
	rounds := make([]*warRound, 0)
	cur := 0
	for true{
		r := this.round()
		rounds = append(rounds, r)
		cur += 1
		if cur >= maxRound{
			break
		}
	}
	return rounds
}

//回合
func (this* armyWar) round() *warRound {

	war := &warRound{}
	war.Battle = make([]battle, 0)
	n := rand.Intn(10)
	attack := this.attackPos
	defense := this.defensePos

	//随机先手
	if n % 2 == 0{
		attack = this.defensePos
		defense = this.attackPos
	}

	//攻击方回合
	for _, posAttack := range attack {
		if posAttack == nil{
			continue
		}
		//计算
		posDefense := this.randArmyPosition(defense)
		hurm := posAttack.Soldiers*posAttack.Force
		def := posDefense.Soldiers*posDefense.Defense

		kill := hurm-def
		if kill > 0{
			posDefense.Soldiers -= kill
		}

		b := battle{AId: posAttack.GId, ALoss: 0, DId: posDefense.GId, DLoss: kill}
		war.Battle = append(war.Battle, b)

		//大营干死了，直接结束
		if posDefense.Position == 1 && posDefense.Soldiers == 0 {
			goto end
		}
	}

	//防守方回合
	for _, posAttack := range defense {
		if posAttack == nil{
			continue
		}

		//计算
		posDefense := this.randArmyPosition(attack)
		hurm := posAttack.Soldiers*posAttack.Force
		def := posDefense.Soldiers*posDefense.Defense

		kill := hurm-def
		if kill > 0{
			posDefense.Soldiers -= kill
		}

		b := battle{AId: posAttack.GId, ALoss: 0, DId: posDefense.GId, DLoss: kill}
		war.Battle = append(war.Battle, b)

		//大营干死了，直接结束
		if posDefense.Position == 1 && posDefense.Soldiers == 0 {
			goto end
		}
	}

	end:
	return war
}

//随机一个目标队伍
func (this* armyWar) randArmyPosition(pos []*armyPosition) *armyPosition{
	isEmpty := true
	for _, v := range pos {
		if v != nil {
			isEmpty = false
			break
		}
	}

	if isEmpty {
		return nil
	}

	for true {
		r := rand.Intn(100)
		index := r % len(pos)
		if pos[index] != nil{
			return pos[index]
		}
	}

	return nil
}