package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/net"
	"slgserver/server/logic/mgr"
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

	out, errCode := mgr.RFMgr.UpFacility(role.RId ,reqObj.CityId, int8(reqObj.FType))
	rsp.Body.Code = errCode
	if errCode == constant.OK{
		rspObj.Facility.Level = out.Level
		rspObj.Facility.Type = out.Type
		rspObj.Facility.Name = out.Name

		//资源产量变化了
		oldValues := facility.FConf.GetValues(reqObj.FType, out.Level-1)
		newValues := facility.FConf.GetValues(reqObj.FType, out.Level)
		additions := facility.FConf.GetAdditions(reqObj.FType)

		roleRes, ok:= mgr.RResMgr.Get(role.RId)
		if ok {
			for i, atype := range additions {
				if atype == facility.TypeWood{
					if len(oldValues) > i{
						roleRes.WoodYield -= oldValues[i]
					}
					roleRes.WoodYield += newValues[i]
				}else if atype == facility.TypeGrain{
					if len(oldValues) > i{
						roleRes.GrainYield -= oldValues[i]
					}
					roleRes.GrainYield += newValues[i]
				}else if atype == facility.TypeIron{
					if len(oldValues) > i{
						roleRes.IronYield -= oldValues[i]
					}
					roleRes.IronYield += newValues[i]
				}else if atype == facility.TypeStone{
					if len(oldValues) > i{
						roleRes.StoneYield -= oldValues[i]
					}
					roleRes.StoneYield += newValues[i]
				}else if atype == facility.TypeTax{
					if len(oldValues) > i{
						roleRes.GoldYield -= oldValues[i]
					}
					roleRes.GoldYield += newValues[i]
				}else if atype == facility.TypeWarehouseLimit {
					roleRes.DepotCapacity = newValues[i]
				}else if atype == facility.TypeCost {
					if len(oldValues) > i{
						city.Cost -= int8(oldValues[i])
					}
					city.Cost += int8(newValues[i])
				}else if atype == facility.TypeDurable{
					if len(oldValues) > i{
						city.MaxDurable -= oldValues[i]
						city.CurDurable -= oldValues[i]
					}
					city.MaxDurable += newValues[i]
					city.CurDurable += newValues[i]
				}

				city.SyncExecute()
				roleRes.SyncExecute()
			}
		}

		if roleRes, ok:= mgr.RResMgr.Get(role.RId); ok {
			rspObj.RoleRes = roleRes.ToProto().(proto.RoleRes)
		}
	}

}

func (this*City) upCity(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpCityReq{}
	rspObj := &proto.UpCityRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
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

	maxLevel := facility.FConf.MaxLevel(facility.Main)
	if city.Level >= maxLevel{
		rsp.Body.Code = constant.UpError
		return
	}

	needRes, ok := facility.FConf.Need(facility.Main, city.Level+1)
	if ok == false{
		rsp.Body.Code = constant.UpError
		return
	}

	ok = mgr.RResMgr.TryUseNeed(role.RId, needRes)
	if ok == false{
		rsp.Body.Code = constant.UpError
		return
	}


	oldValues := facility.FConf.GetValues(facility.Main, city.Level)
	newValues := facility.FConf.GetValues(facility.Main, city.Level+1)
	additions := facility.FConf.GetAdditions(facility.Main)

	for i, atype := range additions {
		if atype == facility.TypeDurable{
			if len(oldValues) > i{
				city.MaxDurable -= oldValues[i]
				city.CurDurable -= oldValues[i]
			}

			city.MaxDurable += newValues[i]
			city.CurDurable += newValues[i]
		}else if atype == facility.TypeCost{
			if len(oldValues) > i{
				city.Cost -= int8(oldValues[i])
			}
			city.Cost += int8(newValues[i])
		}
	}
	city.Level += 1
	city.SyncExecute()

	rspObj.City = city.ToProto().(proto.MapRoleCity)

}