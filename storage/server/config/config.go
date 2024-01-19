package config

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/models"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type IConfig interface {
	GetConfig() *models.Config
}

type Config struct {
	config *models.Config
}

func NewConfig(viper *viper.Viper, logger *zap.Logger) *Config {
	var config models.Config

	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal("error reading config file", zap.Error(err))
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		logger.Fatal("error unmarshaling config file", zap.Error(err))
	}

	config.SetDefault()

	err = config.Validate()
	if err != nil {
		logger.Fatal("error validating config file", zap.Error(err))
	}

	return &Config{config: &config}
}

func (c *Config) Get() *models.Config {
	return c.config
}
