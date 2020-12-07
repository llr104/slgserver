package controller

import (
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/net"
	"slgserver/server/logic"
	"slgserver/server/middleware"
	"slgserver/server/model"
	"slgserver/server/proto"
	"time"
)

var DefaultCoalition= coalition{

}

type coalition struct {

}

func (this *coalition) InitRouter(r *net.Router) {
	g := r.Group("city").Use(middleware.ElapsedTime(), middleware.Log(),
		middleware.CheckLogin(), middleware.CheckRole())

	g.AddRouter("create", this.create)
	g.AddRouter("list", this.list)
	g.AddRouter("join", this.join)
	g.AddRouter("verify", this.verify)
	g.AddRouter("member", this.member)
	g.AddRouter("applyList", this.applyList)

}

//创建联盟
func (this *coalition) create(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.CreateReq{}
	rspObj := &proto.CreateRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	rspObj.Name = reqObj.Name

	c, ok := logic.UnionMgr.Create(reqObj.Name, role.RId)
	if ok {
		rspObj.Id = c.Id
	}else{
		rsp.Body.Code = constant.UnionCreateError
	}
}

//联盟列表
func (this *coalition) list(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ListReq{}
	rspObj := &proto.ListRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK

	l := logic.UnionMgr.List()
	rspObj.List = make([]proto.Union, len(l))
	for i, v := range l {
		rspObj.List[i].Name = v.Name
		rspObj.List[i].Notice = v.Notice
		rspObj.List[i].Cnt = v.Cnt()

		main := make([]proto.Member, 0)
		if r, ok := logic.RMgr.Get(v.Chairman); ok {
			m := proto.Member{Name: r.NickName, RId: r.RId, Title: proto.UnionChairman}
			main = append(main, m)
		}

		if r, ok := logic.RMgr.Get(v.ViceChairman); ok {
			m := proto.Member{Name: r.NickName, RId: r.RId, Title: proto.UnionViceChairman}
			main = append(main, m)
		}
		rspObj.List[i].Major = main
	}
}

//加入
func (this *coalition) join(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.JoinReq{}
	rspObj := &proto.JoinRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	//需要判断是否已经有联盟了，后续加

	_, ok := logic.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	//写入申请列表
	c := &model.CoalitionApply{
		RId: role.RId,
		CoalitionId: reqObj.Id,
		Ctime: time.Now(),
		State: proto.UnionUntreated}

	_, err := db.MasterDB.InsertOne(c)
	if err != nil{
		rsp.Body.Code = constant.DBError
		log.DefaultLog.Warn("db error", zap.Error(err))
		return
	}
}

//审核
func (this *coalition) verify(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

	reqObj := &proto.VerifyReq{}
	rspObj := &proto.VerifyRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.Id = reqObj.Id
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	apply := &model.CoalitionApply{}
	ok, err := db.MasterDB.Table(model.CoalitionApply{}).Where(
		"coalition_id=? and state=?", reqObj.Id, proto.UnionUntreated).Get(&apply)
	if ok && err == nil{
		if u, ok := logic.UnionMgr.Get(apply.Id); ok {

			if u.Chairman != role.RId && u.ViceChairman != u.ViceChairman {
				rsp.Body.Code = constant.PermissionDenied
				return
			}

			if reqObj.Decide == proto.UnionAdopt{
				//同意
			}else{
				//拒绝
			}
			db.MasterDB.Table(apply).ID(apply.Id).Cols("state").Update(apply)
		}else{
			rsp.Body.Code = constant.UnionNotFound
			return
		}

	}else{
		rsp.Body.Code = constant.InvalidParam
	}

}

//成员列表
func (this *coalition) member(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MemberReq{}
	rspObj := &proto.MemberRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.Id = reqObj.Id
	rsp.Body.Code = constant.OK

	union, ok := logic.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	rspObj.Members = make([]proto.Member, 0)
	for _, rid := range union.MemberArray {
		if role, ok := logic.RMgr.Get(rid); ok {
			m := proto.Member{RId: role.RId, Name: role.NickName }
			if rid == union.Chairman {
				m.Title = proto.UnionChairman
			}else if rid == union.ViceChairman {
				m.Title = proto.UnionViceChairman
			}else {
				m.Title = proto.UnionCommon
			}
			rspObj.Members = append(rspObj.Members, m)
		}
	}
}


//申请列表
func (this *coalition) applyList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ApplyReq{}
	rspObj := &proto.ApplyRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	u, ok := logic.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId && u.ViceChairman != u.ViceChairman {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	applys := make([]*model.CoalitionApply, 0)
	err := db.MasterDB.Table(model.CoalitionApply{}).Where(
		"coalition_id=? and state=?", reqObj.Id, 0).Find(&applys)
	if err != nil{
		rsp.Body.Code = constant.DBError
		return
	}

	rspObj.Applys = make([]proto.ApplyItem, 0)
	for _, apply := range applys {
		if r, ok := logic.RMgr.Get(apply.RId);ok{
			a := proto.ApplyItem{Id: apply.Id, RId: apply.RId, NickName: r.NickName}
			rspObj.Applys = append(rspObj.Applys, a)
		}
	}
}