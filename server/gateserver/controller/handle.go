package controller

import (
	"errors"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/log"
	"slgserver/middleware"
	"slgserver/net"
	"strings"
	"sync"
	"time"
)

var DefaultHandle = Handle{
	proxys: make(map[string]map[int64]*proxyClient),
}

type Handle struct {
	proxyMutex sync.Mutex
	proxys     map[string]map[int64]*proxyClient
}

func isAccount(msgName string) bool {
	sArr := strings.Split(msgName, ".")
	prefix := ""
	if len(sArr) == 2{
		prefix = sArr[0]
	}
	if prefix == "account"{
		return true
	}else{
		return false
	}
}

func isChat(msgName string) bool {
	sArr := strings.Split(msgName, ".")
	prefix := ""
	if len(sArr) == 2{
		prefix = sArr[0]
	}
	if prefix == "chat"{
		return true
	}else{
		return false
	}
}

type proxyClient struct {
	proxy	string
	conn	*net.ClientConn
}

func (this*proxyClient) connect() error {
	var dialer = websocket.Dialer{
		Subprotocols:     []string{"p1", "p2"},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 30 * time.Second,
	}
	ws, _, err := dialer.Dial(this.proxy, nil)
	if err == nil{
		this.conn = net.NewClientConn(ws)
		this.conn.Start()
	}
	return err
}

func (this*proxyClient) send(msgName string, msg interface{}) (*net.RspBody, error){
	if this.conn != nil {
		return this.conn.Send(msgName, msg), nil
	}
	return nil, errors.New("conn not found")
}

func newProxyClient(proxy string) *proxyClient {
	return & proxyClient{
		proxy: proxy,
	}
}


func (this*Handle) InitRouter(r *net.Router) {
	g := r.Group("*").Use(middleware.ElapsedTime(), middleware.Log())
	g.AddRouter("*", this.all)
}

func (this*Handle) onPush(conn *net.ClientConn, body *net.RspBody) {
	gc, err := conn.GetProperty("gateConn")
	if err != nil{
		return
	}
	gateConn := gc.(net.WSConn)
	gateConn.Push(body.Name, body.Msg)
}

func (this*Handle) onClose (conn *net.ClientConn) {
	p, err := conn.GetProperty("proxy")
	if err == nil {
		proxyStr := p.(string)
		this.proxyMutex.Lock()
		_, ok := this.proxys[proxyStr]
		if ok {
			c, err := conn.GetProperty("cid")
			if err == nil{
				cid := c.(int64)
				delete(this.proxys[proxyStr], cid)
			}
		}
		this.proxyMutex.Unlock()
	}
}

func (this*Handle) OnServerConnClose (conn *net.ServerConn){
	c, err := conn.GetProperty("cid")
	if err == nil{
		cid := c.(int64)
		this.proxyMutex.Lock()
		for _, m := range this.proxys {
			proxy, ok := m[cid]
			if ok {
				proxy.conn.Close()
			}
			delete(m, cid)
		}
		this.proxyMutex.Unlock()
	}
}

func (this*Handle) all(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	log.DefaultLog.Info("gateserver handle all",
		zap.String("proxyStr", req.Body.Proxy),
		zap.String("msgName", req.Body.Name))

	proxyStr := req.Body.Proxy
	if isAccount(req.Body.Name){
		//转发到登录服务
		proxyStr = "ws://127.0.0.1:8003"
	}else if isChat(req.Body.Name){
		proxyStr = "ws://127.0.0.1:8002"
	} else{
		proxyStr = "ws://127.0.0.1:8001"
	}

	if proxyStr == ""{
		rsp.Body.Code = constant.ProxyNotInConnect
		return
	}

	this.proxyMutex.Lock()
	_, ok := this.proxys[proxyStr]
	if ok == false {
		this.proxys[proxyStr] = make(map[int64]*proxyClient)
	}

	var err error
	var proxy *proxyClient
	d, _ := req.Conn.GetProperty("cid")
	cid := d.(int64)
	proxy, ok = this.proxys[proxyStr][cid]
	if ok == false {
		proxy = newProxyClient(proxyStr)
		this.proxys[proxyStr][cid] = proxy
		//发起链接
		err = proxy.connect()
		if err == nil{
			proxy.conn.SetProperty("cid", cid)
			proxy.conn.SetProperty("proxy", proxyStr)
			proxy.conn.SetProperty("gateConn", req.Conn)
			proxy.conn.SetOnPush(this.onPush)
			proxy.conn.SetOnClose(this.onClose)
		}
	}
	this.proxyMutex.Unlock()

	if err != nil{
		rsp.Body.Code = constant.ProxyConnectError
		return
	}

	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name

	r, err := proxy.send(req.Body.Name, req.Body.Msg)
	if err == nil{
		rsp.Body.Code = r.Code
		rsp.Body.Msg = r.Msg
	}else{
		rsp.Body.Code = constant.ProxyConnectError
		rsp.Body.Msg = nil
	}
}


