package controller

import (
	"github.com/goinggo/mapstructure"
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/middleware"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/proto"
)

var DefaultCity = City{}

type City struct {
}

func (this *City) InitRouter(r *net.Router) {
	g := r.Group("city").Use(middleware.ElapsedTime(), middleware.Log(),
		middleware.CheckLogin(), middleware.CheckRole())

	g.AddRouter("facilities", this.facilities)
	g.AddRouter("upFacility", this.upFacility)

}

func (this *City) facilities(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.FacilitiesReq{}
	rspObj := &proto.FacilitiesRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	city, ok := mgr.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	role := r.(*model.Role)
	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	f, ok := mgr.RFMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	t := f.Facility()

	rspObj.Facilities = make([]proto.Facility, len(t))
	for i, v := range t {
		rspObj.Facilities[i].Name = v.Name
		rspObj.Facilities[i].Level = v.GetLevel()
		rspObj.Facilities[i].Type = v.Type
		rspObj.Facilities[i].UpTime = v.UpTime
	}

}

//升级需要耗费时间，为了减少定时任务，升级这里做成被动触发产生，不做定时任务
func (this *City) upFacility(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpFacilityReq{}
	rspObj := &proto.UpFacilityRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	city, ok := mgr.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	role := r.(*model.Role)
	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	_, ok = mgr.RFMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	out, errCode := mgr.RFMgr.UpFacility(role.RId, reqObj.CityId, reqObj.FType)
	rsp.Body.Code = errCode
	if errCode == constant.OK {
		rspObj.Facility.Level = out.GetLevel()
		rspObj.Facility.Type = out.Type
		rspObj.Facility.Name = out.Name
		rspObj.Facility.UpTime = out.UpTime

		if roleRes, ok := mgr.RResMgr.Get(role.RId); ok {
			rspObj.RoleRes = roleRes.ToProto().(proto.RoleRes)
		}
	}

}
