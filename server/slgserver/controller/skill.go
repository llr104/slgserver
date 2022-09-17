package controller

import (
	"github.com/goinggo/mapstructure"
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/middleware"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/proto"
)

var DefaultSkill = Skill{}

type Skill struct {
}

func (this *Skill) InitRouter(r *net.Router) {
	g := r.Group("skill").Use(middleware.ElapsedTime(), middleware.Log(),
		middleware.CheckLogin(), middleware.CheckRole())

	g.AddRouter("list", this.list)

}

func (this *Skill) list(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.SkillListReq{}
	rspObj := &proto.SkillListRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK
	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	rspObj.List = make([]proto.Skill, 0)
	skills, _ := mgr.SkillMgr.Get(role.RId)
	for _, skill := range skills {
		rspObj.List = append(rspObj.List, skill.ToProto().(proto.Skill))
	}
}
