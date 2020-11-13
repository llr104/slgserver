package controller

import (
	"encoding/json"
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/model"
	"slgserver/net"
	"slgserver/server/entity"
	"slgserver/server/middleware"
	"slgserver/server/proto"
)

var DefaultCity = City{

}
type City struct {

}

func (this*City) InitRouter(r *net.Router) {
	g := r.Group("role").Use(middleware.Log(), middleware.CheckLogin())
	g.AddRouter("facilities", this.facilities)
}

func (this*City) facilities(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.FacilitiesReq{}
	rspObj := &proto.FacilitiesRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK

	r, err := req.Conn.GetProperty("role")
	if err != nil {
		rsp.Body.Code = constant.InvalidParam
		return
	}

	city, err := entity.RCMgr.Get(reqObj.CityId)
	if err != nil {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	role := r.(*model.Role)
	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	f, err := entity.RFMgr.Get(reqObj.CityId)
	if err != nil {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	d,_ := json.Marshal(f)
	rspObj.Facilities = string(d)

}