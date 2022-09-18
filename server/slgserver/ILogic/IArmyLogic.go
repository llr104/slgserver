package ILogic

import "github.com/llr104/slgserver/server/slgserver/model"

type IArmyLogic interface {
	ArmyBack(army *model.Army)
	GetStopArmys(posId int) []*model.Army
	DeleteStopArmy(posId int)
	GetSysArmy(x, y int) []*model.Army
	DelSysArmy(x, y int)
}
