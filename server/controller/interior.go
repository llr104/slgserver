package controller

import (
	"github.com/goinggo/mapstructure"
	"slgserver/constant"
	"slgserver/net"
	"slgserver/server/logic"
	"slgserver/server/middleware"
	"slgserver/server/model"
	"slgserver/server/proto"
)

var DefaultInterior = Interior{}

type Interior struct {

}

func (this*Interior) InitRouter(r *net.Router) {
	g := r.Group("interior").Use(middleware.ElapsedTime(),
		middleware.Log(), middleware.CheckRole())
	g.AddRouter("collection", this.collection)
}

func (this*Interior) collection(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.CollectionReq{}
	rspObj := &proto.CollectionRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	roleRes, ok:= logic.RResMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.DBError
		return

	}
	roleRes.Gold += roleRes.GoldYield
	roleRes.SyncExecute()

}


