package main

import (
	"context"
	"github.com/ArkamFahry/storage/server/config"
	"github.com/ArkamFahry/storage/server/controllers"
	"github.com/ArkamFahry/storage/server/database"
	"github.com/ArkamFahry/storage/server/jobs"
	"github.com/ArkamFahry/storage/server/logger"
	"github.com/ArkamFahry/storage/server/middleware"
	"github.com/ArkamFahry/storage/server/services"
	"github.com/ArkamFahry/storage/server/storage"
	"github.com/ArkamFahry/storage/server/zapfield"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	const op = "main"

	newConfig := config.NewConfig()

	newLogger := logger.NewLogger(newConfig)

	database.NewMigrations(newConfig, newLogger)

	server := fiber.New(fiber.Config{
		ErrorHandler:             middleware.ErrorHandler,
		Immutable:                true,
		EnablePrintRoutes:        true,
		EnableSplittingOnParsers: true,
	})

	server.Use(middleware.Logger(newLogger))

	server.Use(middleware.RequestId())

	server.Use(middleware.KeyAuth(newConfig))

	pgxPoolConfig, err := pgxpool.ParseConfig(newConfig.PostgresUrl)
	if err != nil {
		newLogger.Fatal("error parsing postgres url",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	pgxPoolConfig.ConnConfig.RuntimeParams["search_path"] = "storage"

	pgxPool, err := pgxpool.NewWithConfig(context.Background(), pgxPoolConfig)
	if err != nil {
		newLogger.Fatal("error connecting to postgres",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	s3Config, err := awsConfig.LoadDefaultConfig(
		context.Background(),
		awsConfig.WithRegion(newConfig.S3Region),
		awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(newConfig.S3AccessKeyId, newConfig.S3SecretAccessKey, ""),
		),
	)
	if err != nil {
		newLogger.Fatal("error loading aws s3 config",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	s3Client := s3.NewFromConfig(
		s3Config,
		func(o *s3.Options) {
			o.BaseEndpoint = aws.String(newConfig.S3Endpoint)
			o.UsePathStyle = newConfig.S3ForcePathStyle
			o.EndpointOptions.DisableHTTPS = newConfig.S3DisableSSL
		},
	)

	newStorage := storage.NewStorage(s3Client, newConfig, newLogger)

	riverPgx := riverpgxv5.New(pgxPool)

	riverMigrator := rivermigrate.New[pgx.Tx](riverPgx, nil)

	_, err = riverMigrator.Migrate(context.Background(), rivermigrate.DirectionUp, nil)
	if err != nil {
		newLogger.Fatal("error migrating river jobs schema",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	workers := river.NewWorkers()

	if err = river.AddWorkerSafely[jobs.BucketDeletion](workers, jobs.NewBucketDeletionWorker(pgxPool, newStorage, newLogger)); err != nil {
		newLogger.Fatal("error adding bucket deletion worker",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	if err = river.AddWorkerSafely[jobs.BucketEmptying](workers, jobs.NewBucketEmptyingWorker(pgxPool, newStorage, newLogger)); err != nil {
		newLogger.Fatal("error adding bucket emptying worker",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	if err = river.AddWorkerSafely[jobs.PreSignedUploadSessionCompletion](workers, jobs.NewPreSignedUploadSessionCompletionWorker(pgxPool, newStorage, newLogger)); err != nil {
		newLogger.Fatal("error adding pre signed upload session completion worker",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	if err = river.AddWorkerSafely[jobs.ObjectDeletion](workers, jobs.NewObjectDeletionWorker(pgxPool, newStorage, newLogger)); err != nil {
		newLogger.Fatal("error adding object deletion worker",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	riverClient, err := river.NewClient[pgx.Tx](riverPgx, &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	if err != nil {
		newLogger.Fatal("error creating river client",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	bucketService := services.NewBucketService(pgxPool, riverClient, newLogger)
	controllers.NewBucketController(bucketService).RegisterBucketRoutes(server)

	objectService := services.NewObjectService(pgxPool, newStorage, riverClient, newConfig, newLogger)
	controllers.NewObjectController(objectService).RegisterObjectRoutes(server)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-stop

		newLogger.Info("received interrupt signal. shutting down...", zapfield.Operation(op))

		if err = riverClient.Stop(context.Background()); err != nil {
			newLogger.Error("error stopping river client", zap.Error(err), zapfield.Operation(op))
		}

		pgxPool.Close()

		if err = server.Shutdown(); err != nil {
			newLogger.Error("error shutting down Fiber server", zap.Error(err), zapfield.Operation(op))
		}

		newLogger.Info("shutdown completed...")
		os.Exit(0)
	}()

	err = riverClient.Start(context.Background())
	if err != nil {
		newLogger.Fatal("error starting river client",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	err = server.Listen(":" + newConfig.ServicePort)
	if err != nil {
		newLogger.Fatal("error starting fiber server",
			zap.Error(err),
			zap.String("port", newConfig.ServicePort),
			zapfield.Operation(op),
		)
	}
}
