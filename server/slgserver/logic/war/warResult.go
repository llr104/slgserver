package war

import (
	"math/rand"
	"slgserver/server/slgserver/static_conf"
	"slgserver/server/slgserver/static_conf/general"
	"slgserver/util"
)

const maxRound = 10

type warResult struct {
	camp   *warCamp
	rounds []*warRound
	result int //0失败，1平，2胜利
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

//回合
func (this *warResult) runRound() (*warRound, bool) {

	war := &warRound{}
	n := rand.Intn(10)
	attack := this.camp.attackPos
	defense := this.camp.defensePos

	isEnd := false
	//随机先手
	if n%2 == 0 {
		attack = this.camp.defensePos
		defense = this.camp.attackPos
	}

	for _, att := range attack {

		////////攻击方begin//////////
		if att == nil || att.soldiers == 0 {
			continue
		}

		def, _ := this.camp.randArmyPosition(defense)
		if def == nil {
			isEnd = true
			goto end
		}

		attHarmRatio := general.GenArms.GetHarmRatio(att.arms, def.arms)
		attHarm := float64(util.AbsInt(att.force-def.defense)*att.soldiers) * attHarmRatio * 0.0005
		attKill := int(attHarm)
		attKill = util.MinInt(attKill, def.soldiers)
		def.soldiers -= attKill
		att.general.Exp += attKill * 5

		//大营干死了，直接结束
		if def.position == 0 && def.soldiers == 0 {
			isEnd = true
			goto end
		}
		////////攻击方end//////////

		////////防守方begin//////////
		if def.soldiers == 0 || att.soldiers == 0 {
			continue
		}

		defHarmRatio := general.GenArms.GetHarmRatio(def.arms, att.arms)
		defHarm := float64(util.AbsInt(def.force-att.defense)*def.soldiers) * defHarmRatio * 0.0005
		defKill := int(defHarm)

		defKill = util.MinInt(defKill, att.soldiers)
		att.soldiers -= defKill
		def.general.Exp += defKill * 5

		b := hit{AId: att.general.Id, ALoss: defKill, DId: def.general.Id, DLoss: attKill}
		war.Battle = append(war.Battle, b.to())

		//大营干死了，直接结束
		if att.position == 0 && att.soldiers == 0 {
			isEnd = true
			goto end
		}
		////////防守方end//////////

	}

end:
	return war, isEnd
}

func (this *warResult) getRounds() []*warRound {
	return this.rounds
}
