package ILogic

import "slgserver/server/slgserver/model"

type ISysArmyLogic interface {
	GetArmy(x, y int) []*model.Army
	DelArmy(x, y int)
}

type IArmyLogic interface {
	ArmyBack(army *model.Army)
	Sys() ISysArmyLogic
	GetStopArmys(posId int) []*model.Army
	DeleteStopArmy(posId int)
}
