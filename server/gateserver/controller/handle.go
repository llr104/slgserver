package controller

import (
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"slgserver/config"
	"slgserver/constant"
	"slgserver/log"
	"slgserver/middleware"
	"slgserver/net"
	chat_proto "slgserver/server/chatserver/proto"
	"slgserver/server/slgserver/proto"
	"strings"
	"sync"
)

var GHandle = Handle{
	proxys: make(map[string]map[int64]*net.ProxyClient),
}

type Handle struct {
	proxyMutex sync.Mutex
	proxys     map[string]map[int64]*net.ProxyClient
	slgProxy   string
	chatProxy  string
	loginProxy string
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



func (this*Handle) InitRouter(r *net.Router) {
	this.init()
	g := r.Group("*").Use(middleware.ElapsedTime(), middleware.Log())
	g.AddRouter("*", this.all)
}

func (this*Handle) init() {
	this.slgProxy = config.File.MustValue("gateserver", "slg_proxy", "ws://127.0.0.1:8001")
	this.chatProxy = config.File.MustValue("gateserver", "chat_proxy", "ws://127.0.0.1:8002")
	this.loginProxy = config.File.MustValue("gateserver", "login_proxy", "ws://127.0.0.1:8003")
}

func (this*Handle) onPush(conn *net.ClientConn, body *net.RspBody) {
	gc, err := conn.GetProperty("gateConn")
	if err != nil{
		return
	}
	gateConn := gc.(net.WSConn)
	gateConn.Push(body.Name, body.Msg)
}

func (this*Handle) onProxyClose(conn *net.ClientConn) {
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

func (this*Handle) OnServerConnClose (conn net.WSConn){
	c, err := conn.GetProperty("cid")
	arr := make([]*net.ProxyClient, 0)

	if err == nil{
		cid := c.(int64)
		this.proxyMutex.Lock()
		for _, m := range this.proxys {
			proxy, ok := m[cid]
			if ok {
				arr = append(arr, proxy)
			}
			delete(m, cid)
		}
		this.proxyMutex.Unlock()
	}

	for _, client := range arr {
		client.Close()
	}

}

func (this*Handle) all(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	log.DefaultLog.Info("gateserver handle all begin",
		zap.String("proxyStr", req.Body.Proxy),
		zap.String("msgName", req.Body.Name))
	this.deal(req, rsp)

	if req.Body.Name == "role.enterServer" && rsp.Body.Code == constant.OK  {
		//登录聊天服
		rspObj := &proto.EnterServerRsp{}
		mapstructure.Decode(rsp.Body.Msg, rspObj)
		r := &chat_proto.LoginReq{RId: rspObj.Role.RId, NickName: rspObj.Role.NickName, Token: rspObj.Token}
		reqBody := &net.ReqBody{Seq: 0, Name: "chat.login", Msg: r, Proxy: ""}
		rspBody := &net.RspBody{Seq: 0, Name: "chat.login", Msg: r, Code: 0}
		this.deal(&net.WsMsgReq{Body: reqBody, Conn:req.Conn}, &net.WsMsgRsp{Body: rspBody})
	}

	log.DefaultLog.Info("gateserver handle all end",
		zap.String("proxyStr", req.Body.Proxy),
		zap.String("msgName", req.Body.Name))
}

func (this*Handle) deal(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	//协议转发
	proxyStr := req.Body.Proxy
	if isAccount(req.Body.Name){
		proxyStr = this.loginProxy
	}else if isChat(req.Body.Name){
		proxyStr = this.chatProxy
	} else{
		proxyStr = this.slgProxy
	}

	if proxyStr == ""{
		rsp.Body.Code = constant.ProxyNotInConnect
		return
	}

	this.proxyMutex.Lock()
	_, ok := this.proxys[proxyStr]
	if ok == false {
		this.proxys[proxyStr] = make(map[int64]*net.ProxyClient)
	}

	var err error
	var proxy *net.ProxyClient
	d, _ := req.Conn.GetProperty("cid")
	cid := d.(int64)
	proxy, ok = this.proxys[proxyStr][cid]
	this.proxyMutex.Unlock()

	if ok == false {
		proxy = net.NewProxyClient(proxyStr)

		this.proxyMutex.Lock()
		this.proxys[proxyStr][cid] = proxy
		this.proxyMutex.Unlock()

		//发起链接,这里是阻塞的，所以不要上锁
		err = proxy.Connect()
		if err == nil{
			proxy.SetProperty("cid", cid)
			proxy.SetProperty("proxy", proxyStr)
			proxy.SetProperty("gateConn", req.Conn)
			proxy.SetOnPush(this.onPush)
			proxy.SetOnClose(this.onProxyClose)
		}
	}

	if err != nil {
		this.proxyMutex.Lock()
		delete(this.proxys[proxyStr], cid)
		this.proxyMutex.Unlock()
		rsp.Body.Code = constant.ProxyConnectError
		return
	}

	rsp.Body.Seq = req.Body.Seq
	rsp.Body.Name = req.Body.Name

	r, err := proxy.Send(req.Body.Name, req.Body.Msg)
	if err == nil{
		rsp.Body.Code = r.Code
		rsp.Body.Msg = r.Msg
	}else{
		rsp.Body.Code = constant.ProxyConnectError
		rsp.Body.Msg = nil
	}
}

