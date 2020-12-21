package conn

import (
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"net/http"
	"slgserver/log"
	"slgserver/net"
)

// http升级websocket协议的配置
var wsUpgrader = websocket.Upgrader{
	// 允许所有CORS跨域请求
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}


type server struct {
	addr	string
	router	*net.Router
	needSecret bool
}

func NewServer(addr string, needSecret bool) *server{
	s := server{
		addr: addr,
		needSecret: needSecret,
	}
	return &s
}

func (this* server) Router(router *net.Router) {
	this.router = router
}

func (this* server) Start()  {
	log.DefaultLog.Info("server starting")
	http.HandleFunc("/", this.wsHandler)
	http.ListenAndServe(this.addr, nil)
}


func (this* server) wsHandler(resp http.ResponseWriter, req *http.Request) {

	wsSocket, err := wsUpgrader.Upgrade(resp, req, nil)
	if err != nil {
		return
	}

	conn := ConnMgr.NewConn(wsSocket, this.needSecret)
	log.DefaultLog.Info("client connect", zap.String("addr", wsSocket.RemoteAddr().String()))

	conn.SetRouter(this.router)
	conn.SetOnClose(func(conn *net.WSConn) {
		ConnMgr.RemoveConn(conn)
		log.DefaultLog.Info("client disconnect", zap.String("addr", wsSocket.RemoteAddr().String()))
	})

	conn.Running()
	conn.Handshake()

}
