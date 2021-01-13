package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/middleware"
	"slgserver/net"
	"slgserver/server/slgserver/logic/mgr"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/proto"
	"slgserver/server/slgserver/static_conf"
)

var DefaultGeneral = General{

}

type General struct {

}


func (this*General) InitRouter(r *net.Router) {
	g := r.Group("general").Use(middleware.ElapsedTime(), middleware.Log(),
		middleware.CheckLogin(), middleware.CheckRole())

	g.AddRouter("myGenerals", this.myGenerals)
	g.AddRouter("drawGeneral", this.drawGenerals)
	g.AddRouter("composeGeneral", this.ComposeGeneral)
	g.AddRouter("addPrGeneral", this.AddPrGeneral)

}

func (this*General) myGenerals(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyGeneralReq{}
	rspObj := &proto.MyGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	gs, ok := mgr.GMgr.GetByRIdTryCreate(role.RId)
	if ok {
		rsp.Body.Code = constant.OK
		rspObj.Generals = make([]proto.General, len(gs))
		for i, v := range gs {
			rspObj.Generals[i] = v.ToProto().(proto.General)
		}
	}else{
		rsp.Body.Code = constant.DBError
	}
}


func (this*General) drawGenerals(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.DrawGeneralReq{}
	rspObj := &proto.DrawGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	cost := static_conf.Basic.General.DrawGeneralCost * reqObj.DrawTimes
	ok := mgr.RResMgr.GoldIsEnough(role.RId,cost)
	if ok == false{
		rsp.Body.Code = constant.GoldNotEnough
		return
	}

	limit := static_conf.Basic.General.Limit
	cnt := mgr.GMgr.ActiveCount(role.RId)
	if cnt + reqObj.DrawTimes > limit{
		rsp.Body.Code = constant.OutGeneralLimit
		return
	}

	gs, ok := mgr.GMgr.RandCreateGeneral(role.RId,reqObj.DrawTimes)

	if ok {
		mgr.RResMgr.TryUseGold(role.RId, cost)
		rsp.Body.Code = constant.OK
		rspObj.Generals = make([]proto.General, len(gs))
		for i, v := range gs {
			rspObj.Generals[i] = v.ToProto().(proto.General)
		}
	}else{
		rsp.Body.Code = constant.DBError
	}
}




func (this*General) ComposeGeneral(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ComposeGeneralReq{}
	rspObj := &proto.ComposeGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)



	gs, ok := mgr.GMgr.HasGeneral(role.RId,reqObj.CompId)
	//是否有这个武将
	if ok == false{
		rsp.Body.Code = constant.GeneralNoHas
		return
	}


	//是否都有这个武将
	gss ,ok := mgr.GMgr.HasGenerals(role.RId,reqObj.GIds)
	if ok == false{
		rsp.Body.Code = constant.GeneralNoHas
		return
	}


	ok = true
	for _, v := range gss {
		t := v
		if t.CfgId != gs.CfgId {
			ok = false
		}
	}

	//是否同一个类型的武将
	if ok == false {
		rsp.Body.Code = constant.GeneralNoSame
		return
	}

	//是否超过武将星级
	if gs.Star - gs.StarLv < len(gss){
		rsp.Body.Code = constant.GeneralStarMax
		return
	}

	gs.StarLv += len(gss)
	gs.HasPrPoint += static_conf.Basic.General.PrPoint * len(gss)
	gs.SyncExecute()


	for _, v := range gss {
		t := v
		t.ParentId = gs.Id
		t.ComposeType = model.ComposeStar
		t.SyncExecute()
	}

	rsp.Body.Code = constant.OK

	rspObj.Generals = make([]proto.General, len(gss))
	for i, v := range gss {
		rspObj.Generals[i] = v.ToProto().(proto.General)
	}
	rspObj.Generals = append(rspObj.Generals,gs.ToProto().(proto.General))

}



func (this*General) AddPrGeneral(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.AddPrGeneralReq{}
	rspObj := &proto.AddPrGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	gs, ok := mgr.GMgr.HasGeneral(role.RId,reqObj.CompId)
	//是否有这个武将
	if ok == false{
		rsp.Body.Code = constant.GeneralNoHas
		return
	}

	all:= reqObj.ForceAdd + reqObj.StrategyAdd +  reqObj.DefenseAdd + reqObj.SpeedAdd + reqObj.DestroyAdd
	if gs.HasPrPoint < all{
		rsp.Body.Code = constant.DBError
		return
	}

	gs.ForceAdded = reqObj.ForceAdd
	gs.StrategyAdded = reqObj.StrategyAdd
	gs.DefenseAdded = reqObj.DefenseAdd
	gs.SpeedAdded = reqObj.SpeedAdd
	gs.DestroyAdded = reqObj.DestroyAdd



	gs.UsePrPoint = all
	gs.SyncExecute()

	rsp.Body.Code = constant.OK

	rspObj.Generals = gs.ToProto().(proto.General)

}

