package logger

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/config"
	"go.uber.org/zap"
)

func NewLogger(config *config.Config) *zap.Logger {
	var logger *zap.Logger
	var err error

	serverEnvironment := config.ServerEnvironment

	switch serverEnvironment {
	case "dev":
		logger, err = zap.NewDevelopment()
	case "test":
		logger, err = zap.NewDevelopment()
	case "prod":
		logger, err = zap.NewProduction()
	default:
		logger, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
	}

	return logger
}
