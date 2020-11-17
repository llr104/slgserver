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
	"slgserver/server/entity"
	"slgserver/server/middleware"
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

	r := make([]model.Role, 0)
	err := db.MasterDB.Table(r).Where("uid=?", uid).Find(&r)
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
		rspObj.Role.UId = role.UId
		rspObj.Role.SId = role.SId
		rspObj.Role.RId = role.RId
		rspObj.Role.Sex = role.Sex
		rspObj.Role.NickName = role.NickName
		rspObj.Role.HeadId = role.HeadId
		rspObj.Role.Balance = role.Balance
		rspObj.Role.Profile = role.Profile

		req.Conn.SetProperty("sid", role.SId)
		req.Conn.SetProperty("role", role)

		roleRes := &model.RoleRes{RId: role.RId, Wood: 10000, Iron: 10000,
			Stone: 10000, Grain: 10000, Gold: 10000,
			Decree: 20, WoodYield: 1000, IronYield: 1000,
			StoneYield: 1000, GrainYield: 1000, GoldYield: 1000,
			DepotCapacity: 100000}

		if _, err := db.MasterDB.Insert(roleRes);err != nil {
			log.DefaultLog.Info("role_res create error",
				zap.Int("rid", role.RId),
				zap.Error(err))
			rsp.Body.Code = constant.DBError
		}else{
			rspObj.RoleRes.Gold = roleRes.Gold
			rspObj.RoleRes.Grain = roleRes.Grain
			rspObj.RoleRes.Stone = roleRes.Stone
			rspObj.RoleRes.Iron = roleRes.Iron
			rspObj.RoleRes.Wood = roleRes.Wood
			rspObj.RoleRes.Decree = roleRes.Decree
			rspObj.RoleRes.GoldYield = roleRes.GoldYield
			rspObj.RoleRes.GrainYield = roleRes.GrainYield
			rspObj.RoleRes.StoneYield = roleRes.StoneYield
			rspObj.RoleRes.IronYield = roleRes.IronYield
			rspObj.RoleRes.WoodYield = roleRes.WoodYield
			rspObj.RoleRes.DepotCapacity = roleRes.DepotCapacity
			rsp.Body.Code = constant.OK
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
	citys := make([]model.MapRoleCity, 0)
	//查询是否有城市
	db.MasterDB.Table(new(model.MapRoleCity)).Where("rid=?", role.RId).Find(&citys)
	if len(citys) == 0 {
		//随机生成一个城市
		for true {
			x := rand.Intn(entity.MapWith)
			y := rand.Intn(entity.MapHeight)
			if entity.RBMgr.IsEmpty(x, y) && entity.RCMgr.IsEmpty(x, y){
				//建立城市
				c := &model.MapRoleCity{RId: role.RId, X: x, Y: y, IsMain: 1,
					CurDurable: 100, MaxDurable: 100, Level: 1, Name: role.NickName, CreatedAt: time.Now()}

				//插入
				_, err := db.MasterDB.Table(c).Insert(c)
				if err != nil{
					rsp.Body.Code = constant.DBError
				}else{
					citys = append(citys, *c)
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
		rspObj.Citys[i].CityId = v.CityId
		rspObj.Citys[i].RId = v.RId
		rspObj.Citys[i].X = v.X
		rspObj.Citys[i].Y = v.Y
		rspObj.Citys[i].Level = v.Level
		rspObj.Citys[i].IsMain = v.IsMain!=0
		rspObj.Citys[i].Name = v.Name
		rspObj.Citys[i].CurDurable = v.CurDurable
		rspObj.Citys[i].MaxDurable = v.MaxDurable
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
	roleRes := &model.RoleRes{}
	b, err := db.MasterDB.Table(roleRes).Where("rid=?", role.RId).Get(roleRes)

	if err != nil{
		log.DefaultLog.Warn("myRoleRes db error", zap.Error(err))
		return
	}

	if b {
		rspObj.RoleRes.Gold = roleRes.Gold
		rspObj.RoleRes.Grain = roleRes.Grain
		rspObj.RoleRes.Stone = roleRes.Stone
		rspObj.RoleRes.Iron = roleRes.Iron
		rspObj.RoleRes.Wood = roleRes.Wood
		rspObj.RoleRes.Decree = roleRes.Decree
		rspObj.RoleRes.GoldYield = roleRes.GoldYield
		rspObj.RoleRes.GrainYield = roleRes.GrainYield
		rspObj.RoleRes.StoneYield = roleRes.StoneYield
		rspObj.RoleRes.IronYield = roleRes.IronYield
		rspObj.RoleRes.WoodYield = roleRes.WoodYield
		rspObj.RoleRes.DepotCapacity = roleRes.DepotCapacity
		rsp.Body.Code = constant.OK
	}else{
		rsp.Body.Code = constant.RoleNotExist
	}
}