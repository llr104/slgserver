package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/model"
	"slgserver/net"
	"slgserver/server/logic"
	"slgserver/server/middleware"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
)

var DefaultMap = NationMap{}

type NationMap struct {

}

func (this*NationMap) InitRouter(r *net.Router) {
	g := r.Group("nationMap").Use(middleware.ElapsedTime(), middleware.Log())
	g.AddRouter("config", this.config)
	g.AddRouter("scan", this.scan)
	g.AddRouter("scanBlock", this.scanBlock)
	g.AddRouter("giveUp", this.giveUp, middleware.CheckRole())
}

/*
获取配置
*/
func (this*NationMap) config(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ConfigReq{}
	rspObj := &proto.ConfigRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	m := logic.BCMgr.Maps()
	rspObj.Confs = make([]proto.Conf, len(m))
	i := 0
	for _, v := range m {
		rspObj.Confs[i].Type = v.Type
		rspObj.Confs[i].Name = v.Name
		rspObj.Confs[i].Level = v.Level
		rspObj.Confs[i].Defender = v.Defender
		rspObj.Confs[i].Durable = v.Durable
		rspObj.Confs[i].Grain = v.Grain
		rspObj.Confs[i].Iron = v.Iron
		rspObj.Confs[i].Stone = v.Stone
		rspObj.Confs[i].Wood = v.Wood
		i++
	}
}

/*
扫描地图
*/
func (this*NationMap) scan(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ScanReq{}
	rspObj := &proto.ScanRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	x := reqObj.X
	y := reqObj.Y

	rb := logic.RBMgr.Scan(x, y)
	rspObj.MRBuilds = make([]proto.MapRoleBuild, len(rb))
	for i, v := range rb {
		name := ""
		vRole, ok := logic.RMgr.Get(v.RId)
		if ok {
			name = vRole.NickName
		}
		model_to_proto.MRBuild(v, &rspObj.MRBuilds[i])
		rspObj.MRBuilds[i].RNick = name
	}

	cb := logic.RCMgr.Scan(x, y)
	rspObj.MCBuilds = make([]proto.MapRoleCity, len(cb))
	for i, v := range cb {
		model_to_proto.MCBuild(v, &rspObj.MCBuilds[i])
	}
}

func (this*NationMap) scanBlock(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ScanBlockReq{}
	rspObj := &proto.ScanRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	x := reqObj.X
	y := reqObj.Y

	rb := logic.RBMgr.ScanBlock(x, y, reqObj.Length)
	rspObj.MRBuilds = make([]proto.MapRoleBuild, len(rb))
	for i, v := range rb {
		name := ""
		vRole, ok := logic.RMgr.Get(v.RId)
		if ok {
			name = vRole.NickName
		}

		model_to_proto.MRBuild(v, &rspObj.MRBuilds[i])
		rspObj.MRBuilds[i].RNick = name
	}

	cb := logic.RCMgr.ScanBlock(x, y, reqObj.Length)
	rspObj.MCBuilds = make([]proto.MapRoleCity, len(cb))
	for i, v := range cb {
		model_to_proto.MCBuild(v, &rspObj.MCBuilds[i])
	}
}

func (this*NationMap) giveUp(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.GiveUpReq{}
	rspObj := &proto.GiveUpRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	x := reqObj.X
	y := reqObj.Y

	rspObj.X = x
	rspObj.Y = y

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	rb, ok := logic.RBMgr.PositionBuild(x, y)
	if ok == false{
		rsp.Body.Code = constant.CannotGiveUp
		return
	}

	if rb.RId != role.RId{
		rsp.Body.Code = constant.BuildNotMe
		return
	}

	rr, ok := logic.RResMgr.CutDown(role.RId, rb)
	logic.RBMgr.RemoveFromRole(rb)

	if ok {
		model_to_proto.RRes(rr, &rspObj.RoleRes)
		rsp.Body.Code = constant.OK
	}else{
		rsp.Body.Code = constant.DBError
	}

}
