package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"slgserver/config"
	"slgserver/server/conn"
	"slgserver/server/run"
)

// http升级websocket协议的配置
var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


func getServerAddr() string {
	host := config.File.MustValue("server", "host", "")
	port := config.File.MustValue("server", "port", "8001")
	return host + ":" + port
}


func main() {
	fmt.Println(os.Getwd())
	run.Init()
	needSecret := config.File.MustBool("server", "need_secret", false)
	s := conn.NewServer(getServerAddr(), needSecret)
	s.Router(run.MyRouter)
	s.Start()
}


