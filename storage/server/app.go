package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/ArkamFahry/hyperdrift/storage/server/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/controllers"
	"github.com/ArkamFahry/hyperdrift/storage/server/database/migrations"
	"github.com/ArkamFahry/hyperdrift/storage/server/jobs"
	"github.com/ArkamFahry/hyperdrift/storage/server/logger"
	"github.com/ArkamFahry/hyperdrift/storage/server/middleware"
	"github.com/ArkamFahry/hyperdrift/storage/server/services"
	"github.com/ArkamFahry/hyperdrift/storage/server/storage"
	"github.com/ArkamFahry/hyperdrift/storage/server/zapfield"
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

func NewApp() {
	const op = "AppModule.NewAppModule"

	appConfig := config.NewConfig()

	appLogger := logger.NewLogger(appConfig)

	migrations.NewMigrations(appConfig, appLogger)

	appServer := fiber.New(fiber.Config{
		ErrorHandler:      middleware.ErrorHandler,
		Immutable:         true,
		EnablePrintRoutes: true,
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

	if err = river.AddWorkerSafely[jobs.BucketDelete](workers, jobs.NewBucketDeleteJob(pgxPool, appStorage, appLogger)); err != nil {
		appLogger.Fatal("error adding bucket delete worker",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	if err = river.AddWorkerSafely[jobs.BucketEmpty](workers, jobs.NewBucketEmptyJob(pgxPool, appStorage, appLogger)); err != nil {
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

	bucketService := services.NewBucketService(pgxPool, appLogger, riverClient)
	controllers.NewBucketController(bucketService).RegisterBucketRoutes(appServer)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop

		appLogger.Info("received interrupt signal. shutting down...", zapfield.Operation(op))

		if err = riverClient.Stop(context.Background()); err != nil {
			appLogger.Error("error stopping river client", zap.Error(err), zapfield.Operation(op))
		}

		pgxPool.Close()

		if err = appServer.Shutdown(); err != nil {
			appLogger.Error("error shutting down Fiber server", zap.Error(err), zapfield.Operation(op))
		}

		appLogger.Info("shutdown completed...")
		os.Exit(0)
	}()

	go func() {
		err = riverClient.Start(context.Background())
		if err != nil {
			appLogger.Fatal("error starting river client",
				zap.Error(err),
				zapfield.Operation(op),
			)
		}
	}()

	err = appServer.Listen(":" + appConfig.ServerPort)
	if err != nil {
		appLogger.Fatal("error starting fiber server",
			zap.Error(err),
			zap.String("port", appConfig.ServerPort),
			zapfield.Operation(op),
		)
	}
}
