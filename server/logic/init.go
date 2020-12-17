package logic

import (
	"slgserver/server/logic/mgr"
	"slgserver/server/model"
)


var Union *coalitionLogic
var ArmyLogic *armyLogic

//逻辑相关的初始化放在这里
func Init() {

	//初始化一些方法
	model.ArmyIsInView = armyIsInView
	model.GetUnionId = getUnionId
	model.GetRoleNickName = mgr.RoleNickName
	model.GetParentId = getParentId
	model.GetMainMembers = getMainMembers
	model.GetUnionName = getUnionName

	Union = NewCoalitionLogic()
	ArmyLogic = &armyLogic{
		arriveArmys:	make(chan *model.Army, 100),
		giveUpId:       make(chan int, 100),
		updateArmys:    make(chan *model.Army, 100),
		outArmys:		make(map[int]*model.Army),
		armyByEndTime:	make(map[int64][]*model.Army),
		stopInPosArmys: make(map[int]map[int]*model.Army),
		passbyPosArmys: make(map[int]map[int]*model.Army),
		sys:            NewSysArmy()}

	ArmyLogic.init()
	go ArmyLogic.check()
	go ArmyLogic.running()

}
