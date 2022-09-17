package controller

import (
	"sync"

	"github.com/goinggo/mapstructure"
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/middleware"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/chatserver/logic"
	"github.com/llr104/slgserver/server/chatserver/proto"
	"github.com/llr104/slgserver/util"
	"go.uber.org/zap"
)

var DefaultChat = Chat{
	worldGroup:       logic.NewGroup(),
	unionGroups:      make(map[int]*logic.Group),
	ridToUnionGroups: make(map[int]int),
}

type Chat struct {
	unionMutex	sync.RWMutex
	worldGroup *logic.Group          //世界频道
	unionGroups map[int]*logic.Group //联盟频道
	ridToUnionGroups map[int]int     //rid对应的联盟频道
}

func (this*Chat) InitRouter(r *net.Router) {
	g := r.Group("chat").Use(middleware.ElapsedTime(), middleware.Log())

	g.AddRouter("login", this.login)
	g.AddRouter("logout", this.logout, middleware.CheckRId())
	g.AddRouter("chat", this.chat, middleware.CheckRId())
	g.AddRouter("history", this.history, middleware.CheckRId())
	g.AddRouter("join", this.join, middleware.CheckRId())
	g.AddRouter("exit", this.exit, middleware.CheckRId())
}

func (this*Chat) login(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.LoginReq{}
	rspObj := &proto.LoginRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj.RId = reqObj.RId
	rspObj.NickName = reqObj.NickName

	sess, err := util.ParseSession(reqObj.Token)
	if err != nil{
		rsp.Body.Code = constant.InvalidParam
		return
	}
	if sess.IsValid() == false || sess.Id != reqObj.RId{
		rsp.Body.Code = constant.InvalidParam
		return
	}
	net.ConnMgr.RoleEnter(req.Conn, reqObj.RId)

	this.worldGroup.Enter(logic.NewUser(reqObj.RId, reqObj.NickName))
}

func (this*Chat) logout(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.LogoutReq{}
	rspObj := &proto.LogoutRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj.RId = reqObj.RId

	net.ConnMgr.UserLogout(req.Conn)
	this.worldGroup.Exit(reqObj.RId)
}

func (this*Chat) chat(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ChatReq{}
	rspObj := &proto.ChatMsg{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)

	p, _ := req.Conn.GetProperty("rid")
	rid := p.(int)
	if reqObj.Type == 0 {
		//世界聊天
		rsp.Body.Msg = this.worldGroup.PutMsg(reqObj.Msg, rid, 0)
	}else if reqObj.Type == 1{
		//联盟聊天
		this.unionMutex.RLock()
		id, ok := this.ridToUnionGroups[rid]
		if ok {
			g, ok := this.unionGroups[id]
			if ok {
				g.PutMsg(reqObj.Msg, rid, 1)
			}else{
				log.DefaultLog.Warn("chat not found rid in unionGroups", zap.Int("rid", rid))
			}
		}else{
			log.DefaultLog.Warn("chat not found rid in ridToUnionGroups", zap.Int("rid", rid))
		}
		this.unionMutex.RUnlock()
	}

}

//历史记录
func (this*Chat) history(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.HistoryReq{}
	rspObj := &proto.HistoryRsp{}
	rsp.Body.Code = constant.OK

	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj.Msgs = []proto.ChatMsg{}
	p, _ := req.Conn.GetProperty("rid")
	rid := p.(int)

	if reqObj.Type == 0 {
		r := this.worldGroup.History()
		rspObj.Msgs = r
	}else if reqObj.Type == 1 {
		this.unionMutex.RLock()
		id, ok := this.ridToUnionGroups[rid]
		if ok {
			g, ok := this.unionGroups[id]
			if ok {
				rspObj.Msgs = g.History()
			}
		}
		this.unionMutex.RUnlock()
	}
	rspObj.Type = reqObj.Type
	rsp.Body.Msg = rspObj
}

func (this*Chat) join(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.JoinReq{}
	rspObj := &proto.JoinRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj
	rspObj.Type = reqObj.Type
	mapstructure.Decode(req.Body.Msg, reqObj)
	p, _ := req.Conn.GetProperty("rid")
	rid := p.(int)
	if reqObj.Type == 1 {
		u := this.worldGroup.GetUser(rid)
		if u == nil {
			rsp.Body.Code = constant.InvalidParam
			return
		}

		this.unionMutex.Lock()
		gId, ok := this.ridToUnionGroups[rid]
		if ok {
			if gId != reqObj.Id {
				//联盟聊天只能有一个，顶掉旧的
				if g,ok := this.unionGroups[gId]; ok {
					g.Exit(rid)
				}

				_, ok = this.unionGroups[reqObj.Id]
				if ok == false {
					this.unionGroups[reqObj.Id] = logic.NewGroup()
				}
				this.ridToUnionGroups[rid] = reqObj.Id
				this.unionGroups[reqObj.Id].Enter(u)
			}
		}else{
			//新加入
			_, ok = this.unionGroups[reqObj.Id]
			if ok == false {
				this.unionGroups[reqObj.Id] = logic.NewGroup()
			}
			this.ridToUnionGroups[rid] = reqObj.Id
			this.unionGroups[reqObj.Id].Enter(u)
		}
		this.unionMutex.Unlock()
	}
}

func (this*Chat) exit(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ExitReq{}
	rspObj := &proto.ExitRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj
	rspObj.Type = reqObj.Type
	mapstructure.Decode(req.Body.Msg, reqObj)
	p, _ := req.Conn.GetProperty("rid")
	rid := p.(int)

	if reqObj.Type == 1 {
		this.unionMutex.Lock()
		id, ok := this.ridToUnionGroups[rid]
		if ok {
			g, ok := this.unionGroups[id]
			if ok {
				g.Exit(rid)
			}
		}
		delete(this.ridToUnionGroups, rid)
		this.unionMutex.Unlock()
	}
}
