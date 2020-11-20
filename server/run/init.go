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

	logic.BCMgr.Load()
	logic.NMMgr.Load()
	logic.RCMgr.Load()
	logic.RBMgr.Load()
	logic.RFMgr.Load()
	logic.RResMgr.Load()
	logic.AMgr.Load()

}

func initRouter() {
	controller.DefaultAccount.InitRouter(MyRouter)
	controller.DefaultRole.InitRouter(MyRouter)
	controller.DefaultMap.InitRouter(MyRouter)
	controller.DefaultCity.InitRouter(MyRouter)
	controller.DefaultGeneral.InitRouter(MyRouter)
}