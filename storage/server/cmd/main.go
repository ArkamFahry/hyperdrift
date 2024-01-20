package main

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/api"
	"github.com/ArkamFahry/hyperdrift/storage/server/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/logger"
	"github.com/spf13/viper"
)

func main() {
	newViper := viper.New()
	newLogger := logger.NewLogger(newViper)
	newConfig := config.NewConfig(newViper, newLogger)

	api.NewApi(newLogger, newConfig)
}
