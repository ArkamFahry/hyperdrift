package bucket

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type BucketModule struct {
	app    *fiber.App
	db     *pgxpool.Pool
	logger *zap.Logger
	job    *river.Client[pgx.Tx]
}

func NewBucketModule(app *fiber.App, db *pgxpool.Pool, logger *zap.Logger, job *river.Client[pgx.Tx]) {
	bucketService := NewBucketService(db, logger, job)
	NewBucketController(bucketService).RegisterBucketRoutes(app)
}
