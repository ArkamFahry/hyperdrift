package config

import (
	"errors"
	"github.com/google/uuid"
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
	S3BucketName      string `json:"s3_bucket_name" mapstructure:"s3_bucket_name"`
	S3ForcePathStyle  bool   `json:"s3_force_path_style" mapstructure:"s3_force_path_style"`
	S3DisableSSL      bool   `json:"s3_disable_ssl" mapstructure:"s3_disable_ssl"`

	DefaultBuckets []struct {
		Id                   string   `json:"id" mapstructure:"id"`
		Name                 string   `json:"name" mapstructure:"name"`
		AllowedMimeTypes     []string `json:"allowed_mime_types" mapstructure:"allowed_mime_types"`
		MaxAllowedObjectSize *int64   `json:"max_allowed_object_size" mapstructure:"max_allowed_object_size"`
		Public               bool     `json:"public" mapstructure:"public"`
		Disabled             bool     `json:"disabled" mapstructure:"disabled"`
	} `json:"default_buckets" mapstructure:"default_buckets"`

	DefaultPreSignedUploadUrlExpiresIn   int `json:"default_pre_signed_upload_url_expires_in" mapstructure:"default_pre_signed_upload_url_expires_in"`
	DefaultPreSignedDownloadUrlExpiresIn int `json:"default_pre_signed_download_url_expires_in" mapstructure:"default_pre_signed_download_url_expires_in"`
}

func NewConfig() *Config {
	var config Config

	v := viper.New()
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	v.AddConfigPath(".")
	v.AddConfigPath("../")

	err = v.ReadInConfig()
	if err != nil {
		logger.Fatal("error reading config", zap.Error(err))
	}

	err = v.Unmarshal(&config)
	if err != nil {
		logger.Fatal("error unmarshaling config", zap.Error(err))
	}

	setDefaultConfig(&config)

	err = validateConfig(&config)
	if err != nil {
		logger.Fatal("error validating config", zap.Error(err))
	}

	return &config
}

func setDefaultConfig(config *Config) {
	if config.ServerId == "" {
		config.ServerId = uuid.New().String()
	}

	if config.ServerName == "" {
		config.ServerName = "hyperdrift-storage"
	}

	if config.ServerEnvironment == "" {
		config.ServerEnvironment = "prod"
	}

	if config.ServerHost == "" {
		config.ServerHost = "0.0.0.0"
	}

	if config.ServerPort == "" {
		config.ServerPort = "3001"
	}

	if config.S3Region == "" {
		config.S3Region = "us-east-1"
	}

	if config.DefaultPreSignedUploadUrlExpiresIn == 0 {
		config.DefaultPreSignedUploadUrlExpiresIn = 1800
	}

	if config.DefaultPreSignedDownloadUrlExpiresIn == 0 {
		config.DefaultPreSignedDownloadUrlExpiresIn = 1800
	}
}

func validateConfig(config *Config) error {
	if config.PostgresUrl == "" {
		return errors.New("postgres_url is a required")
	}

	if config.S3Endpoint == "" {
		return errors.New("s3_endpoint is a required")
	}

	if config.S3AccessKeyId == "" {
		return errors.New("s3_access_key_id is a required")
	}

	if config.S3SecretAccessKey == "" {
		return errors.New("s3_secret_access_key is a required")
	}

	if config.S3BucketName == "" {
		return errors.New("s3_bucket_name is a required")
	}

	return nil
}
