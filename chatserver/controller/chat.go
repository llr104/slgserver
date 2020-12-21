package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/chatserver/logic"
	"slgserver/chatserver/proto"
	"slgserver/constant"
	"slgserver/net"
	"slgserver/server/conn"
	"slgserver/server/middleware"
)

var DefaultChat = Chat{
	worldGroup: logic.NewGroup(),
}

type Chat struct {
	worldGroup *logic.Group	//世界频道
}

func (this*Chat) InitRouter(r *net.Router) {
	g := r.Group("chat").Use(middleware.ElapsedTime(), middleware.Log())

	g.AddRouter("login", this.login)
	g.AddRouter("logout", this.logout)
	g.AddRouter("chat", this.chat)
	g.AddRouter("history", this.history)
}

func (this*Chat) login(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.LoginReq{}
	rspObj := &proto.LoginRsp{}
	rsp.Body.Code = constant.OK
	rsp.Body.Msg = rspObj

	mapstructure.Decode(req.Body.Msg, reqObj)
	rspObj.RId = reqObj.RId
	rspObj.NickName = reqObj.NickName

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
	}
}
