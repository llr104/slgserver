package controller

import (
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"math/rand"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/net"
	"slgserver/server"
	"slgserver/server/logic"
	"slgserver/server/middleware"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
	"time"
)

var DefaultRole = Role{}

type Role struct {

}

func (this*Role) InitRouter(r *net.Router) {
	g := r.Group("role").Use(middleware.ElapsedTime(), middleware.Log(), middleware.CheckLogin())
	g.AddRouter("create", this.create)
	g.AddRouter("roleList", this.roleList)
	g.AddRouter("enterServer", this.enterServer)
	g.AddRouter("myCity", this.myCity, middleware.CheckRole())
	g.AddRouter("myRoleRes", this.myRoleRes, middleware.CheckRole())
	g.AddRouter("myRoleBuild", this.myRoleBuild, middleware.CheckRole())
	g.AddRouter("myProperty", this.myProperty, middleware.CheckRole())
}

func (this*Role) create(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.CreateRoleReq{}
	rspObj := &proto.CreateRoleRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	uid, _ := req.Conn.GetProperty("uid")
	reqObj.UId = uid.(int)
	rspObj.Role.UId = reqObj.UId

	r := make([]model.Role, 0)
	has, _ := db.MasterDB.Table(r).Where("uid=? and sid=?", reqObj.UId, reqObj.SId).Get(r)
	if has {
		log.DefaultLog.Info("role has create",
			zap.Int("uid", reqObj.UId),
			zap.Int("sid", reqObj.SId))
		rsp.Body.Code = constant.RoleAlreadyCreate
	}else {

		role := &model.Role{UId: reqObj.UId, SId: reqObj.SId,
			HeadId: reqObj.HeadId, Sex: reqObj.Sex,
			NickName: reqObj.NickName, CreatedAt: time.Now()}

		if _, err := db.MasterDB.Insert(role); err != nil {
			log.DefaultLog.Info("role  create error",
				zap.Int("uid", reqObj.UId),
				zap.Int("sid", reqObj.SId),
				zap.Error(err))
			rsp.Body.Code = constant.DBError
		}else{
			rspObj.Role.RId = role.RId
			rspObj.Role.SId = reqObj.SId
			rspObj.Role.UId = reqObj.UId
			rspObj.Role.NickName = reqObj.NickName
			rspObj.Role.Sex = reqObj.Sex
			rspObj.Role.HeadId = reqObj.HeadId

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

	uid, _ := req.Conn.GetProperty("uid")
	uid = uid.(int)

	r := make([]*model.Role, 0)
	err := db.MasterDB.Table(r).Where("uid=?", uid).Find(&r)
	if err == nil{
		rl := make([]proto.Role, len(r))
		for i, d := range r {
			model_to_proto.Role(d, &rl[i])
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

	uid, _ := req.Conn.GetProperty("uid")
	uid = uid.(int)

	role := &model.Role{}
	b, err := db.MasterDB.Table(role).Where("uid=? and sid=?", uid, reqObj.SId).Get(role)
	if err != nil{
		log.DefaultLog.Warn("enterServer db error", zap.Error(err))
		rsp.Body.Code = constant.DBError
		return
	}
	if b {
		rsp.Body.Code = constant.OK
		model_to_proto.Role(role, &rspObj.Role)

		req.Conn.SetProperty("sid", role.SId)
		req.Conn.SetProperty("role", role)

		server.DefaultConnMgr.RoleEnter(req.Conn)

		var e error = nil
		roleRes, ok := logic.RResMgr.Get(role.RId)
		if ok == false{
			roleRes = &model.RoleRes{RId: role.RId, Wood: 10000, Iron: 10000,
				Stone: 10000, Grain: 10000, Gold: 10000,
				Decree: 20, WoodYield: 1000, IronYield: 1000,
				StoneYield: 1000, GrainYield: 1000, GoldYield: 1000,
				DepotCapacity: 100000}

			_ ,e = db.MasterDB.Insert(roleRes)
			if e != nil {
				log.DefaultLog.Error("insert rres error", zap.Error(e))
			}
		}

		if e == nil {
			logic.RResMgr.Add(roleRes)
			model_to_proto.RRes(roleRes, &rspObj.RoleRes)
			rsp.Body.Code = constant.OK
		}else{
			rsp.Body.Code = constant.DBError
			return
		}

		//查询是否有城市
		_, ok = logic.RCMgr.GetByRId(role.RId)
		if ok == false{
			citys := make([]*model.MapRoleCity, 0)
			//随机生成一个城市
			for true {
				x := rand.Intn(logic.MapWith)
				y := rand.Intn(logic.MapHeight)
				if logic.NMMgr.IsCanBuild(x, y) && logic.RBMgr.IsEmpty(x, y) && logic.RCMgr.IsEmpty(x, y){
					//建立城市
					c := &model.MapRoleCity{RId: role.RId, X: x, Y: y, IsMain: 1,
						CurDurable: 100, MaxDurable: 100, Level: 1, Name: role.NickName, CreatedAt: time.Now()}

					//插入
					_, err := db.MasterDB.Table(c).Insert(c)
					if err != nil{
						rsp.Body.Code = constant.DBError
					}else{
						citys = append(citys, c)
						//更新城市缓存
						logic.RCMgr.Add(c)
					}

					//生成城市里面的设施
					logic.RFMgr.GetAndTryCreate(c.CityId)
					break
				}
			}
		}

	}else{
		rsp.Body.Code = constant.RoleNotExist
	}
}

func (this*Role) myCity(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyCityReq{}
	rspObj := &proto.MyCityRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role, _ := r.(*model.Role)


	citys,ok := logic.RCMgr.GetByRId(role.RId)
	if ok {
		rspObj.Citys = make([]proto.MapRoleCity, len(citys))
		//赋值发送
		for i, v := range citys {
			model_to_proto.MCBuild(v, &rspObj.Citys[i])
		}

	}else{
		rspObj.Citys = make([]proto.MapRoleCity, 0)
	}

}

func (this*Role) myRoleRes(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyRoleResReq{}
	rspObj := &proto.MyRoleResRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	roleRes, ok := logic.RResMgr.Get(role.RId)
	if ok == false{
		rsp.Body.Code = constant.RoleNotExist
		return
	}else{
		model_to_proto.RRes(roleRes, &rspObj.RoleRes)
		rsp.Body.Code = constant.OK
	}
}

func (this*Role) myProperty(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyRolePropertyReq{}
	rspObj := &proto.MyRolePropertyRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	//城市
	c, ok := logic.RCMgr.GetByRId(role.RId)
	if ok {
		rspObj.Citys = make([]proto.MapRoleCity, len(c))
		for i, v := range c {
			model_to_proto.MCBuild(v, &rspObj.Citys[i])
		}
	}else{
		rspObj.Citys = make([]proto.MapRoleCity, 0)
	}


	//建筑
	ra, ok := logic.RBMgr.GetRoleBuild(role.RId)
	if ok {
		rspObj.MRBuilds = make([]proto.MapRoleBuild, len(ra))
		for i, v := range ra {
			model_to_proto.MRBuild(v, &rspObj.MRBuilds[i])
		}
	}else{
		rspObj.MRBuilds = make([]proto.MapRoleBuild, 0)
	}

	//资源
	roleRes, ok := logic.RResMgr.Get(role.RId)
	if ok {
		model_to_proto.RRes(roleRes, &rspObj.RoleRes)
	}else{
		rsp.Body.Code = constant.RoleNotExist
		return
	}

	//武将
	gs, ok := logic.GMgr.GetAndTryCreate(role.RId)
	if ok {
		rspObj.Generals = make([]proto.General, len(gs))
		for i, v := range gs {
			model_to_proto.General(v, &rspObj.Generals[i])
		}
	}else{
		rsp.Body.Code = constant.DBError
		return
	}

	//军队
	ar, ok := logic.AMgr.GetByRId(role.RId)
	if ok {
		rspObj.Armys = make([]proto.Army, len(ar))
		for i, v := range ar {
			model_to_proto.Army(v, &rspObj.Armys[i])
		}
	}else{
		rspObj.Armys = make([]proto.Army, 0)
	}

}

func (this*Role) myRoleBuild(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyRoleBuildReq{}
	rspObj := &proto.MyRoleBuildRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	ra, ok := logic.RBMgr.GetRoleBuild(role.RId)
	if ok {
		rspObj.MRBuilds = make([]proto.MapRoleBuild, len(ra))
		for i, v := range ra {
			model_to_proto.MRBuild(v, &rspObj.MRBuilds[i])
		}
	}else{
		rspObj.MRBuilds = make([]proto.MapRoleBuild, 0)
	}

}