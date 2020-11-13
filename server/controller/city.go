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
	g := r.Group("city").Use(middleware.Log(),
		middleware.CheckLogin(),
		middleware.CheckRole())

	g.AddRouter("facilities", this.facilities)
	g.AddRouter("upFacility", this.upFacility)

}

func (this*City) facilities(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.FacilitiesReq{}
	rspObj := &proto.FacilitiesRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
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

	t := make([]entity.Facility, 0)
	json.Unmarshal([]byte(f.Facilities), &t)

	rspObj.Facilities = make([]proto.Facility, len(t))
	for i, v := range t {
		rspObj.Facilities[i].Name = v.Name
		rspObj.Facilities[i].CLevel = v.CLevel
		rspObj.Facilities[i].MLevel = v.MLevel
		rspObj.Facilities[i].Type = v.Type
	}

}

func (this*City) upFacility(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpFacilityReq{}
	rspObj := &proto.UpFacilityRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
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

	_, err = entity.RFMgr.Get(reqObj.CityId)
	if err != nil {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	out, err := entity.RFMgr.UpFacility(reqObj.CityId, int8(reqObj.FType))
	if err != nil{
		rsp.Body.Code = constant.UpError
	}else{
		rspObj.Facility.CLevel = out.CLevel
		rspObj.Facility.MLevel = out.MLevel
		rspObj.Facility.Type = out.Type
		rspObj.Facility.Name = out.Name
	}

}