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

func NewConfig(viper *viper.Viper, logger *zap.Logger) IConfig {
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

	var iConfig IConfig = &Config{
		config: &config,
	}

	return iConfig
}

func (c *Config) GetConfig() *models.Config {
	return c.config
}
