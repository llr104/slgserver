package controller

import (
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/middleware"
	"slgserver/net"
	"slgserver/server/slgserver/logic"
	"slgserver/server/slgserver/logic/mgr"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/proto"
	"slgserver/server/slgserver/static_conf"
	"time"
)

var DefaultCoalition= coalition{

}

type coalition struct {

}

func (this *coalition) InitRouter(r *net.Router) {
	g := r.Group("union").Use(middleware.ElapsedTime(), middleware.Log(),
		middleware.CheckLogin(), middleware.CheckRole())

	g.AddRouter("create", this.create)
	g.AddRouter("list", this.list)
	g.AddRouter("join", this.join)
	g.AddRouter("verify", this.verify)
	g.AddRouter("member", this.member)
	g.AddRouter("applyList", this.applyList)
	g.AddRouter("exit", this.exit)
	g.AddRouter("dismiss", this.dismiss)
	g.AddRouter("notice", this.notice)
	g.AddRouter("modNotice", this.modNotice)
	g.AddRouter("kick", this.kick)
	g.AddRouter("appoint", this.appoint)
	g.AddRouter("abdicate", this.abdicate)
	g.AddRouter("info", this.info)

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

	has := mgr.RAttrMgr.IsHasUnion(role.RId)
	if has {
		rsp.Body.Code = constant.UnionAlreadyHas
		return
	}

	c, ok := mgr.UnionMgr.Create(reqObj.Name, role.RId)
	if ok {
		rspObj.Id = c.Id
		logic.Union.MemberEnter(role.RId, c.Id)
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

	l := mgr.UnionMgr.List()
	rspObj.List = make([]proto.Union, len(l))
	for i, u := range l {
		rspObj.List[i] = u.ToProto().(proto.Union)
		main := make([]proto.Major, 0)
		if r, ok := mgr.RMgr.Get(u.Chairman); ok {
			m := proto.Major{Name: r.NickName, RId: r.RId, Title: proto.UnionChairman}
			main = append(main, m)
		}

		if r, ok := mgr.RMgr.Get(u.ViceChairman); ok {
			m := proto.Major{Name: r.NickName, RId: r.RId, Title: proto.UnionViceChairman}
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

	has := mgr.RAttrMgr.IsHasUnion(role.RId)
	if has {
		rsp.Body.Code = constant.UnionAlreadyHas
		return
	}

	u, ok := mgr.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if len(u.MemberArray) >= static_conf.Basic.Union.MemberLimit{
		rsp.Body.Code = constant.PeopleIsFull
		return
	}

	//判断当前是否已经有申请
	has, _ = db.MasterDB.Table(model.CoalitionApply{}).Where(
		"union_id=? and state=? and rid=?",
		reqObj.Id, proto.UnionUntreated, role.RId).Get(&model.CoalitionApply{})
	if has {
		rsp.Body.Code = constant.HasApply
		return
	}

	//写入申请列表
	apply := &model.CoalitionApply{
		RId:     role.RId,
		UnionId: reqObj.Id,
		Ctime:   time.Now(),
		State:   proto.UnionUntreated}

	_, err := db.MasterDB.InsertOne(apply)
	if err != nil{
		rsp.Body.Code = constant.DBError
		log.DefaultLog.Warn("db error", zap.Error(err))
		return
	}

	//推送主、副盟主
	apply.SyncExecute()
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
		"id=? and state=?", reqObj.Id, proto.UnionUntreated).Get(apply)
	if ok && err == nil{
		if u, ok := mgr.UnionMgr.Get(apply.UnionId); ok {

			if u.Chairman != role.RId && u.ViceChairman != role.RId {
				rsp.Body.Code = constant.PermissionDenied
				return
			}

			if len(u.MemberArray) >= static_conf.Basic.Union.MemberLimit{
				rsp.Body.Code = constant.PeopleIsFull
				return
			}

			if ok := mgr.RAttrMgr.IsHasUnion(apply.RId); ok {
				rsp.Body.Code = constant.UnionAlreadyHas
			}else{
				if reqObj.Decide == proto.UnionAdopt {
					//同意
					c, ok := mgr.UnionMgr.Get(apply.UnionId)
					if ok {
						c.MemberArray = append(c.MemberArray, apply.RId)
						logic.Union.MemberEnter(apply.RId, apply.UnionId)
						c.SyncExecute()
					}
				}
			}
			apply.State = reqObj.Decide
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

	union, ok := mgr.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	rspObj.Members = make([]proto.Member, 0)
	for _, rid := range union.MemberArray {
		if role, ok := mgr.RMgr.Get(rid); ok {
			m := proto.Member{RId: role.RId, Name: role.NickName }
			if main, ok := mgr.RCMgr.GetMainCity(role.RId); ok {
				m.X = main.X
				m.Y = main.Y
			}

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

	u, ok := mgr.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		rspObj.Applys = make([]proto.ApplyItem, 0)
		return
	}

	applys := make([]*model.CoalitionApply, 0)
	err := db.MasterDB.Table(model.CoalitionApply{}).Where(
		"union_id=? and state=?", reqObj.Id, 0).Find(&applys)
	if err != nil{
		rsp.Body.Code = constant.DBError
		return
	}

	rspObj.Id = reqObj.Id
	rspObj.Applys = make([]proto.ApplyItem, 0)
	for _, apply := range applys {
		if r, ok := mgr.RMgr.Get(apply.RId);ok{
			a := proto.ApplyItem{Id: apply.Id, RId: apply.RId, NickName: r.NickName}
			rspObj.Applys = append(rspObj.Applys, a)
		}
	}
}

//退出
func (this *coalition) exit(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ExitReq{}
	rspObj := &proto.ExitRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if ok := mgr.RAttrMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	attribute, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(attribute.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	//盟主不能退出
	if u.Chairman == role.RId {
		rsp.Body.Code = constant.UnionNotAllowExit
		return
	}

	for i, rid := range u.MemberArray {
		if rid == role.RId{
			u.MemberArray = append(u.MemberArray[:i], u.MemberArray[i+1:]...)
		}
	}

	if u.ViceChairman == role.RId{
		u.ViceChairman = 0
	}

	logic.Union.MemberExit(role.RId)
	u.SyncExecute()

}


//解散
func (this *coalition) dismiss(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.DismissReq{}
	rspObj := &proto.DismissRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if ok := mgr.RAttrMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	attribute, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(attribute.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	//盟主才能解散
	if u.Chairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}
	unionId := attribute.UnionId
	logic.Union.Dismiss(unionId)
}

//公告
func (this *coalition) notice(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.NoticeReq{}
	rspObj := &proto.NoticeRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	u, ok := mgr.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	rspObj.Text = u.Notice
}

//修改公告
func (this *coalition) modNotice(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ModNoticeReq{}
	rspObj := &proto.ModNoticeRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK
	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if len(reqObj.Text) > 200 {
		rsp.Body.Code = constant.ContentTooLong
		return
	}

	if ok := mgr.RAttrMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	attribute, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(attribute.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	rspObj.Text = reqObj.Text
	rspObj.Id = u.Id
	u.Notice = reqObj.Text
	u.SyncExecute()
}

//踢人
func (this *coalition) kick(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.KickReq{}
	rspObj := &proto.KickRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.RId = reqObj.RId

	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if ok := mgr.RAttrMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	opAr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(opAr.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	if role.RId == reqObj.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	target, ok := mgr.RAttrMgr.Get(reqObj.RId)
	if ok {
		if target.UnionId == u.Id{
			for i, rid := range u.MemberArray {
				if rid == reqObj.RId{
					u.MemberArray = append(u.MemberArray[:i], u.MemberArray[i+1:]...)
				}
			}
			if u.ViceChairman == reqObj.RId{
				u.ViceChairman = 0
			}
			logic.Union.MemberExit(reqObj.RId)
			target.UnionId = 0
			u.SyncExecute()
		}else{
			rsp.Body.Code = constant.NotBelongUnion
		}
	}else{
		rsp.Body.Code = constant.NotBelongUnion
	}
}

//任命
func (this *coalition) appoint(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.AppointReq{}
	rspObj := &proto.AppointRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.RId = reqObj.RId

	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if ok := mgr.RAttrMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	opAr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(opAr.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	target, ok := mgr.RAttrMgr.Get(reqObj.RId)
	if ok {
		if target.UnionId == u.Id{
			if reqObj.Title == proto.UnionViceChairman {
				u.ViceChairman = reqObj.RId
				rspObj.Title = reqObj.Title
				u.SyncExecute()
			}else if reqObj.Title == proto.UnionCommon {
				if u.ViceChairman == reqObj.RId{
					u.ViceChairman = 0
				}
				rspObj.Title = reqObj.Title
			}else{
				rsp.Body.Code = constant.InvalidParam
			}
		}else{
			rsp.Body.Code = constant.NotBelongUnion
		}
	}else{
		rsp.Body.Code = constant.NotBelongUnion
	}

}

//禅让
func (this *coalition) abdicate(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.AbdicateReq{}
	rspObj := &proto.AbdicateRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if ok := mgr.RAttrMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	opAr, _ := mgr.RAttrMgr.Get(role.RId)
	u, ok := mgr.UnionMgr.Get(opAr.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	target, ok := mgr.RAttrMgr.Get(reqObj.RId)
	if ok {
		if target.UnionId == u.Id{
			if role.RId == u.Chairman{
				u.Chairman = reqObj.RId
				if u.ViceChairman == reqObj.RId{
					u.ViceChairman = 0
				}
			}else if role.RId == u.ViceChairman {
				u.ViceChairman = reqObj.RId
			}
		}else{
			rsp.Body.Code = constant.NotBelongUnion
		}
	}else{
		rsp.Body.Code = constant.NotBelongUnion
	}

}

//联盟信息
func (this *coalition) info(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.InfoReq{}
	rspObj := &proto.InfoRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.Id = reqObj.Id

	rsp.Body.Code = constant.OK

	u, ok := mgr.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
	}else{
		rspObj.Info = u.ToProto().(proto.Union)
		main := make([]proto.Major, 0)
		if r, ok := mgr.RMgr.Get(u.Chairman); ok {
			m := proto.Major{Name: r.NickName, RId: r.RId, Title: proto.UnionChairman}
			main = append(main, m)
		}

		if r, ok := mgr.RMgr.Get(u.ViceChairman); ok {
			m := proto.Major{Name: r.NickName, RId: r.RId, Title: proto.UnionViceChairman}
			main = append(main, m)
		}
		rspObj.Info.Major = main
	}
}
