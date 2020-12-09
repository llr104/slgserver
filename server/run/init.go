package run

import (
	"slgserver/db"
	"slgserver/net"
	"slgserver/server/controller"
	"slgserver/server/logic"
	"slgserver/server/static_conf"
	"slgserver/server/static_conf/facility"
	"slgserver/server/static_conf/general"
)

var MyRouter = &net.Router{}

func Init() {
	db.TestDB()
	initRouter()

	facility.FConf.Load()
	general.GenBasic.Load()
	general.General.Load()
	static_conf.Basic.Load()
	static_conf.MapBuildConf.Load()

	//需要先加载联盟相关的信息
	logic.UnionMgr.Load()
	logic.RAttributeMgr.Load()

	logic.NMMgr.Load()
	logic.RCMgr.Load()
	logic.RBMgr.Load()
	logic.RFMgr.Load()
	logic.RResMgr.Load()
	logic.GMgr.Load()
	logic.AMgr.Load()

}

func initRouter() {
	controller.DefaultAccount.InitRouter(MyRouter)
	controller.DefaultRole.InitRouter(MyRouter)
	controller.DefaultMap.InitRouter(MyRouter)
	controller.DefaultCity.InitRouter(MyRouter)
	controller.DefaultGeneral.InitRouter(MyRouter)
	controller.DefaultWar.InitRouter(MyRouter)
	controller.DefaultCoalition.InitRouter(MyRouter)
	controller.DefaultInterior.InitRouter(MyRouter)
}