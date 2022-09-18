package logic

import (
	"time"

	"github.com/llr104/slgserver/server/slgserver/logic/army"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/logic/union"
	"github.com/llr104/slgserver/server/slgserver/model"
)

var Union *union.UnionLogic
var ArmyLogic *army.ArmyLogic

func BeforeInit() {
	//初始化一些方法
	model.ArmyIsInView = army.ArmyIsInView
	model.GetUnionId = union.GetUnionId
	model.GetRoleNickName = mgr.RoleNickName
	model.GetParentId = union.GetParentId
	model.GetMainMembers = union.GetMainMembers
	model.GetUnionName = union.GetUnionName
	model.GetYield = mgr.GetYield
	model.GetDepotCapacity = mgr.GetDepotCapacity
	model.GetCityCost = mgr.GetCityCost
	model.GetMaxDurable = mgr.GetMaxDurable
	model.GetCityLv = mgr.GetCityLV
	model.MapResTypeLevel = mgr.NMMgr.MapResTypeLevel
}

//逻辑相关的初始化放在这里
func Init() {
	Union = union.Instance()
	ArmyLogic = army.Instance()
}

func AfterInit() {
	go func() {
		for true {
			time.Sleep(1 * time.Second)
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
