package middleware

import (
	"github.com/driftdev/storage/server/utils"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) fiber.Handler {
	return fiberzap.New(
		fiberzap.Config{
			Logger: logger,
			FieldsFunc: func(ctx *fiber.Ctx) []zap.Field {
				return []zap.Field{
					zap.Int("status", ctx.Response().StatusCode()),
					zap.String("method", ctx.Method()),
					zap.String("route", ctx.Route().Path),
					zap.String("path", ctx.Path()),
					zap.String("ip", ctx.IP()),
					zap.String("user-agent", ctx.Get("User-Agent")),
					zap.String("request-id", utils.RequestId(ctx.Context())),
				}
			},
		},
	)
}
