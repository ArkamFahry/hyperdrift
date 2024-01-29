package main

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/api"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/logger"
	"github.com/ArkamFahry/hyperdrift/storage/server/database/migrations"
)

func main() {
	appConfig := config.NewConfig()

	appLogger := logger.NewLogger(appConfig)

	migrations.NewMigrations(appConfig, appLogger)

	api.NewApi(appLogger, appConfig)
}
