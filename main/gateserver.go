package main

import (
	"fmt"
	"os"
	"slgserver/config"
	"slgserver/server/gateserver"
	"slgserver/server/gateserver/controller"
	"slgserver/server/slgserver/conn"
)

func getGateServerAddr() string {
	host := config.File.MustValue("gateserver", "host", "")
	port := config.File.MustValue("gateserver", "port", "8004")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	gateserver.Init()
	needSecret := config.File.MustBool("gateserver", "need_secret", false)
	s := conn.NewServer(getGateServerAddr(), needSecret)
	s.Router(gateserver.MyRouter)
	s.ConnOnClose(controller.DefaultHandle.OnServerConnClose)
	s.Start()
}
