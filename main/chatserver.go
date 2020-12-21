package main

import (
	"fmt"
	"os"
	"slgserver/chatserver"
	"slgserver/config"
	"slgserver/server/conn"
)

func getChatServerAddr() string {
	host := config.File.MustValue("chatserver", "host", "")
	port := config.File.MustValue("chatserver", "port", "8001")
	return host + ":" + port
}

func main() {
	fmt.Println(os.Getwd())
	chatserver.Init()
	needSecret := config.File.MustBool("chatserver", "need_secret", false)
	s := conn.NewServer(getChatServerAddr(), needSecret)
	s.Router(chatserver.MyRouter)
	s.Start()
}
