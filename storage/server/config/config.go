package config

import (
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	ServerId          string `json:"server_id" mapstructure:"server_id"`
	ServerName        string `json:"server_name" mapstructure:"server_name"`
	ServerEnvironment string `json:"server_environment" mapstructure:"server_environment"`
	ServerHost        string `json:"server_host" mapstructure:"server_host"`
	ServerPort        string `json:"server_port" mapstructure:"server_port"`

	PostgresUrl string `json:"postgres_url" mapstructure:"postgres_url"`

	S3Endpoint        string `json:"s3_endpoint" mapstructure:"s3_endpoint"`
	S3AccessKeyId     string `json:"s3_access_key_id" mapstructure:"s3_access_key_id"`
	S3SecretAccessKey string `json:"s3_secret_access_key" mapstructure:"s3_secret_access_key"`
	S3Region          string `json:"s3_region" mapstructure:"s3_region"`
	S3Bucket          string `json:"s3_bucket" mapstructure:"s3_bucket"`
	S3ForcePathStyle  bool   `json:"s3_force_path_style" mapstructure:"s3_force_path_style"`
	S3DisableSSL      bool   `json:"s3_disable_ssl" mapstructure:"s3_disable_ssl"`
}

func NewConfig(viper *viper.Viper, logger *zap.Logger) *Config {
	var config Config

	err := viper.ReadInConfig()
	if err != nil {
		logger.Fatal("error reading config file", zap.Error(err))
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		logger.Fatal("error unmarshaling config file", zap.Error(err))
	}

	if err != nil {
		logger.Fatal("error validating config file", zap.Error(err))
	}

	return &config
}
