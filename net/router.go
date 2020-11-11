package net

import (
	"go.uber.org/zap"
	"slgserver/log"
	"strings"
)

type HandlerFunc func(req *WsMsgReq, rsp *WsMsgRsp)
type MiddlewareFunc func(HandlerFunc) HandlerFunc

type Group struct {
	prefix     string
	hMap       map[string]HandlerFunc
	middleware []MiddlewareFunc
}

func (this*Group) AddRouter(name string, handlerFunc HandlerFunc) {
	this.hMap[name] = handlerFunc
}

func (this* Group) Use(middleware ...MiddlewareFunc) *Group{
	this.middleware = append(this.middleware, middleware...)
	return this
}

func (this*Group) applyMiddleware(name string) HandlerFunc {
	h, ok := this.hMap[name]
	if ok {
		for i := len(this.middleware) - 1; i >= 0; i-- {
			h = this.middleware[i](h)
		}
	}

	return h
}


func (this*Group) exec(name string, req *WsMsgReq, rsp *WsMsgRsp){
	h := this.applyMiddleware(name)
	if h == nil {
		log.DefaultLog.Warn("Group has not applyMiddleware",
			zap.String("msgName", req.Body.Name))
	}else{
		h(req, rsp)
	}
}

type Router struct {
	groups[] *Group
}

func (this*Router) Group(prefix string) *Group{
	g := &Group{prefix: prefix,
		hMap: make(map[string]HandlerFunc),
	}

	this.groups = append(this.groups, g)
	return g
}

func (this*Router) Run(req *WsMsgReq, rsp *WsMsgRsp) {
	name := req.Body.Name
	msgName := name
	sArr := strings.Split(name, ".")
	prefix := ""
	if len(sArr) == 2{
		prefix = sArr[0]
		msgName = sArr[1]
	}

	for _, g := range this.groups {
		if g.prefix == prefix{
			g.exec(msgName, req, rsp)
		}
	}
}
