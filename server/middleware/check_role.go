package middleware

import (
	"go.uber.org/zap"
	"slgserver/constant"
	"slgserver/log"
	"slgserver/net"
)

func CheckRole() net.MiddlewareFunc {
	return func(next net.HandlerFunc) net.HandlerFunc {
		return func(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

			_, err := req.Conn.GetProperty("role")
			if err != nil {
				rsp.Body.Code = constant.RoleNotInConnect
				log.DefaultLog.Warn("connect not found role",
					zap.String("msgName", req.Body.Name))
				return
			}
			next(req, rsp)
		}
	}
}