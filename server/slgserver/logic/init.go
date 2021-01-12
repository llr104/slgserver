package logic

import (
	"slgserver/server/slgserver/logic/mgr"
	"slgserver/server/slgserver/model"
	"time"
)


var Union *coalitionLogic
var ArmyLogic *armyLogic

func BeforeInit()  {
	//初始化一些方法
	model.ArmyIsInView = armyIsInView
	model.GetUnionId = getUnionId
	model.GetRoleNickName = mgr.RoleNickName
	model.GetParentId = getParentId
	model.GetMainMembers = getMainMembers
	model.GetUnionName = getUnionName
	model.GetYield = mgr.GetYield
	model.GetDepotCapacity = mgr.GetDepotCapacity
	model.GetCityCost = mgr.GetCityCost
	model.GetMaxDurable = mgr.GetMaxDurable
	model.GetCityLv = mgr.GetCityLV
	model.MapResTypeLevel = mgr.NMMgr.MapResTypeLevel
}

//逻辑相关的初始化放在这里
func Init() {

	Union = NewCoalitionLogic()
	ArmyLogic = &armyLogic{
		arriveArmys:    make(chan *model.Army, 100),
		interruptId:    make(chan int, 100),
		updateArmys:    make(chan *model.Army, 100),
		outArmys:       make(map[int]*model.Army),
		endTimeArmys:   make(map[int64][]*model.Army),
		stopInPosArmys: make(map[int]map[int]*model.Army),
		passByPosArmys: make(map[int]map[int]*model.Army),
		sys:            NewSysArmy()}

	ArmyLogic.init()
	go ArmyLogic.check()
	go ArmyLogic.running()

}

func AfterInit() {
	go func() {
		for true {
			time.Sleep(1*time.Second)
			buildIds := mgr.RBMgr.CheckGiveUp()
			for _, buildId := range buildIds {
				ArmyLogic.GiveUp(buildId)
			}

			buildIds = mgr.RBMgr.CheckDestroy()
			for _, buildId := range buildIds {
				ArmyLogic.Interrupt(buildId)
			}
		}
	}()
}