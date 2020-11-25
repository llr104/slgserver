package controller

import (
	"github.com/goinggo/mapstructure"
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/db"
	"slgserver/log"
	"slgserver/model"
	"slgserver/net"
	"slgserver/server"
	"slgserver/server/logic"
	"slgserver/server/middleware"
	"slgserver/server/proto"
	"slgserver/util"
	"time"
)

var DefaultAccount = Account{}

type Account struct {

}

func (this*Account) InitRouter(r *net.Router) {
	g := r.Group("account").Use(middleware.ElapsedTime(), middleware.Log())

	g.AddRouter("login", this.login)
	g.AddRouter("reLogin", this.reLogin)
	g.AddRouter("logout", this.logout, middleware.CheckLogin())
}

func (this*Account) login(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

	reqObj := &proto.LoginReq{}
	rspObj := &proto.LoginRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj

	user := &model.User{}
	ok, err := db.MasterDB.Table(user).Where("username=?", reqObj.Username).Get(user)
	if err!= nil {
		log.DefaultLog.Error("login db error",
			zap.String("username", reqObj.Username),
			zap.Error(err))

		rsp.Body.Code = constant.DBError
	}else{
		if ok {
			pwd := util.Password(reqObj.Password, user.Passcode)
			if pwd != user.Passwd{
				//密码不正确
				log.DefaultLog.Info("login password not right",
					zap.String("username", user.Username))
				rsp.Body.Code = constant.PwdIncorrect
			}else{
				tt := time.Now()
				s := server.NewSession(user.UId, tt)

				sessStr := s.String()

				log.DefaultLog.Info("login",
					zap.String("username", user.Username),
					zap.String("session", sessStr))

				//登录成功，写记录
				lh := &model.LoginHistory{UId: user.UId, CTime: tt, Ip: reqObj.Ip,
					Hardware: reqObj.Hardware, State: model.Login}
				db.MasterDB.Insert(lh)

				ll := &model.LoginLast{}
				ok, _ := db.MasterDB.Table(ll).Where("uid=?", user.UId).Get(ll)
				if ok {
					ll.IsLogout = 0
					ll.Ip = reqObj.Ip
					ll.LoginTime = time.Now()
					ll.Session = sessStr
					ll.Hardware = reqObj.Hardware

					_, err := db.MasterDB.ID(ll.Id).Cols(
						"is_logout", "ip", "login_time", "session", "hardware").Update(ll)
					if err != nil {
						log.DefaultLog.Error("update login_last table fail", zap.Error(err))
					}


				}else{
					ll = &model.LoginLast{UId: user.UId, LoginTime: tt,
						Ip: reqObj.Ip, Session: sessStr,
						Hardware: reqObj.Hardware, IsLogout: 0}
					db.MasterDB.Insert(ll)
				}

				rspObj.Session = sessStr
				rspObj.Password = reqObj.Password
				rspObj.Username = reqObj.Username
				rspObj.UId = user.UId

				rsp.Body.Code = constant.OK

				server.DefaultConnMgr.UserLogin(req.Conn, sessStr, ll.UId)
			}
		}else{
			//数据库出错
			log.DefaultLog.Info("login username not found", zap.String("username", reqObj.Username))
			rsp.Body.Code = constant.UserNotExist
		}
	}

}

func (this*Account) reLogin(req *net.WsMsgReq, rsp *net.WsMsgRsp) {
	reqObj := &proto.ReLoginReq{}
	rspObj := &proto.ReLoginRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	if reqObj.Session == ""{
		rsp.Body.Code = constant.SessionInvalid
		return
	}

	rsp.Body.Msg = rspObj
	rspObj.Session = reqObj.Session

	sess, err := server.ParseSession(reqObj.Session)
	if err != nil{
		rsp.Body.Code = constant.SessionInvalid
	}else{
		if sess.IsValid() {
			//数据库验证一下
			ll := &model.LoginLast{}
			db.MasterDB.Table(ll).Where("uid=?", sess.Uid).Get(ll)

			if ll.Session == reqObj.Session {
				if ll.Hardware == reqObj.Hardware {
					rsp.Body.Code = constant.OK
					server.DefaultConnMgr.UserLogin(req.Conn, reqObj.Session, ll.UId)

					role, ok := logic.RMgr.Get(reqObj.RId)
					if ok && ll.UId == role.UId{
						req.Conn.SetProperty("role", role)
						server.DefaultConnMgr.RoleEnter(req.Conn)
					}

				}else{
					rsp.Body.Code = constant.HardwareIncorrect
				}
			}else{
				rsp.Body.Code = constant.SessionInvalid
			}
		}else{
			rsp.Body.Code = constant.SessionInvalid
		}
	}
}

func (this*Account) logout(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

	reqObj := &proto.LogoutReq{}
	rspObj := &proto.LogoutRsp{}
	mapstructure.Decode(req.Body.Msg, reqObj)
	rsp.Body.Msg = rspObj
	rspObj.UId = reqObj.UId
	rsp.Body.Code = constant.OK

	log.DefaultLog.Info("logout", zap.Int("uid", reqObj.UId))

	tt := time.Now()
	//登出，写记录
	lh := &model.LoginHistory{UId: reqObj.UId, CTime: tt, State: model.Logout}
	db.MasterDB.Insert(lh)

	ll := &model.LoginLast{}
	ok, _ := db.MasterDB.Table(ll).Where("uid=?", reqObj.UId).Get(ll)
	if ok {
		ll.IsLogout = 1
		ll.LogoutTime = time.Now()
		db.MasterDB.ID(ll.Id).Cols("is_logout", "logout_time").Update(ll)

	}else{
		ll = &model.LoginLast{UId: reqObj.UId, LogoutTime: tt, IsLogout: 0}
		db.MasterDB.Insert(ll)
	}

	server.DefaultConnMgr.UserLogout(req.Conn)

}
