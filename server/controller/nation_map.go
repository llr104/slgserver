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

	bb := entity.NMMgr.Scan(x, y)
	rspObj.BBuilds = make([]proto.BaseBuild, len(bb))
	for i, v := range bb {
		rspObj.BBuilds[i].X = v.X
		rspObj.BBuilds[i].Y = v.Y
		rspObj.BBuilds[i].Id = v.Id
		rspObj.BBuilds[i].Type = v.Type
		rspObj.BBuilds[i].Durable = 100
		rspObj.BBuilds[i].Level = v.Level
	}

	rb := entity.RBMgr.Scan(x, y)
	rspObj.RBuilds = make([]proto.RoleBuild, len(rb))
	for i, v := range rb {
		name := ""
		vRole, err := entity.RMgr.Get(v.RId)
		if err == nil {
			name = vRole.NickName
		}

		rspObj.RBuilds[i].X = v.X
		rspObj.RBuilds[i].Y = v.Y
		rspObj.RBuilds[i].Id = v.Id
		rspObj.RBuilds[i].Type = v.Type
		rspObj.RBuilds[i].Durable = v.Durable
		rspObj.RBuilds[i].Level = v.Level
		rspObj.RBuilds[i].RId = v.RId
		rspObj.RBuilds[i].Name = v.Name
		rspObj.RBuilds[i].Defender = v.Defender
		rspObj.RBuilds[i].RNick = name
	}

	cb := entity.RCMgr.Scan(x, y)
	rspObj.CBuilds = make([]proto.RoleCity, len(rb))
	for i, v := range cb {
		rspObj.CBuilds[i].X = v.X
		rspObj.CBuilds[i].Y = v.Y
		rspObj.CBuilds[i].CityId = v.CityId
		rspObj.CBuilds[i].Durable = v.Durable
		rspObj.CBuilds[i].Level = v.Level
		rspObj.CBuilds[i].RId = v.RId
		rspObj.CBuilds[i].Name = v.Name
		rspObj.CBuilds[i].IsMain = v.IsMain == 1
	}
}