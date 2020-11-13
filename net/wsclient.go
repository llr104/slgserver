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

type ReqBody struct {
	Seq     int64		`json:"seq"`
	Name 	string 		`json:"name"`
	Msg		interface{}	`json:"msg"`
}

type RspBody struct {
	Seq     int64		`json:"seq"`
	Name 	string 		`json:"name"`
	Code	int			`json:"code"`
	Msg		interface{}	`json:"msg"`
}

type WsMsgReq struct {
	Body*	ReqBody
	Conn*	WSConn
}

type WsMsgRsp struct {
	Body*	RspBody
}

const HandshakeMsg = "handshake"
const HeartbeatMsg = "heartbeat"

type Handshake struct {
	Key 	string `json:"key"`
}

type Heartbeat struct {
	CTime int64	`json:"ctime"`
	STime int64	`json:"stime"`
}


// 客户端连接
type WSConn struct {
	wsSocket   *websocket.Conn // 底层websocket
	outChan    chan *WsMsgRsp   // 写队列
	mutex      sync.Mutex     // 避免重复关闭管道
	isClosed   bool
	needSecret bool
	router     *Router
	onClose    func(conn* WSConn)
	//链接属性
	property map[string]interface{}
	//保护链接属性修改的锁
	propertyLock sync.RWMutex
}

func NewWSConn(wsSocket *websocket.Conn, needSecret bool) *WSConn {
	conn := &WSConn{
		wsSocket: wsSocket,
		outChan: make(chan *WsMsgRsp, 1000),
		isClosed:false,
		property:make(map[string]interface{}),
		needSecret:needSecret,
	}

	return conn
}


func (conn *WSConn) Running() {
	// 读协程
	go conn.wsReadLoop()
	// 写协程
	go conn.wsWriteLoop()
}

func (conn *WSConn) Addr() string  {
	return conn.wsSocket.RemoteAddr().String()
}

func (conn *WSConn) Send(name string, data interface{}) {
	rsp := &WsMsgRsp{Body: &RspBody{Name: name, Msg: data}}
	conn.outChan <- rsp

}

func (conn *WSConn) wsReadLoop() {
	defer func() {
		if err := recover(); err != nil {
			e := fmt.Sprintf("%v", err)
			log.DefaultLog.Error("wsReadLoop error", zap.String("err", e))
			conn.Close()
		}
	}()

	for {
		// 读一个message
		_, data, err := conn.wsSocket.ReadMessage()
		if err != nil {
			break
		}

		data, err = util.UnZip(data)
		if err != nil {
			log.DefaultLog.Error("wsReadLoop UnZip error", zap.Error(err))
			continue
		}

		body := &ReqBody{}
		if conn.needSecret {
			//检测是否有加密，没有加密发起Handshake
			if secretKey, err:= conn.GetProperty("secretKey");err == nil {
				key := secretKey.(string)
				d, err := util.AesCBCDecrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
				if err != nil {
					log.DefaultLog.Error("AesDecrypt error", zap.Error(err))
					conn.Handshake()
				}else{
					data = d
				}
			}else{
				log.DefaultLog.Info("secretKey not found client need handshake", zap.Error(err))
				conn.Handshake()
				return
			}
		}

		go func() {
			if err := util.Unmarshal(data, body); err == nil {
				req := &WsMsgReq{Conn: conn, Body: body}
				rsp := &WsMsgRsp{Body: &RspBody{Name: body.Name, Seq: req.Body.Seq}}

				if req.Body.Name == HeartbeatMsg {
					h := &Heartbeat{}
					mapstructure.Decode(body.Msg, h)
					h.STime = time.Now().UnixNano()/1e6
					rsp.Body.Msg = h
				}else{
					if conn.router != nil {
						conn.router.Run(req, rsp)
					}
				}
				conn.outChan <- rsp
			}else{
				log.DefaultLog.Error("wsReadLoop Unmarshal error", zap.Error(err))
				conn.Handshake()
			}
		}()
	}

	conn.Close()
}

func (conn *WSConn) wsWriteLoop() {

	defer func() {
		if err := recover(); err != nil {
			log.DefaultLog.Error("wsWriteLoop error")
			conn.Close()
		}
	}()

	for {
		select {
		// 取一个消息
		case msg := <- conn.outChan:
			// 写给websocket
			data, err := util.Marshal(msg.Body)
			if err == nil {
				if secretKey, err:= conn.GetProperty("secretKey"); err == nil {
					key := secretKey.(string)
					data, _ = util.AesCBCEncrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
				}
			}else {
				log.DefaultLog.Error("wsWriteLoop Marshal body error", zap.Error(err))
				return
			}

			if data, err := util.Zip(data); err == nil{
				if err := conn.wsSocket.WriteMessage(websocket.BinaryMessage, data); err != nil {
					conn.Close()
					return
				}
			}
		}
	}
}


func (conn *WSConn) Close() {
	conn.wsSocket.Close()
	conn.mutex.Lock()
	defer conn.mutex.Unlock()
	if !conn.isClosed {
		if conn.onClose != nil{
			conn.onClose(conn)
		}
		conn.isClosed = true
	}

}


//设置链接属性
func (c *WSConn) SetProperty(key string, value interface{}) {
	c.propertyLock.Lock()
	defer c.propertyLock.Unlock()

	c.property[key] = value
}

//获取链接属性
func (c *WSConn) GetProperty(key string) (interface{}, error) {
	c.propertyLock.RLock()
	defer c.propertyLock.RUnlock()

	if value, ok := c.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

func (conn *WSConn) SetRouter(router *Router)  {
	conn.router = router
}

func (conn *WSConn) SetOnClose(hookFunc func (*WSConn))  {
	conn.onClose = hookFunc
}


//移除链接属性
func (conn *WSConn) RemoveProperty(key string) {
	conn.propertyLock.Lock()
	defer conn.propertyLock.Unlock()

	delete(conn.property, key)
}



//握手协议
func (conn *WSConn) Handshake(){

	secretKey := ""
	if conn.needSecret {
		key, err:= conn.GetProperty("secretKey")
		if err == nil {
			secretKey = key.(string)
		}else{
			secretKey = util.RandSeq(16)
		}
	}

	handshake := &Handshake{Key: secretKey}
 	body := &RspBody{Name: HandshakeMsg, Msg: handshake}
 	if data, err := util.Marshal(body); err == nil {
		conn.SetProperty("secretKey", secretKey)

		log.DefaultLog.Info("handshake secretKey",
			zap.String("secretKey", secretKey))

		if data, err = util.Zip(data); err == nil{
			conn.wsSocket.WriteMessage(websocket.BinaryMessage, data)
		}

	}else {
		log.DefaultLog.Error("handshake Marshal body error", zap.Error(err))
	}
}