package middleware

import (
	"fmt"

	"github.com/llr104/slgserver/log"
	"github.com/llr104/slgserver/net"
	"go.uber.org/zap"
)

func Log() net.MiddlewareFunc {
	return func(next net.HandlerFunc) net.HandlerFunc {
		return func(req *net.WsMsgReq, rsp *net.WsMsgRsp) {

			log.DefaultLog.Info("client req",
				zap.String("msgName", req.Body.Name),
				zap.String("data", fmt.Sprintf("%v", req.Body.Msg)))

			next(req, rsp)
		}
	}
}