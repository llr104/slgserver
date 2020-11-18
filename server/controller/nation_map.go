package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/net"
	"slgserver/server/entity"
	"slgserver/server/middleware"
	"slgserver/server/proto"
)

var DefaultMap = NationMap{}

type NationMap struct {

}

func (this*NationMap) InitRouter(r *net.Router) {
	g := r.Group("nationMap").Use(middleware.Log())
	g.AddRouter("config", this.config)
	g.AddRouter("scan", this.scan)
	g.AddRouter("scanBlock", this.scanBlock)
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

	m := entity.BCMgr.Maps()
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

	rb := entity.RBMgr.Scan(x, y)
	rspObj.MRBuilds = make([]proto.MapRoleBuild, len(rb))
	for i, v := range rb {
		name := ""
		vRole, err := entity.RMgr.Get(v.RId)
		if err == nil {
			name = vRole.NickName
		}

		rspObj.MRBuilds[i].X = v.X
		rspObj.MRBuilds[i].Y = v.Y
		rspObj.MRBuilds[i].Type = v.Type
		rspObj.MRBuilds[i].CurDurable = v.CurDurable
		rspObj.MRBuilds[i].MaxDurable = v.MaxDurable
		rspObj.MRBuilds[i].Level = v.Level
		rspObj.MRBuilds[i].RId = v.RId
		rspObj.MRBuilds[i].Name = v.Name
		rspObj.MRBuilds[i].Defender = v.Defender
		rspObj.MRBuilds[i].RNick = name
	}

	cb := entity.RCMgr.Scan(x, y)
	rspObj.MCBuilds = make([]proto.MapRoleCity, len(cb))
	for i, v := range cb {
		rspObj.MCBuilds[i].X = v.X
		rspObj.MCBuilds[i].Y = v.Y
		rspObj.MCBuilds[i].CityId = v.CityId
		rspObj.MCBuilds[i].CurDurable = v.CurDurable
		rspObj.MCBuilds[i].MaxDurable = v.MaxDurable
		rspObj.MCBuilds[i].Level = v.Level
		rspObj.MCBuilds[i].RId = v.RId
		rspObj.MCBuilds[i].Name = v.Name
		rspObj.MCBuilds[i].IsMain = v.IsMain == 1
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

	rb := entity.RBMgr.ScanBlock(x, y, reqObj.Length)
	rspObj.MRBuilds = make([]proto.MapRoleBuild, len(rb))
	for i, v := range rb {
		name := ""
		vRole, err := entity.RMgr.Get(v.RId)
		if err == nil {
			name = vRole.NickName
		}

		rspObj.MRBuilds[i].X = v.X
		rspObj.MRBuilds[i].Y = v.Y
		rspObj.MRBuilds[i].Type = v.Type
		rspObj.MRBuilds[i].CurDurable = v.CurDurable
		rspObj.MRBuilds[i].MaxDurable = v.MaxDurable
		rspObj.MRBuilds[i].Level = v.Level
		rspObj.MRBuilds[i].RId = v.RId
		rspObj.MRBuilds[i].Name = v.Name
		rspObj.MRBuilds[i].Defender = v.Defender
		rspObj.MRBuilds[i].RNick = name
	}

	cb := entity.RCMgr.ScanBlock(x, y, reqObj.Length)
	rspObj.MCBuilds = make([]proto.MapRoleCity, len(cb))
	for i, v := range cb {
		rspObj.MCBuilds[i].X = v.X
		rspObj.MCBuilds[i].Y = v.Y
		rspObj.MCBuilds[i].CityId = v.CityId
		rspObj.MCBuilds[i].CurDurable = v.CurDurable
		rspObj.MCBuilds[i].MaxDurable = v.MaxDurable
		rspObj.MCBuilds[i].Level = v.Level
		rspObj.MCBuilds[i].RId = v.RId
		rspObj.MCBuilds[i].Name = v.Name
		rspObj.MCBuilds[i].IsMain = v.IsMain == 1
	}
}

