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
	g.AddRouter("myCity", this.myCity)
}

func (this*Role) create(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.CreateRoleReq{}
	rspObj := &proto.CreateRoleRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	uid, _ := req.Conn.GetProperty("uid")
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
		req.Conn.SetProperty("role", r)
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

	r, err := req.Conn.GetProperty("role")
	if err != nil{
		if err != nil {
			log.DefaultLog.Warn("myCity but connect not found sid")
			rsp.Body.Code = constant.InvalidParam
			return
		}
	}

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
				c := model.MapRoleCity{RId: role.RId, X: x, Y: y, IsMain: 1,
					Durable: 100, Level: 1, Name: role.NickName, CreatedAt: time.Now()}

				//插入
				cityId, err := db.MasterDB.Table(c).Insert(c)
				if err != nil{
					rsp.Body.Code = constant.DBError
				}else{
					c.CityId = int(cityId)
					citys = append(citys, c)
					//更新城市缓存
					entity.RCMgr.Add(&c)
				}
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
		rspObj.Citys[i].Durable = v.Durable
	}

}