package controller

import (
	"github.com/goinggo/mapstructure"
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/middleware"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/logic"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
)

var DefaultMap = NationMap{}

type NationMap struct {
}

func (this *NationMap) InitRouter(r *net.Router) {
	g := r.Group("nationMap").Use(middleware.ElapsedTime(), middleware.Log())
	g.AddRouter("config", this.config)
	g.AddRouter("scan", this.scan, middleware.CheckRole())
	g.AddRouter("scanBlock", this.scanBlock, middleware.CheckRole())
	g.AddRouter("giveUp", this.giveUp, middleware.CheckRole())
	g.AddRouter("build", this.build, middleware.CheckRole())
	g.AddRouter("upBuild", this.upBuild, middleware.CheckRole())
	g.AddRouter("delBuild", this.delBuild, middleware.CheckRole())

}

/*
获取配置
*/
func (this *NationMap) config(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
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
func (this *NationMap) scan(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ScanReq{}
	rspObj := &proto.ScanRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	x := reqObj.X
	y := reqObj.Y

	rb := mgr.RBMgr.Scan(x, y)
	rspObj.MRBuilds = make([]proto.MapRoleBuild, len(rb))
	for i, v := range rb {
		rspObj.MRBuilds[i] = v.ToProto().(proto.MapRoleBuild)
	}

	cb := mgr.RCMgr.Scan(x, y)
	rspObj.MCBuilds = make([]proto.MapRoleCity, len(cb))
	for i, v := range cb {
		rspObj.MCBuilds[i] = v.ToProto().(proto.MapRoleCity)
	}
}

func (this *NationMap) scanBlock(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ScanBlockReq{}
	rspObj := &proto.ScanRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	x := reqObj.X
	y := reqObj.Y

	rb := mgr.RBMgr.ScanBlock(x, y, reqObj.Length)
	rspObj.MRBuilds = make([]proto.MapRoleBuild, len(rb))
	for i, v := range rb {
		rspObj.MRBuilds[i] = v.ToProto().(proto.MapRoleBuild)
	}

	cb := mgr.RCMgr.ScanBlock(x, y, reqObj.Length)
	rspObj.MCBuilds = make([]proto.MapRoleCity, len(cb))
	for i, v := range cb {
		rspObj.MCBuilds[i] = v.ToProto().(proto.MapRoleCity)
	}

	armys := logic.ArmyLogic.ScanBlock(role.RId, x, y, reqObj.Length)
	rspObj.Armys = make([]proto.Army, len(armys))
	for i, v := range armys {
		rspObj.Armys[i] = v.ToProto().(proto.Army)
	}

}

func (this *NationMap) giveUp(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
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

	if mgr.RBMgr.BuildIsRId(x, y, role.RId) == false {
		rsp.Body.Code = constant.BuildNotMe
		return
	}

	rsp.Body.Code = mgr.RBMgr.GiveUp(x, y)
}

//建造
func (this *NationMap) build(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.BuildReq{}
	rspObj := &proto.BuildRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	x := reqObj.X
	y := reqObj.Y

	rspObj.X = x
	rspObj.Y = y

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if mgr.RBMgr.BuildIsRId(x, y, role.RId) == false {
		rsp.Body.Code = constant.BuildNotMe
		return
	}

	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok == false {
		rsp.Body.Code = constant.BuildNotMe
		return
	}

	if b.IsResBuild() == false || b.IsBusy() {
		rsp.Body.Code = constant.CanNotBuildNew
		return
	}

	cnt := mgr.RBMgr.RoleFortressCnt(role.RId)
	if cnt >= static_conf.Basic.Build.FortressLimit {
		rsp.Body.Code = constant.CanNotBuildNew
		return
	}

	cfg, ok := static_conf.MapBCConf.BuildConfig(reqObj.Type, 1)
	if ok == false {
		rsp.Body.Code = constant.InvalidParam
		return
	}

	code := mgr.RResMgr.TryUseNeed(role.RId, cfg.Need)
	if code != constant.OK {
		rsp.Body.Code = code
		return
	}

	b.BuildOrUp(*cfg)
	b.SyncExecute()

}

func (this *NationMap) upBuild(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpBuildReq{}
	rspObj := &proto.UpBuildRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	x := reqObj.X
	y := reqObj.Y

	rspObj.X = x
	rspObj.Y = y

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if mgr.RBMgr.BuildIsRId(x, y, role.RId) == false {
		rsp.Body.Code = constant.BuildNotMe
		return
	}

	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok == false {
		rsp.Body.Code = constant.BuildNotMe
		return
	}

	if b.IsHaveModifyLVAuth() == false || b.IsInGiveUp() || b.IsBusy() {
		rsp.Body.Code = constant.CanNotUpBuild
		return
	}

	cfg, ok := static_conf.MapBCConf.BuildConfig(b.Type, b.Level+1)
	if ok == false {
		rsp.Body.Code = constant.InvalidParam
		return
	}

	code := mgr.RResMgr.TryUseNeed(role.RId, cfg.Need)
	if code != constant.OK {
		rsp.Body.Code = code
		return
	}
	b.BuildOrUp(*cfg)
	b.SyncExecute()
	rspObj.Build = b.ToProto().(proto.MapRoleBuild)
}

func (this *NationMap) delBuild(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpBuildReq{}
	rspObj := &proto.UpBuildRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	x := reqObj.X
	y := reqObj.Y

	rspObj.X = x
	rspObj.Y = y

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if mgr.RBMgr.BuildIsRId(x, y, role.RId) == false {
		rsp.Body.Code = constant.BuildNotMe
		return
	}

	rsp.Body.Code = mgr.RBMgr.Destroy(x, y)
	b, ok := mgr.RBMgr.PositionBuild(x, y)
	if ok {
		rspObj.Build = b.ToProto().(proto.MapRoleBuild)
	}
}
