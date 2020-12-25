package net

import (
	"errors"
	"fmt"
	"github.com/forgoer/openssl"
	"github.com/goinggo/mapstructure"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"slgserver/log"
	"slgserver/util"
	"sync"
	"time"
)



// 客户端连接
type ServerConn struct {
	wsSocket   	*websocket.Conn // 底层websocket
	outChan    	chan *WsMsgRsp  // 写队列
	isClosed   	bool
	needSecret 	bool
	Seq			int64
	router     	*Router
	beforeClose func(conn WSConn)
	onClose    	func(conn WSConn)
	//链接属性
	property 	map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
}

func NewServerConn(wsSocket *websocket.Conn, needSecret bool) *ServerConn {
	conn := &ServerConn{
		wsSocket: wsSocket,
		outChan: make(chan *WsMsgRsp, 1000),
		isClosed:false,
		property:make(map[string]interface{}),
		needSecret:needSecret,
		Seq: 0,
	}

	return conn
}

//开启异步
func (this *ServerConn) Start() {
	go this.wsReadLoop()
	go this.wsWriteLoop()
}

func (this *ServerConn) Addr() string  {
	return this.wsSocket.RemoteAddr().String()
}

func (this *ServerConn) Push(name string, data interface{}) {
	rsp := &WsMsgRsp{Body: &RspBody{Name: name, Msg: data, Seq: 0}}
	this.outChan <- rsp
}

func (this *ServerConn) Send(name string, data interface{}) {
	this.Seq += 1
	rsp := &WsMsgRsp{Body: &RspBody{Name: name, Msg: data, Seq: this.Seq}}
	this.outChan <- rsp
}

func (this *ServerConn) wsReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			e := fmt.Sprintf("%v", err)
			log.DefaultLog.Error("wsReadLoop error", zap.String("err", e))
			this.Close()
		}
	}()

	for {
		// 读一个message
		_, data, err := this.wsSocket.ReadMessage()
		if err != nil {
			break
		}

		data, err = util.UnZip(data)
		if err != nil {
			log.DefaultLog.Error("wsReadLoop UnZip error", zap.Error(err))
			continue
		}

		body := &ReqBody{}
		if this.needSecret {
			//检测是否有加密，没有加密发起Handshake
			if secretKey, err:= this.GetProperty("secretKey"); err == nil {
				key := secretKey.(string)
				d, err := util.AesCBCDecrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
				if err != nil {
					log.DefaultLog.Error("AesDecrypt error", zap.Error(err))
					this.Handshake()
				}else{
					data = d
				}
			}else{
				log.DefaultLog.Info("secretKey not found client need handshake", zap.Error(err))
				this.Handshake()
				return
			}
		}

		go func() {
			if err := util.Unmarshal(data, body); err == nil {
				req := &WsMsgReq{Conn: this, Body: body}
				rsp := &WsMsgRsp{Body: &RspBody{Name: body.Name, Seq: req.Body.Seq}}

				if req.Body.Name == HeartbeatMsg {
					h := &Heartbeat{}
					mapstructure.Decode(body.Msg, h)
					h.STime = time.Now().UnixNano()/1e6
					rsp.Body.Msg = h
				}else{
					if this.router != nil {
						this.router.Run(req, rsp)
					}
				}
				this.outChan <- rsp
			}else{
				log.DefaultLog.Error("wsReadLoop Unmarshal error", zap.Error(err))
				this.Handshake()
			}
		}()
	}

	this.Close()
}

func (this *ServerConn) wsWriteLoop() {

	defer func() {
		if err := recover(); err != nil {
			log.DefaultLog.Error("wsWriteLoop error")
			this.Close()
		}
	}()

	for {
		select {
		// 取一个消息
		case msg := <- this.outChan:
			// 写给websocket
			this.write(msg.Body)
		}
	}
}


func (this *ServerConn) write(msg interface{}) error{
	data, err := util.Marshal(msg)
	if err == nil {
		if this.needSecret {
			if secretKey, err:= this.GetProperty("secretKey"); err == nil {
				key := secretKey.(string)
				data, _ = util.AesCBCEncrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
			}
		}
	}else {
		log.DefaultLog.Error("wsWriteLoop Marshal body error", zap.Error(err))
		return err
	}

	if data, err := util.Zip(data); err == nil{
		if err := this.wsSocket.WriteMessage(websocket.BinaryMessage, data); err != nil {
			this.Close()
			return err
		}
	}else{
		return err
	}
	return nil
}

func (this *ServerConn) Close() {
	this.wsSocket.Close()
	if !this.isClosed {
		this.isClosed = true

		if this.beforeClose != nil{
			this.beforeClose(this)
		}

		if this.onClose != nil{
			this.onClose(this)
		}
	}
}

//设置链接属性
func (this *ServerConn) SetProperty(key string, value interface{}) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	this.property[key] = value
}

//获取链接属性
func (this *ServerConn) GetProperty(key string) (interface{}, error) {
	this.propertyLock.RLock()
	defer this.propertyLock.RUnlock()

	if value, ok := this.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

func (this *ServerConn) SetRouter(router *Router)  {
	this.router = router
}

func (this *ServerConn) SetOnClose(hookFunc func (WSConn))  {
	this.onClose = hookFunc
}

func (this *ServerConn) SetOnBeforeClose(hookFunc func (WSConn))  {
	this.beforeClose = hookFunc
}

//移除链接属性
func (this *ServerConn) RemoveProperty(key string) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	delete(this.property, key)
}

//握手协议
func (this *ServerConn) Handshake(){

	secretKey := ""
	if this.needSecret {
		key, err:= this.GetProperty("secretKey")
		if err == nil {
			secretKey = key.(string)
		}else{
			secretKey = util.RandSeq(16)
		}
	}

	handshake := &Handshake{Key: secretKey}
 	body := &RspBody{Name: HandshakeMsg, Msg: handshake}
 	if data, err := util.Marshal(body); err == nil {
 		if secretKey != ""{
			this.SetProperty("secretKey", secretKey)
		}else{
			this.RemoveProperty("secretKey")
		}

		log.DefaultLog.Info("handshake secretKey",
			zap.String("secretKey", secretKey))

		if data, err = util.Zip(data); err == nil{
			this.wsSocket.WriteMessage(websocket.BinaryMessage, data)
		}

	}else {
		log.DefaultLog.Error("handshake Marshal body error", zap.Error(err))
	}
}