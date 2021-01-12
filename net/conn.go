package net

import (
	"context"
	"time"
)

type ReqBody struct {
	Seq     int64		`json:"seq"`
	Name 	string 		`json:"name"`
	Msg		interface{}	`json:"msg"`
	Proxy	string		`json:"proxy"`
}

type RspBody struct {
	Seq     int64		`json:"seq"`
	Name 	string 		`json:"name"`
	Code	int			`json:"code"`
	Msg		interface{}	`json:"msg"`
}

type WsMsgReq struct {
	Body	*ReqBody
	Conn	WSConn
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

type WSConn interface {
	SetProperty(key string, value interface{})
	GetProperty(key string) (interface{}, error)
	RemoveProperty(key string)
	Addr() string
	Push(name string, data interface{})
}

type syncCtx struct {
	ctx 		context.Context
	cancel		context.CancelFunc
	outChan    	chan *RspBody
}

func newSyncCtx() *syncCtx {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	return &syncCtx{ctx: ctx, cancel: cancel, outChan: make(chan *RspBody)}
}

func (this* syncCtx) wait() *RspBody{
	defer this.cancel()
	select {
	case data := <- this.outChan:
		return data
	case <-this.ctx.Done():
		return nil
	}
}