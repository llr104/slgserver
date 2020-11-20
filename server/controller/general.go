package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/model"
	"slgserver/net"
	"slgserver/server/logic"
	"slgserver/server/middleware"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
	"slgserver/server/static_conf"
	"slgserver/server/static_conf/facility"
	"slgserver/server/static_conf/general"
	"time"
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
	g.AddRouter("conscript", this.conscript)
	g.AddRouter("assignArmy", this.assignArmy)

}

func (this*General) myGenerals(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyGeneralReq{}
	rspObj := &proto.MyGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	gs, ok := logic.GMgr.GetAndTryCreate(role.RId)
	if ok {
		rsp.Body.Code = constant.OK
		rspObj.Generals = make([]proto.General, len(gs))
		for i, v := range gs {
			model_to_proto.General(v, &rspObj.Generals[i])
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
	city,ok := logic.RCMgr.Get(reqObj.CityId)
	if ok == false{
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	as, _ := logic.AMgr.GetByCity(reqObj.CityId)
	rspObj.Armys = make([]proto.Army, len(as))
	for i, v := range as {
		model_to_proto.Army(v, &rspObj.Armys[i])
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

	city, ok := logic.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	g, ok := logic.GMgr.FindGeneral(reqObj.GeneralId)
	if ok == false{
		rsp.Body.Code = constant.GeneralNotFound
		return
	}

	if g.RId != role.RId{
		rsp.Body.Code = constant.GeneralNotMe
		return
	}

	army, err := logic.AMgr.GetOrCreate(role.RId, reqObj.CityId, reqObj.Order)
	if err != nil{
		rsp.Body.Code = constant.DBError
		return
	}

	if army.State != model.ArmyIdle{
		rsp.Body.Code = constant.ArmyBusy
		return
	}

	//配置逻辑
	if reqObj.Position == 0{
		if army.FirstId == g.Id{
			army.FirstId = 0
			army.FirstSoldierCnt = 0
		}else if army.SecondId == g.Id{
			army.SecondId = 0
			army.SecondSoldierCnt = 0
		}else if army.ThirdId == g.Id{
			army.ThirdId = 0
			army.ThirdSoldierCnt = 0
		}
		g.Order = 0
		g.CityId = 0
		g.NeedUpdate = true
	}else{
		if reqObj.Position == 1 {
			//旧的下阵
			if army.FirstId != 0 {
				if oldG, ok := logic.GMgr.FindGeneral(army.FirstId); ok{
					oldG.CityId = 0
					oldG.Order = 0
					oldG.NeedUpdate = true
				}
			}

			army.FirstSoldierCnt = 0
			army.FirstId = g.Id
		}else if reqObj.Position == 2 {
			//旧的下阵
			if army.SecondId != 0 {
				if oldG, ok := logic.GMgr.FindGeneral(army.SecondId); ok{
					oldG.CityId = 0
					oldG.Order = 0
					oldG.NeedUpdate = true
				}
			}

			army.SecondSoldierCnt = 0
			army.SecondId = g.Id
		}else if reqObj.Position == 3 {
			//旧的下阵
			if army.ThirdId != 0 {
				if oldG, ok := logic.GMgr.FindGeneral(army.ThirdId); ok{
					oldG.CityId = 0
					oldG.Order = 0
					oldG.NeedUpdate = true
				}
			}

			army.ThirdSoldierCnt = 0
			army.ThirdId = g.Id
		}
		//新的上阵
		g.Order = reqObj.Position
		g.CityId = reqObj.CityId
		g.NeedUpdate = true
	}

	if c, ok := logic.RCMgr.Get(army.CityId); ok{
		army.FromX = c.X
		army.FromY = c.Y
	}

	army.NeedUpdate = true
	//队伍
	model_to_proto.Army(army, &rspObj.Army)
}

//征兵
func (this*General) conscript(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ConscriptReq{}
	rspObj := &proto.ConscriptRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	if reqObj.ArmyId <= 0 || reqObj.FirstCnt < 0 || reqObj.SecondCnt < 0 || reqObj.ThirdCnt < 0{
		rsp.Body.Code = constant.InvalidParam
		return
	}

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	army,ok := logic.AMgr.Get(reqObj.ArmyId)
	if ok == false{
		rsp.Body.Code = constant.ArmyNotFound
		return
	}

	if role.RId != army.RId{
		rsp.Body.Code = constant.ArmyNotMe
		return
	}

	addCnts := []int{reqObj.FirstCnt, reqObj.SecondCnt, reqObj.ThirdCnt}
	armyIds := []int{army.FirstId, army.SecondId, army.ThirdId}
	armyCnts := []int{army.FirstSoldierCnt, army.SecondSoldierCnt, army.ThirdSoldierCnt}

	//判断是否超过上限
	for i, gid := range armyIds {
		if gid == 0 {
			if i == 0{
				reqObj.FirstCnt = 0
				addCnts[0] = 0
			}else if i==1 {
				reqObj.SecondCnt = 0
				addCnts[1] = 0
			}else if i==2 {
				reqObj.ThirdCnt = 0
				addCnts[2] = 0
			}
			continue
		}
		if g, ok := logic.GMgr.FindGeneral(gid); ok {
			l, e := general.GenBasic.GetLevel(g.Level)
			if e == nil{
				if l.Soldiers < addCnts[i]+armyCnts[i]{
					rsp.Body.Code = constant.OutArmyLimit
					return
				}
			}else{
				rsp.Body.Code = constant.InvalidParam
				return
			}
		}
	}

	//开始征兵
	total := reqObj.FirstCnt + reqObj.SecondCnt + reqObj.ThirdCnt
	conscript := static_conf.Basic.ConScript
	needWood := total*conscript.CostWood
	needGrain := total*conscript.CostGrain
	needIron := total*conscript.CostIron
	needStone := total*conscript.CostStone
	needGold := total*conscript.CostGold

	nr := facility.NeedRes{Grain: needGrain, Wood: needWood,
		Gold: needGold, Iron: needIron, Decree: 0,
		Stone: needStone}

	if ok := logic.RResMgr.TryUseNeed(army.RId, &nr); ok {
		army.FirstSoldierCnt += reqObj.FirstCnt
		army.SecondSoldierCnt += reqObj.SecondCnt
		army.ThirdSoldierCnt += reqObj.ThirdCnt
		army.NeedUpdate = true

		//队伍
		model_to_proto.Army(army, &rspObj.Army)

		//资源
		if rRes, ok := logic.RResMgr.Get(role.RId); ok {
			model_to_proto.RRes(rRes, &rspObj.RoleRes)
		}

		rsp.Body.Code = constant.OK
	}else{
		rsp.Body.Code = constant.ResNotEnough
	}
}

//派遣队伍
func (this*General) assignArmy(req *net.WsMsgReq, rsp *net.WsMsgRsp){
	reqObj := &proto.AssignArmyReq{}
	rspObj := &proto.AssignArmyRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if reqObj.State != 1 && reqObj.State != 2{
		rsp.Body.Code = constant.InvalidParam
		return
	}

	if reqObj.X < 0 || reqObj.X >= logic.MapWith ||
		reqObj.Y < 0 || reqObj.Y >= logic.MapHeight{
		rsp.Body.Code = constant.InvalidParam
		return
	}

	army,ok := logic.AMgr.Get(reqObj.ArmyId)
	if ok == false{
		rsp.Body.Code = constant.ArmyNotFound
		return
	}

	if role.RId != army.RId{
		rsp.Body.Code = constant.ArmyNotMe
		return
	}

	if army.State != 0 {
		rsp.Body.Code = constant.ArmyBusy
		return
	}

	army.Start = time.Now()
	//先写死1分钟到达
	army.End = time.Now().Add(5*time.Second)
	army.ToX = reqObj.X
	army.ToY = reqObj.Y
	army.State = reqObj.State

	army.NeedUpdate = true
	model_to_proto.Army(army, &rspObj.Army)
	logic.AMgr.PushAction(army)
}

