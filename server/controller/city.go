package controller

import (
	"encoding/json"
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/net"
	"slgserver/server/logic"
	"slgserver/server/middleware"
	"slgserver/server/model"
	"slgserver/server/proto"
	"slgserver/server/static_conf/facility"
)

var DefaultCity = City{

}
type City struct {

}

func (this*City) InitRouter(r *net.Router) {
	g := r.Group("city").Use(middleware.ElapsedTime(), middleware.Log(),
		middleware.CheckLogin(), middleware.CheckRole())

	g.AddRouter("facilities", this.facilities)
	g.AddRouter("upFacility", this.upFacility)
	g.AddRouter("upCity", this.upCity)

}

func (this*City) facilities(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.FacilitiesReq{}
	rspObj := &proto.FacilitiesRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	city, ok := logic.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	role := r.(*model.Role)
	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	f, ok := logic.RFMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	t := make([]logic.Facility, 0)
	json.Unmarshal([]byte(f.Facilities), &t)

	rspObj.Facilities = make([]proto.Facility, len(t))
	for i, v := range t {
		rspObj.Facilities[i].Name = v.Name
		rspObj.Facilities[i].Level = v.Level
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
	city, ok := logic.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	role := r.(*model.Role)
	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	_, ok = logic.RFMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	out, errCode := logic.RFMgr.UpFacility(role.RId ,reqObj.CityId, int8(reqObj.FType))
	rsp.Body.Code = errCode
	if errCode == constant.OK{
		rspObj.Facility.Level = out.Level
		rspObj.Facility.Type = out.Type
		rspObj.Facility.Name = out.Name

		//资源产量变化了
		oldValues := facility.FConf.GetValues(reqObj.FType, out.Level-1)
		newValues := facility.FConf.GetValues(reqObj.FType, out.Level)
		additions := facility.FConf.GetAdditions(reqObj.FType)

		roleRes, ok:= logic.RResMgr.Get(role.RId)
		if ok {
			for i, atype := range additions {
				if atype == facility.TypeWood{
					roleRes.WoodYield -= oldValues[i]
					roleRes.WoodYield += newValues[i]
				}else if atype == facility.TypeGrain{
					roleRes.GrainYield -= oldValues[i]
					roleRes.GrainYield += newValues[i]
				}else if atype == facility.TypeIron{
					roleRes.IronYield -= oldValues[i]
					roleRes.IronYield += newValues[i]
				}else if atype == facility.TypeStone{
					roleRes.StoneYield -= oldValues[i]
					roleRes.StoneYield += newValues[i]
				}else if atype == facility.TypeGold{
					roleRes.GoldYield -= oldValues[i]
					roleRes.GoldYield += newValues[i]
				}else if atype == facility.TypeWarehouseLimit {
					roleRes.DepotCapacity = newValues[i]
				}
				roleRes.SyncExecute()
			}
		}

		if roleRes, ok:= logic.RResMgr.Get(role.RId); ok {
			rspObj.RoleRes = roleRes.ToProto().(proto.RoleRes)
		}
	}

}

func (this*City) upCity(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpCityReq{}
	rspObj := &proto.UpCityRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.CityId = reqObj.CityId
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	city, ok := logic.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	role := r.(*model.Role)
	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	_, ok = logic.RFMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

}