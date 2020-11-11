package main

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"slgserver/config"
	"slgserver/log"
	"slgserver/net"
	"slgserver/server"
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

	run.Init()
	log.DefaultLog.Info("slg server starting")
	http.HandleFunc("/", wsHandler)
	http.ListenAndServe(getServerAddr(), nil)
}



func wsHandler(resp http.ResponseWriter, req *http.Request) {
	needSecret := config.File.MustBool("server", "need_secret", false)

	wsSocket, err := wsUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}

	conn := server.DefaultConnMgr.NewConn(wsSocket, needSecret)
	log.DefaultLog.Info("client connect", zap.String("addr", wsSocket.RemoteAddr().String()))

	conn.SetRouter(run.MyRouter)
	conn.SetOnClose(func(conn *net.WSConn) {
		server.DefaultConnMgr.RemoveConn(conn)
		log.DefaultLog.Info("client disconnect", zap.String("addr", wsSocket.RemoteAddr().String()))
	})

	conn.Running()
	conn.Handshake()

}
