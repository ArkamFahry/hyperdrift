package models

import "errors"

type Config struct {
	AppId   string `json:"app_id" mapstructure:"app_id"`
	AppName string `json:"app_name" mapstructure:"app_name"`
	AppHost string `json:"app_host" mapstructure:"app_host"`
	AppPort int    `json:"app_port" mapstructure:"app_port"`

	PostgresUrl string `json:"postgres_url" mapstructure:"postgres_url"`

	S3Endpoint        string `json:"s3_endpoint" mapstructure:"s3_endpoint"`
	S3AccessKeyId     string `json:"s3_access_key_id" mapstructure:"s3_access_key_id"`
	S3SecretAccessKey string `json:"s3_secret_access_key" mapstructure:"s3_secret_access_key"`
	S3Region          string `json:"s3_region" mapstructure:"s3_region"`
	S3Bucket          string `json:"s3_bucket" mapstructure:"s3_bucket"`
	S3ForcePathStyle  bool   `json:"s3_force_path_style" mapstructure:"s3_force_path_style"`
	S3DisableSSL      bool   `json:"s3_disable_ssl" mapstructure:"s3_disable_ssl"`
}

func (c *Config) SetDefault() {
	if c.AppId == "" {
		c.AppId = "hyperdrift-storage"
	}

	if c.AppName == "" {
		c.AppName = "hyperdrift-storage"
	}

	if c.AppHost == "" {
		c.AppHost = "0.0.0.0"
	}

	if c.AppPort == 0 {
		c.AppPort = 8000
	}

	if c.S3Region == "" {
		c.S3Region = "us-east-1"
	}
}

func (c *Config) Validate() error {
	if c.PostgresUrl == "" {
		return errors.New("postgres_url is required in configuration file or environment variable")
	}

	if c.S3Endpoint == "" {
		return errors.New("s3_endpoint is required in configuration file or environment variable")
	}

	if c.S3AccessKeyId == "" {
		return errors.New("s3_access_key_id is required in configuration file or environment variable")
	}

	if c.S3SecretAccessKey == "" {
		return errors.New("s3_secret_access_key is required in configuration file or environment variable")
	}

	if c.S3Region == "" {
		return errors.New("s3_region is required in configuration file or environment variable")
	}

	if c.S3Bucket == "" {
		return errors.New("s3_bucket is required in configuration file or environment variable")
	}

	return nil
}
