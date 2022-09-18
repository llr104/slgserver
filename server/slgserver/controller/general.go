package controller

import (
	"github.com/goinggo/mapstructure"
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/middleware"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/server/slgserver/static_conf/skill"
)

var DefaultGeneral = General{}

type General struct {
}

func (this *General) InitRouter(r *net.Router) {
	g := r.Group("general").Use(middleware.ElapsedTime(), middleware.Log(),
		middleware.CheckLogin(), middleware.CheckRole())

	g.AddRouter("myGenerals", this.myGenerals)
	g.AddRouter("drawGeneral", this.drawGenerals)
	g.AddRouter("composeGeneral", this.composeGeneral)
	g.AddRouter("addPrGeneral", this.addPrGeneral)
	g.AddRouter("convert", this.convert)
	g.AddRouter("upSkill", this.upSkill)
	g.AddRouter("downSkill", this.downSkill)
	g.AddRouter("lvSkill", this.lvSkill)

}

func (this *General) myGenerals(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyGeneralReq{}
	rspObj := &proto.MyGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	gs, ok := mgr.GMgr.GetOrCreateByRId(role.RId)
	if ok {
		rsp.Body.Code = constant.OK
		rspObj.Generals = make([]proto.General, 0)
		for _, v := range gs {
			rspObj.Generals = append(rspObj.Generals, v.ToProto().(proto.General))
		}
	} else {
		rsp.Body.Code = constant.DBError
	}
}

func (this *General) drawGenerals(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.DrawGeneralReq{}
	rspObj := &proto.DrawGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	cost := static_conf.Basic.General.DrawGeneralCost * reqObj.DrawTimes
	ok := mgr.RResMgr.GoldIsEnough(role.RId, cost)
	if ok == false {
		rsp.Body.Code = constant.GoldNotEnough
		return
	}

	limit := static_conf.Basic.General.Limit
	cnt := mgr.GMgr.Count(role.RId)
	if cnt+reqObj.DrawTimes > limit {
		rsp.Body.Code = constant.OutGeneralLimit
		return
	}

	gs, ok := mgr.GMgr.RandCreateGeneral(role.RId, reqObj.DrawTimes)

	if ok {
		mgr.RResMgr.TryUseGold(role.RId, cost)
		rsp.Body.Code = constant.OK
		rspObj.Generals = make([]proto.General, len(gs))
		for i, v := range gs {
			rspObj.Generals[i] = v.ToProto().(proto.General)
		}
	} else {
		rsp.Body.Code = constant.DBError
	}
}

func (this *General) composeGeneral(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ComposeGeneralReq{}
	rspObj := &proto.ComposeGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	gs, ok := mgr.GMgr.HasGeneral(role.RId, reqObj.CompId)
	//是否有这个武将
	if ok == false {
		rsp.Body.Code = constant.GeneralNoHas
		return
	}

	//是否都有这个武将
	gss, ok := mgr.GMgr.HasGenerals(role.RId, reqObj.GIds)
	if ok == false {
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
	if int(gs.Star-gs.StarLv) < len(gss) {
		rsp.Body.Code = constant.GeneralStarMax
		return
	}

	gs.StarLv += int8(len(gss))
	gs.HasPrPoint += static_conf.Basic.General.PrPoint * len(gss)
	gs.SyncExecute()

	for _, v := range gss {
		t := v
		t.ParentId = gs.Id
		t.State = model.GeneralComposeStar
		t.SyncExecute()
	}

	rspObj.Generals = make([]proto.General, len(gss))
	for i, v := range gss {
		rspObj.Generals[i] = v.ToProto().(proto.General)
	}
	rspObj.Generals = append(rspObj.Generals, gs.ToProto().(proto.General))

}

func (this *General) addPrGeneral(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.AddPrGeneralReq{}
	rspObj := &proto.AddPrGeneralRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	gs, ok := mgr.GMgr.HasGeneral(role.RId, reqObj.CompId)
	//是否有这个武将
	if ok == false {
		rsp.Body.Code = constant.GeneralNoHas
		return
	}

	all := reqObj.ForceAdd + reqObj.StrategyAdd + reqObj.DefenseAdd + reqObj.SpeedAdd + reqObj.DestroyAdd
	if gs.HasPrPoint < all {
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

func (this *General) convert(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ConvertReq{}
	rspObj := &proto.ConvertRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	roleRes, ok := mgr.RResMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.DBError
		return
	}

	gold := 0
	okArray := make([]int, 0)
	for _, gid := range reqObj.GIds {
		g, ok := mgr.GMgr.GetByGId(gid)
		if ok && g.Order == 0 {
			okArray = append(okArray, gid)
			gold += 10 * int(g.Star) * (1 + int(g.StarLv))
			g.State = model.GeneralConvert
			g.SyncExecute()
		}
	}

	roleRes.Gold += gold
	rspObj.AddGold = gold
	rspObj.Gold = roleRes.Gold
	rspObj.GIds = okArray

	roleRes.SyncExecute()
}

func (this *General) upSkill(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpDownSkillReq{}
	rspObj := &proto.UpDownSkillRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	rspObj.Pos = reqObj.Pos
	rspObj.CfgId = reqObj.CfgId
	rspObj.GId = reqObj.GId

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if reqObj.Pos < 0 || reqObj.Pos >= model.SkillLimit {
		rsp.Body.Code = constant.InvalidParam
		return
	}

	g, ok := mgr.GMgr.GetByGId(reqObj.GId)
	if ok == false {
		rsp.Body.Code = constant.GeneralNotFound
		return
	}

	if g.RId != role.RId {
		rsp.Body.Code = constant.GeneralNotMe
		return
	}

	skill, ok := mgr.SkillMgr.GetSkillOrCreate(role.RId, reqObj.CfgId)
	if ok == false {
		rsp.Body.Code = constant.DBError
		return
	}

	if skill.IsInLimit() == false {
		rsp.Body.Code = constant.OutSkillLimit
		return
	}

	if skill.ArmyIsIn(g.CurArms) == false {
		rsp.Body.Code = constant.OutArmNotMatch
		return
	}

	if g.UpSkill(skill.Id, reqObj.CfgId, reqObj.Pos) == false {
		rsp.Body.Code = constant.UpSkillError
		return
	}
	skill.UpSkill(g.Id)
	g.SyncExecute()
	skill.SyncExecute()
}

func (this *General) downSkill(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpDownSkillReq{}
	rspObj := &proto.UpDownSkillRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	rspObj.Pos = reqObj.Pos
	rspObj.CfgId = reqObj.CfgId
	rspObj.GId = reqObj.GId

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if reqObj.Pos < 0 || reqObj.Pos >= model.SkillLimit {
		rsp.Body.Code = constant.InvalidParam
		return
	}

	g, ok := mgr.GMgr.GetByGId(reqObj.GId)
	if ok == false {
		rsp.Body.Code = constant.GeneralNotFound
		return
	}

	if g.RId != role.RId {
		rsp.Body.Code = constant.GeneralNotMe
		return
	}

	skill, ok := mgr.SkillMgr.GetSkillOrCreate(role.RId, reqObj.CfgId)
	if ok == false {
		rsp.Body.Code = constant.DBError
		return
	}

	if g.DownSkill(skill.Id, reqObj.Pos) == false {
		rsp.Body.Code = constant.DownSkillError
		return
	}
	skill.DownSkill(g.Id)
	g.SyncExecute()
	skill.SyncExecute()
}

func (this *General) lvSkill(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.LvSkillReq{}
	rspObj := &proto.LvSkillRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	rspObj.Pos = reqObj.Pos
	rspObj.GId = reqObj.GId

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	g, ok := mgr.GMgr.GetByGId(reqObj.GId)
	if ok == false {
		rsp.Body.Code = constant.GeneralNotFound
		return
	}

	if g.RId != role.RId {
		rsp.Body.Code = constant.GeneralNotMe
		return
	}

	gSkill, err := g.PosSkill(reqObj.Pos)
	if err != nil {
		rsp.Body.Code = constant.PosNotSkill
		return
	}

	skillCfg, ok := skill.Skill.GetCfg(gSkill.CfgId)
	if ok == false {
		rsp.Body.Code = constant.PosNotSkill
		return
	}

	if gSkill.Lv > len(skillCfg.Levels) {
		rsp.Body.Code = constant.SkillLevelFull
		return
	}

	gSkill.Lv += 1
	g.SyncExecute()
}
