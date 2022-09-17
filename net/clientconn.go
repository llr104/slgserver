package net

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/forgoer/openssl"
	"github.com/goinggo/mapstructure"
	"github.com/gorilla/websocket"
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/util"
	"go.uber.org/zap"
)

// 客户端连接
type ClientConn struct {
	wsSocket   		*websocket.Conn // 底层websocket
	isClosed   		bool
	Seq				int64
	onClose    		func(conn*ClientConn)
	onPush    		func(conn*ClientConn, body*RspBody)
	//链接属性
	property 		map[string]interface{}
	//保护链接属性修改的锁
	propertyLock  	sync.RWMutex
	syncCtxs      	map[int64]*syncCtx
	syncLock      	sync.RWMutex
	handshakeChan 	chan bool
	handshake		bool
}

func NewClientConn(wsSocket *websocket.Conn) *ClientConn {
	conn := &ClientConn{
		wsSocket:      wsSocket,
		isClosed:      false,
		property:      make(map[string]interface{}),
		Seq:           0,
		syncCtxs:      make(map[int64]*syncCtx),
		handshakeChan: make(chan bool),
	}

	return conn
}

func (this *ClientConn) waitHandshake() bool{
	if this.handshake == false{
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		select {
			case _ = <-this.handshakeChan:{
				log.DefaultLog.Info("recv handshakeChan")
				return true
			}
			case <-ctx.Done():{
				log.DefaultLog.Info("recv handshakeChan timeout")
				return false
			}
		}
	}
	return true
}

func (this *ClientConn)	Start() bool{
	this.handshake = false
	go this.wsReadLoop()
	return this.waitHandshake()
}

func (this *ClientConn) Addr() string  {
	return this.wsSocket.RemoteAddr().String()
}

func (this *ClientConn) Push(name string, data interface{}) {
	rsp := &WsMsgRsp{Body: &RspBody{Name: name, Msg: data, Seq: 0}}
	this.write(rsp.Body)
}

func (this *ClientConn) Send(name string, data interface{}) *RspBody{
	this.syncLock.Lock()
	sync := newSyncCtx()
	this.Seq += 1
	seq := this.Seq
	req := ReqBody{Name: name, Msg: data, Seq: seq}
	this.syncCtxs[this.Seq] = sync
	this.syncLock.Unlock()

	rsp := &RspBody{Code: constant.OK, Name: name, Seq: seq }
	err := this.write(req)
	if err != nil{
		sync.cancel()
	}else{
		r := sync.wait()
		if r == nil{
			rsp.Code = constant.ProxyConnectError
		}else{
			rsp = r
		}
	}

	this.syncLock.Lock()
	delete(this.syncCtxs, seq)
	this.syncLock.Unlock()

	return rsp
}

func (this *ClientConn) wsReadLoop() {
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

		//需要检测是否有加密
		body := &RspBody{}
		if secretKey, err := this.GetProperty("secretKey"); err == nil {
			key := secretKey.(string)
			d, err := util.AesCBCDecrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
			if err != nil {
				log.DefaultLog.Error("AesDecrypt error", zap.Error(err))
			}else{
				data = d
			}
		}

		if err := util.Unmarshal(data, body); err == nil {
			if body.Seq == 0 {
				if body.Name == HandshakeMsg{
					h := Handshake{}
					mapstructure.Decode(body.Msg, &h)
					log.DefaultLog.Info("client 收到握手协议", zap.String("data", string(data)))
					if h.Key != ""{
						this.SetProperty("secretKey", h.Key)
					}else{
						this.RemoveProperty("secretKey")
					}
					this.handshake = true
					this.handshakeChan <- true
				}else{
					//推送，需要推送到指定的代理连接
					if this.onPush != nil{
						this.onPush(this, body)
					}else{
						log.DefaultLog.Warn("clientconn not deal push")
					}
				}
			}else{
				this.syncLock.RLock()
				s, ok := this.syncCtxs[body.Seq]
				this.syncLock.RUnlock()
				if ok {
					s.outChan <- body
				}else{
					log.DefaultLog.Warn("seq not found sync",
						zap.Int64("seq", body.Seq),
						zap.String("msgName", body.Name))
				}
			}

		}else{
			log.DefaultLog.Error("wsReadLoop Unmarshal error", zap.Error(err))
		}
	}

	this.Close()
}


func (this *ClientConn) write(msg interface{}) error{
	data, err := util.Marshal(msg)
	if err == nil {
		if secretKey, err:= this.GetProperty("secretKey"); err == nil {
			key := secretKey.(string)
			log.DefaultLog.Info("secretKey", zap.String("secretKey", key))
			data, _ = util.AesCBCEncrypt(data, []byte(key), []byte(key), openssl.ZEROS_PADDING)
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

func (this *ClientConn) Close() {
	this.wsSocket.Close()
	if !this.isClosed {
		this.isClosed = true
		if this.onClose != nil{
			this.onClose(this)
		}
	}
}

//设置链接属性
func (this *ClientConn) SetProperty(key string, value interface{}) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	this.property[key] = value
}

//获取链接属性
func (this *ClientConn) GetProperty(key string) (interface{}, error) {
	this.propertyLock.RLock()
	defer this.propertyLock.RUnlock()

	if value, ok := this.property[key]; ok {
		return value, nil
	} else {
		return nil, errors.New("no property found")
	}
}

//移除链接属性
func (this *ClientConn) RemoveProperty(key string) {
	this.propertyLock.Lock()
	defer this.propertyLock.Unlock()

	delete(this.property, key)
}

func (this *ClientConn) SetOnClose(hookFunc func (*ClientConn))  {
	this.onClose = hookFunc
}

func (this *ClientConn) SetOnPush(hookFunc func (*ClientConn, *RspBody))  {
	this.onPush = hookFunc
}
