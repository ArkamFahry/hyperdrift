package middleware

import (
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func Logger(logger *zap.Logger) fiber.Handler {
	return fiberzap.New(
		fiberzap.Config{
			Logger: logger,
			FieldsFunc: func(c *fiber.Ctx) []zap.Field {
				return []zap.Field{
					zap.Int("status", c.Response().StatusCode()),
					zap.String("method", c.Method()),
					zap.String("route", c.Route().Path),
					zap.String("path", c.Path()),
					zap.String("ip", c.IP()),
					zap.String("user-agent", c.Get("User-Agent")),
					zap.String("request-id", c.Context().Value("request_id").(string)),
				}
			},
		},
	)
}
