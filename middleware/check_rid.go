package middleware

import (
	"github.com/llr104/slgserver/constant"
	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"go.uber.org/zap"
)

func CheckRId() net.MiddlewareFunc {
	return func(next net.HandlerFunc) net.HandlerFunc {
		return func(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

			_, err := req.Conn.GetProperty("rid")
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