package main

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/common/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database/migrations"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/logger"
)

func main() {
	appConfig := config.NewConfig()

	appLogger := logger.NewLogger(appConfig)

	migrations.NewMigrations(appConfig, appLogger)

	NewAppModule(appLogger, appConfig)
}
