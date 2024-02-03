package main

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket"
	bucketJobs "github.com/ArkamFahry/hyperdrift/storage/server/bucket/jobs"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database/migrations"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/logger"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/storage"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/zapfield"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/contrib/fiberzap/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"go.uber.org/zap"
)

func NewAppModule() {
	const op = "AppModule.NewAppModule"

	appConfig := config.NewConfig()

	appLogger := logger.NewLogger(appConfig)

	migrations.NewMigrations(appConfig, appLogger)

	appServer := fiber.New(fiber.Config{
		Immutable: true,
	})

	appServer.Use(fiberzap.New(fiberzap.Config{
		Logger: appLogger,
	}))

	pgxPool, err := pgxpool.New(context.Background(), appConfig.PostgresUrl)
	if err != nil {
		appLogger.Fatal("error connecting to postgres",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	s3Config, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(appConfig.S3Region),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(appConfig.S3AccessKeyId, appConfig.S3SecretAccessKey, ""),
		),
	)

	s3Client := s3.NewFromConfig(
		s3Config,
		func(o *s3.Options) {
			o.BaseEndpoint = aws.String(appConfig.S3Endpoint)
			o.UsePathStyle = appConfig.S3ForcePathStyle
			o.EndpointOptions.DisableHTTPS = appConfig.S3DisableSSL
		},
	)

	appStorage := storage.NewS3Storage(s3Client, appConfig, appLogger)

	workers := river.NewWorkers()

	if err = river.AddWorkerSafely[bucketJobs.BucketDelete](workers, bucketJobs.NewBucketDeleteJob(pgxPool, appStorage, appLogger)); err != nil {
		appLogger.Fatal("error adding bucket delete worker",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	if err = river.AddWorkerSafely[bucketJobs.BucketEmpty](workers, bucketJobs.NewBucketEmptyJob(pgxPool, appStorage, appLogger)); err != nil {
		appLogger.Fatal("error adding bucket empty worker",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	riverClient, err := river.NewClient[pgx.Tx](riverpgxv5.New(pgxPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		appLogger.Fatal("error creating river client",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	bucket.NewBucketModule(appServer, pgxPool, appLogger, riverClient)

	err = appServer.Listen(":" + appConfig.ServerPort)
	if err != nil {
		appLogger.Fatal("error starting fiber server",
			zap.Error(err),
			zap.String("port", appConfig.ServerPort),
			zapfield.Operation(op),
		)
	}
}
