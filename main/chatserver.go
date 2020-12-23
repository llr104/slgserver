package main

import (
	"fmt"
	"os"
	"slgserver/config"
	"slgserver/net"
	"slgserver/server/chatserver"
	"slgserver/server/slgserver/conn"
)

func getChatServerAddr() string {
	host := config.File.MustValue("chatserver", "host", "")
	port := config.File.MustValue("chatserver", "port", "8002")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	chatserver.Init()
	needSecret := config.File.MustBool("chatserver", "need_secret", false)
	s := conn.NewServer(getChatServerAddr(), needSecret)
	s.Router(chatserver.MyRouter)
	s.ConnOnClose(func(sconn *net.ServerConn) {
		conn.ConnMgr.RemoveConn(sconn)
	})
	s.Start()
}
