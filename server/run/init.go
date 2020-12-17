package run

import (
	"slgserver/db"
	"slgserver/net"
	"slgserver/server/controller"
	"slgserver/server/logic"
	"slgserver/server/logic/mgr"
	"slgserver/server/static_conf"
	"slgserver/server/static_conf/facility"
	"slgserver/server/static_conf/general"
)

var MyRouter = &net.Router{}

func Init() {
	db.TestDB()

	facility.FConf.Load()
	general.GenBasic.Load()
	general.General.Load()
	static_conf.Basic.Load()
	static_conf.MapBuildConf.Load()

	//需要先加载联盟相关的信息
	mgr.UnionMgr.Load()
	mgr.RAttrMgr.Load()
	mgr.NMMgr.Load()
	mgr.RCMgr.Load()
	mgr.RBMgr.Load()
	mgr.RFMgr.Load()
	mgr.RResMgr.Load()
	mgr.GMgr.Load()
	mgr.AMgr.Load()

	logic.Init()

	//全部初始化完才注册路由，防止服务器还没启动就绪收到请求
	initRouter()
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