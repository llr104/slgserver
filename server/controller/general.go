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
	g.AddRouter("dispose", this.dispose)
	g.AddRouter("armyList", this.armyList)

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
			rspObj.Generals[i].Order = v.Order
			rspObj.Generals[i].Cost = v.Cost
			rspObj.Generals[i].Speed = v.Speed
			rspObj.Generals[i].Defense = v.Defense
			rspObj.Generals[i].Strategy = v.Strategy
			rspObj.Generals[i].Force = v.Force
			rspObj.Generals[i].Name = v.Name
			rspObj.Generals[i].Id = v.Id
			rspObj.Generals[i].CfgId = v.CfgId
			rspObj.Generals[i].Destroy = v.Destroy
			rspObj.Generals[i].Level = v.Level
			rspObj.Generals[i].Exp = v.Exp
		}
	}else{
		rsp.Body.Code = constant.DBError
	}
}

//队伍列表
func (this*General) armyList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ArmyListReq{}
	rspObj := &proto.ArmyListRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	rspObj.CityId = reqObj.CityId

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	city,err := entity.RCMgr.Get(reqObj.CityId)
	if err != nil{
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	as, _ := entity.AMgr.GetByCity(reqObj.CityId)
	rspObj.Armys = make([]proto.Army, len(as))
	for i, v := range as {
		rspObj.Armys[i].Id = v.Id
		rspObj.Armys[i].Order = v.Order
		rspObj.Armys[i].CityId = v.CityId
		rspObj.Armys[i].FirstId = v.FirstId
		rspObj.Armys[i].SecondId = v.SecondId
		rspObj.Armys[i].ThirdId = v.ThirdId
		rspObj.Armys[i].FirstSoldierCnt = v.FirstSoldierCnt
		rspObj.Armys[i].SecondSoldierCnt = v.SecondSoldierCnt
		rspObj.Armys[i].ThirdSoldierCnt = v.ThirdSoldierCnt

	}
}

//配置武将
func (this*General) dispose(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

	reqObj := &proto.DisposeReq{}
	rspObj := &proto.DisposeRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if reqObj.Order<=0 || reqObj.Order>5 || reqObj.Position<0 || reqObj.Position>3{
		rsp.Body.Code = constant.InvalidParam
		return
	}

	city, err := entity.RCMgr.Get(reqObj.CityId)
	if err != nil {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	g, err := entity.GMgr.FindGeneral(reqObj.GeneralId)
	if err != nil{
		rsp.Body.Code = constant.GeneralNotFound
		return
	}

	if g.RId != role.RId{
		rsp.Body.Code = constant.GeneralNotMe
		return
	}

	a, err := entity.AMgr.GetOrCreate(role.RId, reqObj.CityId, reqObj.Order)
	if err != nil{
		rsp.Body.Code = constant.DBError
		return
	}

	//配置逻辑
	if reqObj.Position == 0{
		if a.FirstId == g.Id{
			a.FirstId = 0
			a.FirstSoldierCnt = 0
		}else if a.SecondId == g.Id{
			a.SecondId = 0
			a.SecondSoldierCnt = 0
		}else if a.ThirdId == g.Id{
			a.ThirdId = 0
			a.ThirdSoldierCnt = 0
		}
		g.Order = 0
		g.CityId = 0
	}else{
		if reqObj.Position == 1 {
			//旧的下阵
			if oldG, err := entity.GMgr.FindGeneral(a.ThirdId); err==nil{
				oldG.CityId = 0
				oldG.Order = 0
			}
			a.FirstSoldierCnt = 0
			a.FirstId = g.Id
		}else if reqObj.Position == 2 {
			//旧的下阵
			if oldG, err := entity.GMgr.FindGeneral(a.SecondId); err==nil{
				oldG.CityId = 0
				oldG.Order = 0
			}
			a.SecondSoldierCnt = 0
			a.SecondId = g.Id
		}else if reqObj.Position == 2 {
			//旧的下阵
			if oldG, err := entity.GMgr.FindGeneral(a.ThirdId); err==nil{
				oldG.CityId = 0
				oldG.Order = 0
			}
			a.ThirdSoldierCnt = 0
			a.ThirdId = g.Id
		}
		//新的上阵
		g.Order = reqObj.Position
		g.CityId = reqObj.CityId
	}

	a.NeedUpdate = true

	rspObj.Army.CityId = a.CityId
	rspObj.Army.Id = a.Id
	rspObj.Army.Order = a.Order
	rspObj.Army.FirstId = a.FirstId
	rspObj.Army.SecondId = a.SecondId
	rspObj.Army.ThirdId = a.ThirdId
	rspObj.Army.FirstSoldierCnt = a.FirstSoldierCnt
	rspObj.Army.SecondSoldierCnt = a.SecondSoldierCnt
	rspObj.Army.ThirdSoldierCnt = a.ThirdSoldierCnt
}
