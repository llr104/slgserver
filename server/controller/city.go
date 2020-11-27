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
		if facility.FPRC.IsContain(reqObj.FType){
			oldP, ok1 := facility.FPRC.GetProduce(reqObj.FType, int(out.Level-1))
			newP, ok2 := facility.FPRC.GetProduce(reqObj.FType, int(out.Level))
			roleRes, ok:= logic.RResMgr.Get(role.RId)
			if ok {
				if facility.FPRC.MF.Type == reqObj.FType{
					if ok1 && ok2{
						roleRes.GrainYield -= oldP.Yield
						roleRes.GrainYield += newP.Yield
					}
				}else if facility.FPRC.FMC.Type == reqObj.FType {
					if ok1 && ok2{
						roleRes.WoodYield -= oldP.Yield
						roleRes.WoodYield += newP.Yield
					}
				}else if facility.FPRC.LTC.Type == reqObj.FType {
					if ok1 && ok2{
						roleRes.IronYield -= oldP.Yield
						roleRes.IronYield += newP.Yield
					}
				}else if facility.FPRC.CSC.Type == reqObj.FType {
					if ok1 && ok2{
						roleRes.StoneYield -= oldP.Yield
						roleRes.StoneYield += newP.Yield
					}
				}else if facility.FPRC.MJ.Type == reqObj.FType {
					if ok1 && ok2{
						roleRes.GoldYield -= oldP.Yield
						roleRes.GoldYield += newP.Yield
					}
				}
				roleRes.Execute()
			}
		}else if facility.FWareHouse.IsContain(reqObj.FType){
			if roleRes, ok:= logic.RResMgr.Get(role.RId); ok {
				limit := facility.FWareHouse.Limit(int(out.Level))
				roleRes.DepotCapacity = limit
			}
		}

		if roleRes, ok:= logic.RResMgr.Get(role.RId); ok {
			rspObj.RoleRes = roleRes.ToProto().(proto.RoleRes)
		}
	}

}