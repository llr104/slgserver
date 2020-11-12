package middleware

import (
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/log"
	"slgserver/net"
)

func CheckLogin() net.MiddlewareFunc {
	return func(next net.HandlerFunc) net.HandlerFunc {
		return func(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

			_, err := req.Conn.GetProperty("uid")
			if err != nil {
				log.DefaultLog.Warn("connect not found uid",
					zap.String("msgName", req.Body.Name))
				rsp.Body.Code = constant.InvalidParam
				return
			}

			next(req, rsp)
		}
	}
}