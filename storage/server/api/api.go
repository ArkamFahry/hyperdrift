package api

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/config"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func NewApi(logger *zap.Logger, config *config.Config) {

	app := fiber.New(fiber.Config{
		Immutable: true,
	})

	app.Use(fiberzap.New(fiberzap.Config{
		Logger: logger,
	}))

	port := config.ServerPort

	err := app.Listen(":" + port)
	if err != nil {
		logger.Fatal("error starting fiber server",
			zap.Error(err),
			zap.String("port", port),
		)
	}
}
