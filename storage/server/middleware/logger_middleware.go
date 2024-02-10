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
		},
	)
}
