package main

import (
	"fmt"
	"os"
	"slgserver/config"
	"slgserver/net"
	"slgserver/server/slgserver/conn"
	"slgserver/server/slgserver/run"
)


func getServerAddr() string {
	host := config.File.MustValue("slgserver", "host", "")
	port := config.File.MustValue("slgserver", "port", "8001")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	run.Init()
	needSecret := config.File.MustBool("slgserver", "need_secret", false)
	s := conn.NewServer(getServerAddr(), needSecret)
	s.Router(run.MyRouter)
	s.ConnOnClose(func(sconn *net.ServerConn) {
		conn.ConnMgr.RemoveConn(sconn)
	})
	s.Start()
}


