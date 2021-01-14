package controller

import (
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/middleware"
	"slgserver/net"
	"slgserver/server/slgserver/model"
	"slgserver/server/slgserver/proto"
)

var DefaultWar = War{}

type War struct {

}

func (this*War) InitRouter(r *net.Router) {
	g := r.Group("war").Use(middleware.ElapsedTime(),
		middleware.Log(), middleware.CheckRole())
	g.AddRouter("report", this.report)
	g.AddRouter("read", this.read)
}

func (this*War) report(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.WarReportReq{}
	rspObj := &proto.WarReportRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	//查询最近30条战报吧
	l := make([]*model.WarReport, 0)
	err := db.MasterDB.Table(model.WarReport{}).Where("a_rid=? or d_rid=?",
		role.RId, role.RId).Desc("ctime").Limit(100, 0).Find(&l)

	if err != nil{
		log.DefaultLog.Warn("db error", zap.Error(err))
		rsp.Body.Code = constant.DBError
		return
	}

	rspObj.List = make([]proto.WarReport, len(l))
	for i, v := range l {
		rspObj.List[i] = v.ToProto().(proto.WarReport)
	}
}

func (this*War) read(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.WarReadReq{}
	rspObj := &proto.WarReadRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	rspObj.Id = reqObj.Id

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	if reqObj.Id == 0 {
		m := model.WarReport{AttackIsRead: true, DefenseIsRead: true}
		_, err := db.MasterDB.Table(model.WarReport{}).Where("a_rid=?", role.RId).Cols("a_is_read").Update(m)
		if err != nil {
			log.DefaultLog.Error("db error", zap.Error(err))
		}
		_, err = db.MasterDB.Table(model.WarReport{}).Where("d_rid=?", role.RId).Cols("d_is_read").Update(m)
		if err != nil {
			log.DefaultLog.Error("db error", zap.Error(err))
		}
	}else{
		wr := &model.WarReport{}
		ok, err := db.MasterDB.Table(model.WarReport{}).Where("id=?",
			reqObj.Id).Get(wr)

		if err != nil {
			log.DefaultLog.Warn("db error", zap.Error(err))
			rsp.Body.Code = constant.DBError
			return
		}

		if ok {
			if wr.AttackRid == role.RId {
				wr.AttackIsRead = true
				db.MasterDB.Table(wr).ID(wr.Id).Cols("a_is_read").Update(wr)
				rsp.Body.Code = constant.OK
			}else if wr.DefenseRid == role.RId {
				wr.DefenseIsRead = true
				db.MasterDB.Table(wr).ID(wr.Id).Cols("d_is_read").Update(wr)
				rsp.Body.Code = constant.OK
			}else{
				rsp.Body.Code = constant.InvalidParam
			}
		}
	}

}
