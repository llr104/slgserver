package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/chatserver/logic"
	"slgserver/chatserver/proto"
	"slgserver/constant"
	"slgserver/net"
	"slgserver/server/conn"
	"slgserver/server/middleware"
	"slgserver/util"
	"sync"
)

var DefaultChat = Chat{
	worldGroup: logic.NewGroup(),
	unionGroups: make(map[int]*logic.Group),
}

type Chat struct {
	unionMutex	sync.RWMutex
	worldGroup *logic.Group				//世界频道
	unionGroups map[int]*logic.Group	//联盟频道
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

	conn.ConnMgr.RoleEnter(req.Conn, reqObj.RId)
	this.worldGroup.Enter(logic.NewUser(reqObj.RId, reqObj.NickName))
}


func (this*Chat) logout(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.LogoutReq{}
	rspObj := &proto.LogoutRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj.RId = reqObj.RId

	conn.ConnMgr.UserLogout(req.Conn)
	this.worldGroup.Exit(reqObj.RId)
}

func (this*Chat) chat(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ChatReq{}
	rspObj := &proto.ChatMsg{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)

	rid, err := req.Conn.GetProperty("rid")
	if err == nil {
		if reqObj.Type == 0 {
			rsp.Body.Msg = this.worldGroup.PutMsg(reqObj.Msg, rid.(int))
		}
	}

}

//历史记录
func (this*Chat) history(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.HistoryReq{}
	rspObj := &proto.HistoryRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj
	rspObj.Type = reqObj.Type
	mapstructure.Decode(req.Body.Msg, reqObj)

	if reqObj.Type == 0 {
		r := this.worldGroup.History()
		rspObj.Msgs = r
	}else if reqObj.Type == 1 {
		this.unionMutex.RLock()
		g, ok := this.unionGroups[reqObj.Id]
		this.unionMutex.RUnlock()
		if ok {
			rspObj.Msgs = g.History()
		}
	}
}

func (this*Chat) join(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.JoinReq{}
	rspObj := &proto.JoinRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj
	rspObj.Type = reqObj.Type
	mapstructure.Decode(req.Body.Msg, reqObj)
	rid, _ := req.Conn.GetProperty("rid")

	if reqObj.Type == 1 {
		u := this.worldGroup.GetUser(rid.(int))
		if u == nil {
			rsp.Body.Code = constant.InvalidParam
			return
		}

		this.unionMutex.Lock()
		_, ok := this.unionGroups[reqObj.Id]
		if ok == false {
			this.unionGroups[reqObj.Id] = logic.NewGroup()
		}
		this.unionGroups[reqObj.Id].Enter(u)
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
	rid, _ := req.Conn.GetProperty("rid")

	if reqObj.Type == 1 {

		this.unionMutex.RLock()
		_, ok := this.unionGroups[reqObj.Id]
		if ok {
			this.unionGroups[reqObj.Id].Exit(rid.(int))
		}
		this.unionMutex.RUnlock()
	}
}
