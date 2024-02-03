package main

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database/migrations"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/logger"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/zapfield"
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

	port := appConfig.ServerPort

	pgxPool, err := pgxpool.New(context.Background(), appConfig.PostgresUrl)
	if err != nil {
		appLogger.Fatal("error connecting to postgres",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	riverPgx := riverpgxv5.New(pgxPool)

	riverClient, err := river.NewClient[pgx.Tx](riverPgx, nil)
	if err != nil {
		appLogger.Fatal("error creating river client",
			zap.Error(err),
			zapfield.Operation(op),
		)
	}

	bucket.NewBucketModule(appServer, pgxPool, appLogger, riverClient)

	err = appServer.Listen(":" + port)
	if err != nil {
		appLogger.Fatal("error starting fiber server",
			zap.Error(err),
			zap.String("port", port),
			zapfield.Operation(op),
		)
	}
}
