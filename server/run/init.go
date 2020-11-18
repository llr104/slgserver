package run

import (
	"slgserver/db"
	"slgserver/net"
	"slgserver/server/controller"
	"slgserver/server/entity"
	"slgserver/server/static_conf/facility"
	"slgserver/server/static_conf/general"
)

var MyRouter = &net.Router{}

func Init() {
	db.TestDB()
	initRouter()

	facility.FConf.Load()
	general.GenBasic.Load()

	entity.BCMgr.Load()
	entity.NMMgr.Load()
	entity.RCMgr.Load()
	entity.RBMgr.Load()
	entity.RFMgr.Load()
	entity.RResMgr.Load()
	entity.AMgr.Load()

}

func initRouter() {
	controller.DefaultAccount.InitRouter(MyRouter)
	controller.DefaultRole.InitRouter(MyRouter)
	controller.DefaultMap.InitRouter(MyRouter)
	controller.DefaultCity.InitRouter(MyRouter)
	controller.DefaultGeneral.InitRouter(MyRouter)
}