package run

import (
	"slgserver/db"
	"slgserver/net"
	"slgserver/server/controller"
	"slgserver/server/entity"
)

var MyRouter = &net.Router{}

func Init() {
	db.TestDB()
	initRouter()

	entity.BCMgr.Load()
	entity.NMMgr.Load()
	entity.RCMgr.Load()
	entity.RBMgr.Load()

}

func initRouter() {
	controller.DefaultAccount.InitRouter(MyRouter)
	controller.DefaultRole.InitRouter(MyRouter)
	controller.DefaultMap.InitRouter(MyRouter)
}