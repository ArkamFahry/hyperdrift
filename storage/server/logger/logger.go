package logger

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

func NewLogger(viper *viper.Viper) *zap.Logger {
	var logger *zap.Logger
	var err error

	logLevel := viper.GetString("LOG_LEVEL")

	switch logLevel {
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
