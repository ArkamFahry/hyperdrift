package config

import (
	"errors"
	"github.com/ArkamFahry/storage/server/zapfield"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

type Config struct {
	ServiceId          string `json:"service_id" mapstructure:"service_id"`
	ServiceName        string `json:"service_name" mapstructure:"service_name"`
	ServiceEnvironment string `json:"service_environment" mapstructure:"service_environment"`
	ServiceHost        string `json:"service_host" mapstructure:"service_host"`
	ServicePort        string `json:"service_port" mapstructure:"service_port"`
	ServiceApiKey      string `json:"service_api_key" mapstructure:"service_api_key"`

	PostgresUrl string `json:"postgres_url" mapstructure:"postgres_url"`

	S3Endpoint        string `json:"s3_endpoint" mapstructure:"s3_endpoint"`
	S3AccessKeyId     string `json:"s3_access_key_id" mapstructure:"s3_access_key_id"`
	S3SecretAccessKey string `json:"s3_secret_access_key" mapstructure:"s3_secret_access_key"`
	S3Bucket          string `json:"s3_bucket" mapstructure:"s3_bucket"`
	S3Region          string `json:"s3_region" mapstructure:"s3_region"`
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

	DefaultPreSignedUploadUrlExpiry   int64 `json:"default_pre_signed_upload_url_expiry" mapstructure:"default_pre_signed_upload_url_expiry"`
	DefaultPreSignedDownloadUrlExpiry int64 `json:"default_pre_signed_download_url_expiry" mapstructure:"default_pre_signed_download_url_expiry"`
}

func (c *Config) SetDefaults() {
	if c.ServiceId == "" {
		c.ServiceId = uuid.New().String()
	}

	if c.ServiceName == "" {
		c.ServiceName = "hyperdrift-storage"
	}

	if c.ServiceEnvironment == "" {
		c.ServiceEnvironment = "prod"
	}

	if c.ServiceHost == "" {
		c.ServiceHost = "0.0.0.0"
	}

	if c.ServicePort == "" {
		c.ServicePort = "3001"
	}

	if c.S3Region == "" {
		c.S3Region = "us-east-1"
	}

	if c.DefaultPreSignedUploadUrlExpiry == 0 {
		c.DefaultPreSignedUploadUrlExpiry = 120
	}

	if c.DefaultPreSignedDownloadUrlExpiry == 0 {
		c.DefaultPreSignedDownloadUrlExpiry = 300
	}
}

func (c *Config) IsValid() error {
	if c.ServiceApiKey == "" {
		return errors.New("server_api_key is a required")
	}

	if c.PostgresUrl == "" {
		return errors.New("postgres_url is a required")
	}

	if c.S3Endpoint == "" {
		return errors.New("s3_endpoint is a required")
	}

	if c.S3AccessKeyId == "" {
		return errors.New("s3_access_key_id is a required")
	}

	if c.S3SecretAccessKey == "" {
		return errors.New("s3_secret_access_key is a required")
	}

	if c.S3Bucket == "" {
		return errors.New("s3_bucket_name is a required")
	}

	return nil
}

func NewConfig() *Config {
	const op = "config.NewConfig"

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
		logger.Fatal("error reading config", zap.Error(err), zapfield.Operation(op))
	}

	err = v.Unmarshal(&config)
	if err != nil {
		logger.Fatal("error unmarshaling config", zap.Error(err), zapfield.Operation(op))
	}

	config.SetDefaults()

	err = config.IsValid()
	if err != nil {
		logger.Fatal("error validating config", zap.Error(err), zapfield.Operation(op))
	}

	return &config
}
