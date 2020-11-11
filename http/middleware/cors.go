package middleware

import (
	"github.com/labstack/echo/v4"
)

func Cors() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(ctx echo.Context) error {

			rsp := ctx.Response()
			rsp.Header().Set("Access-Control-Allow-Origin", "*")
			rsp.Header().Set("Access-Control-Allow-Origin", "*")
			rsp.Header().Set("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
			rsp.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS, POST, PUT, DELETE")
			rsp.Header().Set("Content-Type", "application/json")

			if err := next(ctx); err != nil {
				return err
			}

			return nil
		}
	}
}

