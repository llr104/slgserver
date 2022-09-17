package controller

import (
	"math/rand"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/goinggo/mapstructure"
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/db"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/middleware"
	"github.com/llr104/slgserver/net"
	"github.com/llr104/slgserver/server/slgserver/global"
	"github.com/llr104/slgserver/server/slgserver/logic/mgr"
	"github.com/llr104/slgserver/server/slgserver/model"
	"github.com/llr104/slgserver/server/slgserver/pos"
	"github.com/llr104/slgserver/server/slgserver/proto"
	"github.com/llr104/slgserver/server/slgserver/static_conf"
	"github.com/llr104/slgserver/util"
	"go.uber.org/zap"
)

var DefaultRole = Role{}

type Role struct {
}

func (this *Role) InitRouter(r *net.Router) {
	g := r.Group("role").Use(middleware.ElapsedTime(), middleware.Log())

	g.AddRouter("enterServer", this.enterServer)
	g.AddRouter("create", this.create, middleware.CheckLogin())
	g.AddRouter("roleList", this.roleList, middleware.CheckLogin())
	g.AddRouter("myCity", this.myCity, middleware.CheckRole())
	g.AddRouter("myRoleRes", this.myRoleRes, middleware.CheckRole())
	g.AddRouter("myRoleBuild", this.myRoleBuild, middleware.CheckRole())
	g.AddRouter("myProperty", this.myProperty, middleware.CheckRole())
	g.AddRouter("upPosition", this.upPosition, middleware.CheckRole())
	g.AddRouter("posTagList", this.posTagList, middleware.CheckRole())
	g.AddRouter("opPosTag", this.opPosTag, middleware.CheckRole())
}

func (this *Role) create(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.CreateRoleReq{}
	rspObj := &proto.CreateRoleRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	uid, _ := req.Conn.GetProperty("uid")
	reqObj.UId = uid.(int)
	rspObj.Role.UId = reqObj.UId

	r := make([]model.Role, 0)
	has, _ := db.MasterDB.Table(r).Where("uid=?", reqObj.UId).Get(r)
	if has {
		log.DefaultLog.Info("role has create", zap.Int("uid", reqObj.UId))
		rsp.Body.Code = constant.RoleAlreadyCreate
	} else {

		role := &model.Role{UId: reqObj.UId, HeadId: reqObj.HeadId, Sex: reqObj.Sex,
			NickName: reqObj.NickName, CreatedAt: time.Now()}

		if _, err := db.MasterDB.Insert(role); err != nil {
			log.DefaultLog.Info("role  create error",
				zap.Int("uid", reqObj.UId), zap.Error(err))
			e, _ := err.(*mysql.MySQLError)
			if 1062 == e.Number {
				rsp.Body.Code = constant.RoleNameExist
			} else {
				rsp.Body.Code = constant.DBError
			}
		} else {
			rspObj.Role.RId = role.RId
			rspObj.Role.UId = reqObj.UId
			rspObj.Role.NickName = reqObj.NickName
			rspObj.Role.Sex = reqObj.Sex
			rspObj.Role.HeadId = reqObj.HeadId

			rsp.Body.Code = constant.OK
		}
	}
}

func (this *Role) roleList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.RoleListReq{}
	rspObj := &proto.RoleListRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	uid, _ := req.Conn.GetProperty("uid")
	uid = uid.(int)

	r := make([]*model.Role, 0)
	err := db.MasterDB.Table(r).Where("uid=?", uid).Find(&r)
	if err == nil {
		rl := make([]proto.Role, len(r))
		for i, v := range r {
			rl[i] = v.ToProto().(proto.Role)
		}
		rspObj.Roles = rl
		rsp.Body.Code = constant.OK
	} else {
		rsp.Body.Code = constant.DBError
	}
}

func (this *Role) enterServer(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.EnterServerReq{}
	rspObj := &proto.EnterServerRsp{}
	rspObj.Time = time.Now().UnixNano() / 1e6

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	//校验session是否合法
	sess, err := util.ParseSession(reqObj.Session)
	if err != nil || sess.IsValid() == false {
		rsp.Body.Code = constant.SessionInvalid
		return
	}
	uid := sess.Id
	req.Conn.SetProperty("uid", uid)

	role := &model.Role{}
	b, err := db.MasterDB.Table(role).Where("uid=?", uid).Get(role)
	if err != nil {
		log.DefaultLog.Warn("enterServer db error", zap.Error(err))
		rsp.Body.Code = constant.DBError
		return
	}
	if b {
		rsp.Body.Code = constant.OK
		rspObj.Role = role.ToProto().(proto.Role)
		req.Conn.SetProperty("role", role)
		net.ConnMgr.RoleEnter(req.Conn, role.RId)

		var e error = nil
		roleRes, ok := mgr.RResMgr.Get(role.RId)
		if ok == false {

			roleRes = &model.RoleRes{RId: role.RId,
				Wood:   static_conf.Basic.Role.Wood,
				Iron:   static_conf.Basic.Role.Iron,
				Stone:  static_conf.Basic.Role.Stone,
				Grain:  static_conf.Basic.Role.Grain,
				Gold:   static_conf.Basic.Role.Gold,
				Decree: static_conf.Basic.Role.Decree}
			_, e = db.MasterDB.Insert(roleRes)
			if e != nil {
				log.DefaultLog.Error("insert rres error", zap.Error(e))
			}
		}

		if e == nil {
			mgr.RResMgr.Add(roleRes)
			rspObj.RoleRes = roleRes.ToProto().(proto.RoleRes)
			rsp.Body.Code = constant.OK
		} else {
			rsp.Body.Code = constant.DBError
			return
		}

		//玩家的一些属性
		if _, ok := mgr.RAttrMgr.TryCreate(role.RId); ok == false {
			rsp.Body.Code = constant.DBError
			return
		}

		//查询是否有城市
		_, ok = mgr.RCMgr.GetByRId(role.RId)
		if ok == false {
			citys := make([]*model.MapRoleCity, 0)
			//随机生成一个城市
			for true {
				x := rand.Intn(global.MapWith)
				y := rand.Intn(global.MapHeight)
				if mgr.NMMgr.IsCanBuildCity(x, y) {
					//建立城市
					c := &model.MapRoleCity{RId: role.RId, X: x, Y: y,
						IsMain:     1,
						CurDurable: static_conf.Basic.City.Durable,
						Name:       role.NickName,
						CreatedAt:  time.Now(),
					}

					//插入
					_, err := db.MasterDB.Table(c).Insert(c)
					if err != nil {
						rsp.Body.Code = constant.DBError
					} else {
						citys = append(citys, c)
						//更新城市缓存
						mgr.RCMgr.Add(c)
					}

					//生成城市里面的设施
					mgr.RFMgr.GetAndTryCreate(c.CityId, c.RId)
					break
				}
			}
		}
		rspObj.Token = util.NewSession(role.RId, time.Now()).String()
	} else {
		rsp.Body.Code = constant.RoleNotExist
	}
}

func (this *Role) myCity(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyCityReq{}
	rspObj := &proto.MyCityRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role, _ := r.(*model.Role)

	citys, ok := mgr.RCMgr.GetByRId(role.RId)
	if ok {
		rspObj.Citys = make([]proto.MapRoleCity, len(citys))
		//赋值发送
		for i, v := range citys {
			rspObj.Citys[i] = v.ToProto().(proto.MapRoleCity)
		}

	} else {
		rspObj.Citys = make([]proto.MapRoleCity, 0)
	}

}

func (this *Role) myRoleRes(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyRoleResReq{}
	rspObj := &proto.MyRoleResRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	roleRes, ok := mgr.RResMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.RoleNotExist
		return
	} else {
		rspObj.RoleRes = roleRes.ToProto().(proto.RoleRes)
		rsp.Body.Code = constant.OK
	}
}

func (this *Role) myProperty(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyRolePropertyReq{}
	rspObj := &proto.MyRolePropertyRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	//城市
	c, ok := mgr.RCMgr.GetByRId(role.RId)
	if ok {
		rspObj.Citys = make([]proto.MapRoleCity, len(c))
		for i, v := range c {
			rspObj.Citys[i] = v.ToProto().(proto.MapRoleCity)
		}
	} else {
		rspObj.Citys = make([]proto.MapRoleCity, 0)
	}

	//建筑
	ra, ok := mgr.RBMgr.GetRoleBuild(role.RId)
	if ok {
		rspObj.MRBuilds = make([]proto.MapRoleBuild, len(ra))
		for i, v := range ra {
			rspObj.MRBuilds[i] = v.ToProto().(proto.MapRoleBuild)
		}
	} else {
		rspObj.MRBuilds = make([]proto.MapRoleBuild, 0)
	}

	//资源
	roleRes, ok := mgr.RResMgr.Get(role.RId)
	if ok {
		rspObj.RoleRes = roleRes.ToProto().(proto.RoleRes)
	} else {
		rsp.Body.Code = constant.RoleNotExist
		return
	}

	//武将
	gs, ok := mgr.GMgr.GetOrCreateByRId(role.RId)
	if ok {
		rspObj.Generals = make([]proto.General, 0)
		for _, v := range gs {
			rspObj.Generals = append(rspObj.Generals, v.ToProto().(proto.General))
		}
	} else {
		rsp.Body.Code = constant.DBError
		return
	}

	//军队
	ar, ok := mgr.AMgr.GetByRId(role.RId)
	if ok {
		rspObj.Armys = make([]proto.Army, len(ar))
		for i, v := range ar {
			rspObj.Armys[i] = v.ToProto().(proto.Army)
		}
	} else {
		rspObj.Armys = make([]proto.Army, 0)
	}

}

func (this *Role) myRoleBuild(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.MyRoleBuildReq{}
	rspObj := &proto.MyRoleBuildRsp{}

	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)

	ra, ok := mgr.RBMgr.GetRoleBuild(role.RId)
	if ok {
		rspObj.MRBuilds = make([]proto.MapRoleBuild, len(ra))
		for i, v := range ra {
			rspObj.MRBuilds[i] = v.ToProto().(proto.MapRoleBuild)
		}
	} else {
		rspObj.MRBuilds = make([]proto.MapRoleBuild, 0)
	}

}

func (this *Role) upPosition(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.UpPositionReq{}
	rspObj := &proto.UpPositionRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	rspObj.X = reqObj.X
	rspObj.Y = reqObj.Y

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	pos.RPMgr.Push(reqObj.X, reqObj.Y, role.RId)

}

func (this *Role) posTagList(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.PosTagListReq{}
	rspObj := &proto.PosTagListRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	attr, ok := mgr.RAttrMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.RoleNotExist
		return
	}

	rspObj.PosTags = attr.PosTagArray

}

func (this *Role) opPosTag(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.PosTagReq{}
	rspObj := &proto.PosTagRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rsp.Body.Code = constant.OK

	rspObj.X = reqObj.X
	rspObj.Y = reqObj.Y
	rspObj.Type = reqObj.Type
	rspObj.Name = reqObj.Name

	r, _ := req.Conn.GetProperty("role")
	role := r.(*model.Role)
	attr, ok := mgr.RAttrMgr.Get(role.RId)
	if ok == false {
		rsp.Body.Code = constant.RoleNotExist
		return
	}
	if reqObj.Type == 0 {
		attr.RemovePosTag(reqObj.X, reqObj.Y)
		attr.SyncExecute()
	} else if reqObj.Type == 1 {

		limit := static_conf.Basic.Role.PosTagLimit
		if int(limit) >= len(attr.PosTagArray) {
			attr.AddPosTag(reqObj.X, reqObj.Y, reqObj.Name)
			attr.SyncExecute()
		} else {
			rsp.Body.Code = constant.OutPosTagLimit
		}
	} else {
		rsp.Body.Code = constant.InvalidParam
	}
}
