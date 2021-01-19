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
	"slgserver/server/slgserver/static_conf/facility"
	"time"
)

var DefaultInterior = Interior{}

type Interior struct {

}

func (this*Interior) InitRouter(r *net.Router) {
	g := r.Group("interior").Use(middleware.ElapsedTime(),
		middleware.Log(), middleware.CheckRole())
	g.AddRouter("collect", this.collect)
	g.AddRouter("openCollect", this.openCollect)
	g.AddRouter("transform", this.transform)

}

func (this*Interior) collect(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.CollectionReq{}
	rspObj := &proto.CollectionRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	roleRes, ok:= mgr.RResMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.DBError
		return
	}

	roleAttr, ok:= mgr.RAttrMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.DBError
		return
	}

	curTime := time.Now()
	lastTime := roleAttr.LastCollectTime
	if curTime.YearDay() != lastTime.YearDay() || curTime.Year() != lastTime.Year(){
		roleAttr.CollectTimes = 0
		roleAttr.LastCollectTime = time.Time{}
	}

	timeLimit := static_conf.Basic.Role.CollectTimesLimit
	//是否超过征收次数上限
	if roleAttr.CollectTimes >= timeLimit{
		rsp.Body.Code = constant.OutCollectTimesLimit
		return
	}

	//cd内不能操作
	need := lastTime.Add(time.Duration(static_conf.Basic.Role.CollectTimesLimit)*time.Second)
	if curTime.Before(need){
		rsp.Body.Code = constant.InCdCanNotOperate
		return
	}

	gold := mgr.GetYield(roleRes.RId).Gold
	rspObj.Gold = gold
	roleRes.Gold += gold
	roleRes.SyncExecute()

	roleAttr.LastCollectTime = curTime
	roleAttr.CollectTimes += 1
	roleAttr.SyncExecute()

	interval := static_conf.Basic.Role.CollectInterval
	if roleAttr.CollectTimes >= timeLimit {
		y, m, d := roleAttr.LastCollectTime.Add(24*time.Hour).Date()
		nextTime := time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("IST", 3600))
		rspObj.NextTime = nextTime.UnixNano()/1e6
	}else{
		nextTime := roleAttr.LastCollectTime.Add(time.Duration(interval)*time.Second)
		rspObj.NextTime = nextTime.UnixNano()/1e6
	}

	rspObj.CurTimes = roleAttr.CollectTimes
	rspObj.Limit = timeLimit

}

func (this*Interior) openCollect(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.OpenCollectionReq{}
	rspObj := &proto.OpenCollectionRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	roleAttr, ok:= mgr.RAttrMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.DBError
		return
	}

	interval := static_conf.Basic.Role.CollectInterval
	timeLimit := static_conf.Basic.Role.CollectTimesLimit
	rspObj.Limit = timeLimit
	rspObj.CurTimes = roleAttr.CollectTimes
	if roleAttr.LastCollectTime.IsZero() {
		rspObj.NextTime = 0
	}else{
		if roleAttr.CollectTimes >= timeLimit {
			y, m, d := roleAttr.LastCollectTime.Add(24*time.Hour).Date()
			nextTime := time.Date(y, m, d, 0, 0, 0, 0, time.FixedZone("IST", 3600))
			rspObj.NextTime = nextTime.UnixNano()/1e6
		}else{
			nextTime := roleAttr.LastCollectTime.Add(time.Duration(interval)*time.Second)
			rspObj.NextTime = nextTime.UnixNano()/1e6
		}
	}
}


func (this*Interior) transform(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.TransformReq{}
	rspObj := &proto.TransformRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	roleRes, ok:= mgr.RResMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.DBError
		return
	}

	main, _ := mgr.RCMgr.GetMainCity(role.RId)

	lv := mgr.RFMgr.GetFacilityLv(main.CityId, facility.JiShi)
	if lv <= 0{
		rsp.Body.Code = constant.NotHasJiShi
		return
	}

	len := 4
	ret := make([]int, len)

	for i := 0 ;i < len; i++{
		//ret[i] = reqObj.To[i] - reqObj.From[i]
		if reqObj.From[i] > 0{
			ret[i] = -reqObj.From[i]
		}

		if reqObj.To[i] > 0{
			ret[i] = reqObj.To[i]
		}
	}


	if roleRes.Wood + ret[0] < 0{
		rsp.Body.Code = constant.InvalidParam
		return
	}

	if roleRes.Iron + ret[1] < 0{
		rsp.Body.Code = constant.InvalidParam
		return
	}

	if roleRes.Stone + ret[2] < 0{
		rsp.Body.Code = constant.InvalidParam
		return
	}

	if roleRes.Grain + ret[3] < 0{
		rsp.Body.Code = constant.InvalidParam
		return
	}

	roleRes.Wood += ret[0]
	roleRes.Iron += ret[1]
	roleRes.Stone += ret[2]
	roleRes.Grain += ret[3]
	roleRes.SyncExecute()

}

