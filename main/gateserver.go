package main

import (
	"fmt"
	"os"
	"slgserver/config"
	"slgserver/net"
	"slgserver/server/gateserver"
	"slgserver/server/gateserver/controller"
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
	s := net.NewServer(getGateServerAddr(), needSecret)
	s.Router(gateserver.MyRouter)
	s.SetOnBeforeClose(controller.GHandle.OnServerConnClose)
	s.Start()
}
