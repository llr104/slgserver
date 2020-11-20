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
	"slgserver/server/entity"
	"slgserver/server/middleware"
	"slgserver/server/model_to_proto"
	"slgserver/server/proto"
	"time"
)

var DefaultRole = Role{}

type Role struct {

}

func (this*Role) InitRouter(r *net.Router) {
	g := r.Group("role").Use(middleware.Log(), middleware.CheckLogin())
	g.AddRouter("create", this.create)
	g.AddRouter("roleList", this.roleList)
	g.AddRouter("enterServer", this.enterServer)
	g.AddRouter("myCity", this.myCity, middleware.CheckRole())
	g.AddRouter("myRoleRes", this.myRoleRes, middleware.CheckRole())
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
		var roleRes *model.RoleRes
		if roleRes, err = entity.RResMgr.Get(role.RId); err != nil{
			roleRes := &model.RoleRes{RId: role.RId, Wood: 10000, Iron: 10000,
				Stone: 10000, Grain: 10000, Gold: 10000,
				Decree: 20, WoodYield: 1000, IronYield: 1000,
				StoneYield: 1000, GrainYield: 1000, GoldYield: 1000,
				DepotCapacity: 100000}

			_ ,e = db.MasterDB.Insert(roleRes)
			if e != nil {
				log.DefaultLog.Error("insert roleRes error", zap.Error(e))
			}
		}

		if e == nil {
			entity.RResMgr.Add(roleRes)
			model_to_proto.RRes(roleRes, &rspObj.RoleRes)
			rsp.Body.Code = constant.OK
		}else{
			rsp.Body.Code = constant.DBError
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
	citys := make([]*model.MapRoleCity, 0)
	//查询是否有城市
	db.MasterDB.Table(new(model.MapRoleCity)).Where("rid=?", role.RId).Find(&citys)
	if len(citys) == 0 {
		//随机生成一个城市
		for true {
			x := rand.Intn(entity.MapWith)
			y := rand.Intn(entity.MapHeight)
			if entity.NMMgr.IsCanBuild(x, y) && entity.RBMgr.IsEmpty(x, y) && entity.RCMgr.IsEmpty(x, y){
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
					entity.RCMgr.Add(c)
				}

				//生成城市里面的设施
				entity.RFMgr.GetAndTryCreate(c.CityId)
				break
			}
		}
	}

	//赋值发送
	rspObj.Citys = make([]proto.MapRoleCity, len(citys))
	for i, v := range citys {
		model_to_proto.MCBuild(v, &rspObj.Citys[i])
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

	roleRes, err := entity.RResMgr.Get(role.RId)
	if err != nil{
		log.DefaultLog.Warn("myRoleRes db error", zap.Error(err))
		rsp.Body.Code = constant.RoleNotExist
		return
	}else{
		model_to_proto.RRes(roleRes, &rspObj.RoleRes)
		rsp.Body.Code = constant.OK
	}
}