package middleware

import (
	"slgserver/constant"
	"slgserver/net"
)

func CheckRole() net.MiddlewareFunc {
	return func(next net.HandlerFunc) net.HandlerFunc {
		return func(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

			_, err := req.Conn.GetProperty("role")
			if err != nil {
				rsp.Body.Code = constant.InvalidParam
				return
			}
			next(req, rsp)
		}
	}
}