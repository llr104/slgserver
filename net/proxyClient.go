package net

import (
	"errors"
	"github.com/gorilla/websocket"
	"time"
)

type ProxyClient struct {
	proxy	string
	conn	*ClientConn
}

func (this*ProxyClient) Connect() error {
	var dialer = websocket.Dialer{
		Subprotocols:     []string{"p1", "p2"},
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 30 * time.Second,
	}
	ws, _, err := dialer.Dial(this.proxy, nil)
	if err == nil{
		this.conn = NewClientConn(ws)
		this.conn.Start()
	}
	return err
}

func (this*ProxyClient) Send(msgName string, msg interface{}) (*RspBody, error){
	if this.conn != nil {
		return this.conn.Send(msgName, msg), nil
	}
	return nil, errors.New("conn not found")
}


func (this *ProxyClient) SetOnClose(hookFunc func (*ClientConn))  {
	if this.conn != nil {
		this.conn.SetOnClose(hookFunc)
	}
}

func (this *ProxyClient) SetOnPush(hookFunc func (*ClientConn, *RspBody))  {
	if this.conn != nil {
		this.conn.SetOnPush(hookFunc)
	}
}

func (this *ProxyClient) SetProperty(key string, value interface{}) {
	if this.conn != nil {
		this.conn.SetProperty(key, value)
	}
}

func (this *ProxyClient) Close() {
	if this.conn != nil {
		this.conn.Close()
	}
}


func NewProxyClient(proxy string) *ProxyClient {
	return & ProxyClient{
		proxy: proxy,
	}
}

