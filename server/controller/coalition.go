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
	"slgserver/server/static_conf"
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

	has := logic.RAttributeMgr.IsHasUnion(role.RId)
	if has {
		rsp.Body.Code = constant.UnionAlreadyHas
		return
	}

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
	for i, u := range l {
		rspObj.List[i] = u.ToProto().(proto.Union)
		main := make([]proto.Member, 0)
		if r, ok := logic.RMgr.Get(u.Chairman); ok {
			m := proto.Member{Name: r.NickName, RId: r.RId, Title: proto.UnionChairman}
			main = append(main, m)
		}

		if r, ok := logic.RMgr.Get(u.ViceChairman); ok {
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

	has := logic.RAttributeMgr.IsHasUnion(role.RId)
	if has {
		rsp.Body.Code = constant.UnionAlreadyHas
		return
	}

	u, ok := logic.UnionMgr.Get(reqObj.Id)
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
	c := &model.CoalitionApply{
		RId: role.RId,
		UnionId: reqObj.Id,
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
		"id=? and state=?", reqObj.Id, proto.UnionUntreated).Get(apply)
	if ok && err == nil{
		if u, ok := logic.UnionMgr.Get(apply.UnionId); ok {

			if u.Chairman != role.RId && u.ViceChairman != u.ViceChairman {
				rsp.Body.Code = constant.PermissionDenied
				return
			}

			if len(u.MemberArray) >= static_conf.Basic.Union.MemberLimit{
				rsp.Body.Code = constant.PeopleIsFull
				return
			}

			if ok := logic.RAttributeMgr.IsHasUnion(apply.RId); ok {
				rsp.Body.Code = constant.UnionAlreadyHas
			}else{
				if reqObj.Decide == proto.UnionAdopt{
					//同意
					c, ok := logic.UnionMgr.Get(apply.UnionId)
					if ok {
						c.MemberArray = append(c.MemberArray, apply.RId)
						logic.RAttributeMgr.EnterUnion(apply.RId, apply.UnionId)

						if citys, ok := logic.RCMgr.GetByRId(apply.RId); ok {
							for _, city := range citys {
								city.UnionId = apply.UnionId
								city.SyncExecute()
							}
						}

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
		"union_id=? and state=?", reqObj.Id, 0).Find(&applys)
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

//退出
func (this *coalition) exit(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ExitReq{}
	rspObj := &proto.ExitRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if ok := logic.RAttributeMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	attribute, _ := logic.RAttributeMgr.Get(role.RId)
	u, ok := logic.UnionMgr.Get(attribute.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	//盟主、副盟主不能退出
	if u.Chairman == role.RId || u.ViceChairman == u.ViceChairman {
		rsp.Body.Code = constant.UnionNotAllowExit
		return
	}

	for i, rid := range u.MemberArray {
		if rid == role.RId{
			u.MemberArray = append(u.MemberArray[:i], u.MemberArray[i+1:]...)
		}
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

	if ok := logic.RAttributeMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	attribute, _ := logic.RAttributeMgr.Get(role.RId)
	u, ok := logic.UnionMgr.Get(attribute.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	//盟主才能解散
	if u.Chairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	logic.UnionMgr.Remove(attribute.UnionId)

	for _, rid := range u.MemberArray {
		logic.Union.MemberExit(rid)
	}

	u.State = model.UnionDismiss
	u.MemberArray = []int{}
	attribute.UnionId = 0
	u.SyncExecute()


}

//公告
func (this *coalition) notice(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.NoticeReq{}
	rspObj := &proto.NoticeRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	u, ok := logic.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	rspObj.Text = u.Notice
}

//修改公告
func (this *coalition) modNotice(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ModNoticeReq{}
	rspObj := &proto.ModNoticeReq{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	rsp.Body.Code = constant.OK
	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if len(reqObj.Text) > 200 {
		rsp.Body.Code = constant.ContentTooLong
		return
	}

	if ok := logic.RAttributeMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	attribute, _ := logic.RAttributeMgr.Get(role.RId)
	u, ok := logic.UnionMgr.Get(attribute.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	rspObj.Text = reqObj.Text
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

	if ok := logic.RAttributeMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	opAr, _ := logic.RAttributeMgr.Get(role.RId)
	u, ok := logic.UnionMgr.Get(opAr.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	target, ok := logic.RAttributeMgr.Get(reqObj.RId)
	if ok {
		if target.UnionId == u.Id{
			for i, rid := range u.MemberArray {
				if rid == reqObj.RId{
					u.MemberArray = append(u.MemberArray[:i], u.MemberArray[i+1:]...)
				}
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

	if ok := logic.RAttributeMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	opAr, _ := logic.RAttributeMgr.Get(role.RId)
	u, ok := logic.UnionMgr.Get(opAr.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	target, ok := logic.RAttributeMgr.Get(reqObj.RId)
	if ok {
		if target.UnionId == u.Id{
			if reqObj.Title == proto.UnionViceChairman{
				u.ViceChairman = reqObj.RId
				rspObj.Title = reqObj.Title
				u.SyncExecute()
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

	if ok := logic.RAttributeMgr.IsHasUnion(role.RId); ok == false {
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	opAr, _ := logic.RAttributeMgr.Get(role.RId)
	u, ok := logic.UnionMgr.Get(opAr.UnionId)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
		return
	}

	if u.Chairman != role.RId && u.ViceChairman != role.RId {
		rsp.Body.Code = constant.PermissionDenied
		return
	}

	target, ok := logic.RAttributeMgr.Get(reqObj.RId)
	if ok {
		if target.UnionId == u.Id{
			if role.RId == u.Chairman{
				u.Chairman = reqObj.RId
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

	u, ok := logic.UnionMgr.Get(reqObj.Id)
	if ok == false{
		rsp.Body.Code = constant.UnionNotFound
	}else{
		rspObj.Info = u.ToProto().(proto.Union)
		main := make([]proto.Member, 0)
		if r, ok := logic.RMgr.Get(u.Chairman); ok {
			m := proto.Member{Name: r.NickName, RId: r.RId, Title: proto.UnionChairman}
			main = append(main, m)
		}

		if r, ok := logic.RMgr.Get(u.ViceChairman); ok {
			m := proto.Member{Name: r.NickName, RId: r.RId, Title: proto.UnionViceChairman}
			main = append(main, m)
		}
		rspObj.Info.Major = main
	}
}
