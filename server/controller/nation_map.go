package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/net"
	"slgserver/server/logic"
	"slgserver/server/middleware"
	"slgserver/server/model"
	"slgserver/server/proto"
	"slgserver/server/static_conf"
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

	m := static_conf.MapBuildConf.Cfg
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
		logic.RoleBuildExtra(v)
		rspObj.MRBuilds[i] = v.ToProto().(proto.MapRoleBuild)
	}

	cb := logic.RCMgr.Scan(x, y)
	rspObj.MCBuilds = make([]proto.MapRoleCity, len(cb))
	for i, v := range cb {
		rspObj.MCBuilds[i] = v.ToProto().(proto.MapRoleCity)
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
		rspObj.MRBuilds[i] = v.ToProto().(proto.MapRoleBuild)
	}

	cb := logic.RCMgr.ScanBlock(x, y, reqObj.Length)
	rspObj.MCBuilds = make([]proto.MapRoleCity, len(cb))
	for i, v := range cb {
		rspObj.MCBuilds[i] = v.ToProto().(proto.MapRoleCity)
	}

	armys := logic.ArmyLogic.ScanBlock(x, y, reqObj.Length)
	rspObj.Armys = make([]proto.Army, len(armys))
	for i, v := range armys {
		rspObj.Armys[i] = v.ToProto().(proto.Army)
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

	if logic.RBMgr.BuildIsRId(x, y, role.RId) == false{
		rsp.Body.Code = constant.BuildNotMe
		return
	}

	rb, _ := logic.RBMgr.PositionBuild(x, y)
	rr, ok := logic.RResMgr.CutDown(role.RId, rb)
	logic.RBMgr.RemoveFromRole(rb)

	//移除该地驻守
	logic.ArmyLogic.GiveUp(logic.ToPosition(reqObj.X, reqObj.Y))

	if ok {
		rspObj.RoleRes = rr.ToProto().(proto.RoleRes)
		rsp.Body.Code = constant.OK
	}else{
		rsp.Body.Code = constant.DBError
	}

}
