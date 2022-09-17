package controller

import (
	"time"

	"github.com/goinggo/mapstructure"
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/middleware"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/logic"
	"github.com/llr104/slgserver/server/slgserver/logic/check"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/server/slgserver/static_conf/facility"
	"github.com/llr104/slgserver/server/slgserver/static_conf/general"
)

var DefaultArmy = Army{}

type Army struct {
}

func (this *Army) InitRouter(r *net.Router) {
	g := r.Group("army").Use(middleware.ElapsedTime(), middleware.Log(),
		middleware.CheckLogin(), middleware.CheckRole())

	g.AddRouter("myList", this.myList)
	g.AddRouter("myOne", this.myOne)
	g.AddRouter("dispose", this.dispose)
	g.AddRouter("conscript", this.conscript)
	g.AddRouter("assign", this.assign)

}

//我的军队列表
func (this *Army) myList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ArmyListReq{}
	rspObj := &proto.ArmyListRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	rspObj.CityId = reqObj.CityId

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	city, ok := mgr.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	as, _ := mgr.AMgr.GetByCity(reqObj.CityId)
	rspObj.Armys = make([]proto.Army, len(as))
	for i, v := range as {
		rspObj.Armys[i] = v.ToProto().(proto.Army)
	}
}

//我的某一个军队
func (this *Army) myOne(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ArmyOneReq{}
	rspObj := &proto.ArmyOneRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	city, ok := mgr.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	a, ok := mgr.AMgr.GetByCityOrder(reqObj.CityId, reqObj.Order)
	if ok {
		rspObj.Army = a.ToProto().(proto.Army)
	} else {
		rsp.Body.Code = constant.ArmyNotFound
	}

}

//配置武将
func (this *Army) dispose(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

	reqObj := &proto.DisposeReq{}
	rspObj := &proto.DisposeRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if reqObj.Order <= 0 || reqObj.Order > 5 || reqObj.Position < -1 || reqObj.Position > 2 {
		rsp.Body.Code = constant.InvalidParam
		return
	}

	city, ok := mgr.RCMgr.Get(reqObj.CityId)
	if ok == false {
		rsp.Body.Code = constant.CityNotExist
		return
	}

	if city.RId != role.RId {
		rsp.Body.Code = constant.CityNotMe
		return
	}

	//校场每升一级一个队伍
	jc, ok := mgr.RFMgr.GetFacility(city.CityId, facility.JiaoChang)
	if ok == false || jc.GetLevel() < reqObj.Order {
		rsp.Body.Code = constant.ArmyNotEnough
		return
	}

	newG, ok := mgr.GMgr.GetByGId(reqObj.GeneralId)
	if ok == false {
		rsp.Body.Code = constant.GeneralNotFound
		return
	}

	if newG.RId != role.RId {
		rsp.Body.Code = constant.GeneralNotMe
		return
	}

	army, err := mgr.AMgr.GetOrCreate(role.RId, reqObj.CityId, reqObj.Order)
	if err != nil {
		rsp.Body.Code = constant.DBError
		return
	}

	if army.FromX != city.X || army.FromY != city.Y {
		rsp.Body.Code = constant.ArmyIsOutside
		return
	}

	//下阵
	if reqObj.Position == -1 {
		for pos, g := range army.Gens {
			if g != nil && g.Id == newG.Id {

				//征兵中不能下阵
				if army.PositionCanModify(pos) == false {
					if army.Cmd == model.ArmyCmdConscript {
						rsp.Body.Code = constant.GeneralBusy
					} else {
						rsp.Body.Code = constant.ArmyBusy
					}
					return
				}

				army.GeneralArray[pos] = 0
				army.SoldierArray[pos] = 0
				army.Gens[pos] = nil
				army.SyncExecute()
				break
			}
		}
		newG.Order = 0
		newG.CityId = 0
		newG.SyncExecute()
	} else {

		//征兵中不能下阵
		if army.PositionCanModify(reqObj.Position) == false {
			if army.Cmd == model.ArmyCmdConscript {
				rsp.Body.Code = constant.GeneralBusy
			} else {
				rsp.Body.Code = constant.ArmyBusy
			}
			return
		}

		if newG.CityId != 0 {
			rsp.Body.Code = constant.GeneralBusy
			return
		}

		if mgr.AMgr.IsRepeat(role.RId, newG.CfgId) == false {
			rsp.Body.Code = constant.GeneralRepeat
			return
		}

		//判断是否能配前锋
		tst, ok := mgr.RFMgr.GetFacility(city.CityId, facility.TongShuaiTing)
		if reqObj.Position == 2 && (ok == false || tst.GetLevel() < reqObj.Order) {
			rsp.Body.Code = constant.TongShuaiNotEnough
			return
		}

		//判断cost
		cost := general.General.Cost(newG.CfgId)
		for i, g := range army.Gens {
			if g == nil || i == reqObj.Position {
				continue
			}
			cost += general.General.Cost(g.CfgId)
		}

		if mgr.GetCityCost(city.CityId) < cost {
			rsp.Body.Code = constant.CostNotEnough
			return
		}

		oldG := army.Gens[reqObj.Position]
		if oldG != nil {
			//旧的下阵
			oldG.CityId = 0
			oldG.Order = 0
			oldG.SyncExecute()
		}

		//新的上阵
		army.GeneralArray[reqObj.Position] = reqObj.GeneralId
		army.Gens[reqObj.Position] = newG
		army.SoldierArray[reqObj.Position] = 0

		newG.Order = reqObj.Order
		newG.CityId = reqObj.CityId
		newG.SyncExecute()
	}

	army.FromX = city.X
	army.FromY = city.Y
	army.SyncExecute()
	//队伍
	rspObj.Army = army.ToProto().(proto.Army)
}

//征兵
func (this *Army) conscript(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ConscriptReq{}
	rspObj := &proto.ConscriptRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	if reqObj.ArmyId <= 0 || len(reqObj.Cnts) != 3 ||
		reqObj.Cnts[0] < 0 || reqObj.Cnts[1] < 0 || reqObj.Cnts[2] < 0 {
		rsp.Body.Code = constant.InvalidParam
		return
	}

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	army, ok := mgr.AMgr.Get(reqObj.ArmyId)
	if ok == false {
		rsp.Body.Code = constant.ArmyNotFound
		return
	}

	if role.RId != army.RId {
		rsp.Body.Code = constant.ArmyNotMe
		return
	}

	//判断该位置是否能征兵
	for pos, cnt := range reqObj.Cnts {
		if cnt > 0 {
			if army.Gens[pos] == nil {
				rsp.Body.Code = constant.InvalidParam
				return
			}
			if army.PositionCanModify(pos) == false {
				rsp.Body.Code = constant.GeneralBusy
				return
			}
		}
	}

	lv := mgr.RFMgr.GetFacilityLv(army.CityId, facility.MBS)
	if lv <= 0 {
		rsp.Body.Code = constant.BuildMBSNotFound
		return
	}

	//判断是否超过上限
	for i, g := range army.Gens {
		if g == nil {
			continue
		}

		l, e := general.GenBasic.GetLevel(g.Level)
		add := mgr.RFMgr.GetAdditions(army.CityId, facility.TypeSoldierLimit)
		if e == nil {
			if l.Soldiers+add[0] < reqObj.Cnts[i]+army.SoldierArray[i] {
				rsp.Body.Code = constant.OutArmyLimit
				return
			}
		} else {
			rsp.Body.Code = constant.InvalidParam
			return
		}
	}

	//开始征兵
	total := 0
	for _, n := range reqObj.Cnts {
		total += n
	}

	conscript := static_conf.Basic.ConScript
	needWood := total * conscript.CostWood
	needGrain := total * conscript.CostGrain
	needIron := total * conscript.CostIron
	needStone := total * conscript.CostStone
	needGold := total * conscript.CostGold

	nr := facility.NeedRes{Grain: needGrain, Wood: needWood,
		Gold: needGold, Iron: needIron, Decree: 0,
		Stone: needStone}

	if code := mgr.RResMgr.TryUseNeed(army.RId, nr); code == constant.OK {
		curTime := time.Now().Unix()
		for i, _ := range army.SoldierArray {
			if reqObj.Cnts[i] > 0 {
				army.ConscriptCntArray[i] = reqObj.Cnts[i]
				army.ConscriptTimeArray[i] = int64(reqObj.Cnts[i]*conscript.CostTime) + curTime - 2
			}
		}

		army.Cmd = model.ArmyCmdConscript
		army.SyncExecute()

		//队伍
		rspObj.Army = army.ToProto().(proto.Army)

		//资源
		if rRes, ok := mgr.RResMgr.Get(role.RId); ok {
			rspObj.RoleRes = rRes.ToProto().(proto.RoleRes)
		}
		rsp.Body.Code = constant.OK
	} else {
		rsp.Body.Code = constant.ResNotEnough
	}
}

//派遣队伍
func (this *Army) assign(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.AssignArmyReq{}
	rspObj := &proto.AssignArmyRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	army, ok := mgr.AMgr.Get(reqObj.ArmyId)
	if ok == false {
		rsp.Body.Code = constant.ArmyNotFound
		return
	}

	if role.RId != army.RId {
		rsp.Body.Code = constant.ArmyNotMe
		return
	}

	if reqObj.Cmd == model.ArmyCmdBack {
		rsp.Body.Code = this.__back__(army)
	} else if reqObj.Cmd == model.ArmyCmdAttack {
		rsp.Body.Code = this.__attack__(reqObj, army, role)
	} else if reqObj.Cmd == model.ArmyCmdDefend {
		rsp.Body.Code = this.__defend__(reqObj, army, role)
	} else if reqObj.Cmd == model.ArmyCmdReclamation {
		rsp.Body.Code = this.__reclamation__(reqObj, army, role)
	} else if reqObj.Cmd == model.ArmyCmdTransfer {
		rsp.Body.Code = this.__transfer__(reqObj, army, role)
	}

	rspObj.Army = army.ToProto().(proto.Army)
}

func (this *Army) __pre__(reqObj *proto.AssignArmyReq, army *model.Army, role *model.Role) int {

	if reqObj.X < 0 || reqObj.X >= global.MapWith ||
		reqObj.Y < 0 || reqObj.Y >= global.MapHeight {
		return constant.InvalidParam
	}

	if army.IsCanOutWar() == false {
		if army.Cmd != model.ArmyCmdIdle {
			return constant.ArmyBusy
		} else {
			return constant.ArmyNotMain
		}
	}

	if army.IsIdle() == false {
		return constant.ArmyBusy
	}

	//判断该地是否是能攻击类型
	cfg, ok := mgr.NMMgr.PositionBuild(reqObj.X, reqObj.Y)
	if ok == false || cfg.Type == 0 {
		return constant.InvalidParam
	}

	if check.IsCanArrive(reqObj.X, reqObj.Y, role.RId) == false {
		return constant.UnReachable
	}

	return constant.OK
}

func (this *Army) __after__(reqObj *proto.AssignArmyReq, army *model.Army) int {
	//最后才消耗体力
	cost := static_conf.Basic.General.CostPhysicalPower
	ok := mgr.GMgr.PhysicalPowerIsEnough(army, cost)
	if ok == false {
		return constant.PhysicalPowerNotEnough
	}

	if reqObj.Cmd == model.ArmyCmdReclamation || reqObj.Cmd == model.ArmyCmdTransfer {
		cost := static_conf.Basic.General.ReclamationCost
		if mgr.RResMgr.DecreeIsEnough(army.RId, cost) == false {
			return constant.DecreeNotEnough
		} else {
			mgr.RResMgr.TryUseDecree(army.RId, cost)
		}
	}

	mgr.GMgr.TryUsePhysicalPower(army, cost)

	army.ToX = reqObj.X
	army.ToY = reqObj.Y
	army.Cmd = reqObj.Cmd
	army.State = model.ArmyRunning

	if global.IsDev() {
		army.Start = time.Now()
		army.End = time.Now().Add(40 * time.Second)
	} else {
		speed := mgr.AMgr.GetSpeed(army)
		t := mgr.TravelTime(speed, army.FromX, army.FromY, army.ToX, army.ToY)
		army.Start = time.Now()
		army.End = time.Now().Add(time.Duration(t) * time.Millisecond)
	}

	logic.ArmyLogic.PushAction(army)

	return constant.OK
}

//返回
func (this *Army) __back__(army *model.Army) int {
	if army.Cmd == model.ArmyCmdAttack ||
		army.Cmd == model.ArmyCmdDefend ||
		army.Cmd == model.ArmyCmdReclamation {
		logic.ArmyLogic.ArmyBack(army)
	} else if army.IsIdle() {
		city, ok := mgr.RCMgr.Get(army.CityId)
		if ok {
			if city.X != army.FromX || city.Y != army.FromY {
				logic.ArmyLogic.ArmyBack(army)
			}
		}

	}
	return constant.OK
}

//攻击
func (this *Army) __attack__(reqObj *proto.AssignArmyReq, army *model.Army, role *model.Role) int {

	code := this.__pre__(reqObj, army, role)
	if code != constant.OK {
		return code
	}

	if check.IsCanArrive(reqObj.X, reqObj.Y, role.RId) == false {
		return constant.UnReachable
	}

	//免战
	if check.IsWarFree(reqObj.X, reqObj.Y) {
		return constant.BuildWarFree
	}

	if check.IsCanDefend(reqObj.X, reqObj.Y, role.RId) == true {
		return constant.BuildCanNotAttack
	}

	return this.__after__(reqObj, army)
}

//驻守
func (this *Army) __defend__(reqObj *proto.AssignArmyReq, army *model.Army, role *model.Role) int {

	code := this.__pre__(reqObj, army, role)
	if code != constant.OK {
		return code
	}

	if check.IsCanDefend(reqObj.X, reqObj.Y, role.RId) == false {
		return constant.BuildCanNotDefend
	}

	return this.__after__(reqObj, army)
}

func (this *Army) __reclamation__(reqObj *proto.AssignArmyReq, army *model.Army, role *model.Role) int {

	code := this.__pre__(reqObj, army, role)
	if code != constant.OK {
		return code
	}

	if mgr.RBMgr.BuildIsRId(reqObj.X, reqObj.Y, role.RId) == false {
		return constant.BuildNotMe
	}

	return this.__after__(reqObj, army)
}

func (this *Army) __transfer__(reqObj *proto.AssignArmyReq, army *model.Army, role *model.Role) int {

	code := this.__pre__(reqObj, army, role)
	if code != constant.OK {
		return code
	}

	if army.FromX == reqObj.X && army.FromY == reqObj.Y {
		return constant.CanNotTransfer
	}

	if mgr.RBMgr.BuildIsRId(reqObj.X, reqObj.Y, role.RId) == false {
		return constant.BuildNotMe
	}

	b, ok := mgr.RBMgr.PositionBuild(reqObj.X, reqObj.Y)
	if ok == false {
		return constant.BuildNotMe
	}

	if b.Level <= 0 || b.IsHasTransferAuth() == false {
		return constant.CanNotTransfer
	}

	cnt := 0
	if b.IsRoleFortress() {
		cnt = static_conf.MapBCConf.GetHoldArmyCnt(b.Type, b.Level)
	} else {
		cnt = 5
	}

	if cnt > mgr.AMgr.BelongPosArmyCnt(b.RId, b.X, b.Y) {
		return this.__after__(reqObj, army)
	} else {
		return constant.HoldIsFull
	}

}
