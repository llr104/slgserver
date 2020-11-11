package controller

import (
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/net"
	"slgserver/server/middleware"
	"slgserver/server/proto"
	"time"
)

var DefaultRole = Role{}

type Role struct {

}

func (this*Role) InitRouter(r *net.Router) {
	g := r.Group("role").Use(middleware.Log())
	g.AddRouter("create", this.create)
	g.AddRouter("roleList", this.roleList)
	g.AddRouter("enterServer", this.enterServer)
}

func (this*Role) create(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.CreateRoleReq{}
	rspObj := &proto.CreateRoleRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	uid, err := req.Conn.GetProperty("uid")
	if err != nil {
		log.DefaultLog.Warn("create but connect not found uid")
		rsp.Body.Code = constant.InvalidParam
		return
	}
	reqObj.UId = uid.(int)
	rspObj.UId = reqObj.UId

	r := make([]model.Role, 0)
	has, _ := db.MasterDB.Table(r).Where("uid=? and sid=?", reqObj.UId, reqObj.SId).Get(r)
	if has {
		log.DefaultLog.Info("role has create",
			zap.Int("uid", reqObj.UId),
			zap.Int("sid", reqObj.SId))
		rsp.Body.Code = constant.RoleAlreadyCreate
	}else {
		rr := &model.Role{UId: reqObj.UId, SId: reqObj.SId,
			HeadId: reqObj.HeadId, Sex: reqObj.Sex,
			NickName: reqObj.NickName, CreatedAt: time.Now()}

		rId, err := db.MasterDB.Insert(rr)
		if err != nil {
			log.DefaultLog.Info("role  create error",
				zap.Int("uid", reqObj.UId),
				zap.Int("sid", reqObj.SId),
				zap.Error(err))

			rsp.Body.Code = constant.DBError
		}else{
			rspObj.RId = int(rId)
			rspObj.SId = reqObj.SId
			rspObj.UId = reqObj.UId
			rspObj.NickName = reqObj.NickName
			rspObj.Sex = reqObj.Sex
			rspObj.HeadId = reqObj.HeadId
			rsp.Body.Code = constant.OK
		}
	}
}

func (this*Role) roleList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.RoleListReq{}
	rspObj := &proto.RoleListRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	uid, err := req.Conn.GetProperty("uid")
	if err != nil {
		log.DefaultLog.Warn("roleList but connect not found uid")
		rsp.Body.Code = constant.InvalidParam
		return
	}


	r := make([]model.Role, 0)
	err = db.MasterDB.Table(r).Where("uid=?", uid).Find(r)
	if err == nil{
		rl := make([]proto.Role, len(r))
		for i, d := range r {
			rl[i].UId = d.UId
			rl[i].SId = d.SId
			rl[i].RId = d.RId
			rl[i].Sex = d.Sex
			rl[i].NickName = d.NickName
			rl[i].HeadId = d.HeadId
			rl[i].Balance = d.Balance
			rl[i].Profile = d.Profile
		}
		rspObj.Roles = rl
		rsp.Body.Code = constant.OK
	}else{
		rsp.Body.Code = constant.DBError
	}
}

func (this*Role) enterServer(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.EnterServerReq{}
	rspObj := &proto.EnterServerRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	uid, err := req.Conn.GetProperty("uid")
	if err != nil {
		log.DefaultLog.Warn("enterServer but connect not found uid")
		rsp.Body.Code = constant.InvalidParam
		return
	}

	uid = uid.(int)
	r := &model.Role{}
	b, err := db.MasterDB.Table(r).Where("uid=? and sid=?", uid, reqObj.SId).Get(r)
	if b && err == nil {
		rsp.Body.Code = constant.OK
		rspObj.Role.UId = r.UId
		rspObj.Role.SId = r.SId
		rspObj.Role.RId = r.RId
		rspObj.Role.Sex = r.Sex
		rspObj.Role.NickName = r.NickName
		rspObj.Role.HeadId = r.HeadId
		rspObj.Role.Balance = r.Balance
		rspObj.Role.Profile = r.Profile

		req.Conn.SetProperty("sid", r.SId)
	}else{
		rsp.Body.Code = constant.RoleNotExist
	}

}