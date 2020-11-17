package run

import (
	"slgserver/db"
	"slgserver/net"
	"slgserver/server/controller"
	"slgserver/server/entity"
	"slgserver/server/static_conf"
)

var MyRouter = &net.Router{}

func Init() {
	db.TestDB()
	initRouter()

	static_conf.FConf.Load()
	static_conf.FGEN.Load()
	static_conf.FPRC.Load()
	static_conf.FMBS.Load()
	static_conf.FARMY.Load()
	static_conf.FCAMP.Load()
	static_conf.FBarrack.Load()
	static_conf.FFCT.Load()
	static_conf.FWALL.Load()
	static_conf.FMarket.Load()
	static_conf.FSJT.Load()


	entity.BCMgr.Load()
	entity.NMMgr.Load()
	entity.RCMgr.Load()
	entity.RBMgr.Load()
	entity.RFMgr.Load()
	entity.RResMgr.Load()

}

func initRouter() {
	controller.DefaultAccount.InitRouter(MyRouter)
	controller.DefaultRole.InitRouter(MyRouter)
	controller.DefaultMap.InitRouter(MyRouter)
	controller.DefaultCity.InitRouter(MyRouter)
}