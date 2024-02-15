package jobs

import (
	"context"
	"github.com/ArkamFahry/storage/server/database"
	"github.com/ArkamFahry/storage/server/storage"
	"github.com/ArkamFahry/storage/server/zapfield"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type BucketDeletion struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (BucketDeletion) Kind() string {
	return "bucket.deletion"
}

type BucketDeletionWorker struct {
	queries *database.Queries
	storage *storage.S3Storage
	logger  *zap.Logger
	river.WorkerDefaults[BucketDeletion]
}

func (w *BucketDeletionWorker) Work(ctx context.Context, bucketDeletion *river.Job[BucketDeletion]) error {
	const op = "BucketDeletionWorker.Work"

	limit := int32(100)
	offset := int32(0)

	for {
		objects, err := w.queries.ListObjectsByBucketIdPaged(ctx, &database.ListObjectsByBucketIdPagedParams{
			BucketID: bucketDeletion.Args.Id,
			Offset:   offset,
			Limit:    limit,
		})
		if err != nil {
			w.logger.Error(
				"failed to list objects from queries",
				zap.String("bucket", bucketDeletion.Args.Name),
				zapfield.Operation(op),
				zap.Error(err),
			)
			return err
		}
		if len(objects) == 0 {
			break
		}

		for _, object := range objects {
			err = w.storage.DeleteObject(ctx, &storage.ObjectDelete{
				Bucket: bucketDeletion.Args.Name,
				Name:   object.Name,
			})
			if err != nil {
				w.logger.Error(
					"failed to delete object from storage",
					zap.String("bucket", bucketDeletion.Args.Name),
					zap.String("name", object.Name),
					zapfield.Operation(op),
					zap.Error(err),
				)
				return err
			}
			err = w.queries.DeleteObject(ctx, object.ID)
			if err != nil {
				w.logger.Error(
					"failed to delete object from database",
					zap.String("bucket", bucketDeletion.Args.Name),
					zap.String("name", object.Name),
					zapfield.Operation(op),
					zap.Error(err),
				)
				return err
			}
		}

		offset += limit
	}

	err := w.queries.DeleteBucket(ctx, bucketDeletion.Args.Id)
	if err != nil {
		w.logger.Error(
			"failed to delete bucket from database",
			zap.String("bucket", bucketDeletion.Args.Name),
			zapfield.Operation(op),
			zap.Error(err),
		)
		return err
	}
	if err != nil {
		return err
	}

	return nil
}

func NewBucketDeletionWorker(db *pgxpool.Pool, storage *storage.S3Storage, logger *zap.Logger) *BucketDeletionWorker {
	return &BucketDeletionWorker{
		queries: database.New(db),
		storage: storage,
		logger:  logger,
	}
}
