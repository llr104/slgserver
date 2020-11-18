package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/model"
	"slgserver/net"
	"slgserver/server/entity"
	"slgserver/server/middleware"
	"slgserver/server/proto"
)

var DefaultGeneral = General{

}

type General struct {

}


func (this*General) InitRouter(r *net.Router) {
	g := r.Group("general").Use(middleware.Log(),
		middleware.CheckLogin(),
		middleware.CheckRole())

	g.AddRouter("myGenerals", this.myGenerals)

}

func (this*General) myGenerals(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyGeneralReq{}
	rspObj := &proto.MyGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	gs, err := entity.GMgr.GetAndTryCreate(role.RId)
	if err == nil {
		rsp.Body.Code = constant.OK
		rspObj.Generals = make([]proto.General, len(gs))
		for i, v := range gs {
			rspObj.Generals[i].CityId = v.CityId
			rspObj.Generals[i].ArmyId = v.ArmyId
			rspObj.Generals[i].Cost = v.Cost
			rspObj.Generals[i].Speed = v.Speed
			rspObj.Generals[i].Defense = v.Defense
			rspObj.Generals[i].Strategy = v.Strategy
			rspObj.Generals[i].Force = v.Force
			rspObj.Generals[i].Name = v.Name
			rspObj.Generals[i].Id = v.Id
			rspObj.Generals[i].CfgId = v.CfgId
			rspObj.Generals[i].Destroy = v.Destroy
		}
	}else{
		rsp.Body.Code = constant.DBError
	}


}