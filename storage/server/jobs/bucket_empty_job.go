package jobs

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/storage"
	"github.com/ArkamFahry/hyperdrift/storage/server/zapfield"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type BucketEmpty struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (BucketEmpty) Kind() string {
	return "bucket.empty"
}

type BucketEmptyWorker struct {
	queries *database.Queries
	storage *storage.S3Storage
	logger  *zap.Logger
	river.WorkerDefaults[BucketEmpty]
}

func (w *BucketEmptyWorker) Work(ctx context.Context, bucketEmpty *river.Job[BucketEmpty]) error {
	const op = "BucketEmptyWorker.Work"

	limit := int32(100)
	offset := int32(0)

	for {
		objects, err := w.queries.ListObjectsByBucketIdPaged(ctx, &database.ListObjectsByBucketIdPagedParams{
			BucketID: bucketEmpty.Args.Id,
			Offset:   offset,
			Limit:    limit,
		})
		if err != nil {
			w.logger.Error(
				"failed to list objects from queries",
				zap.String("bucket", bucketEmpty.Args.Name),
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
				Bucket: bucketEmpty.Args.Name,
				Name:   object.Name,
			})
			if err != nil {
				w.logger.Error(
					"failed to delete object from storage",
					zap.String("bucket", bucketEmpty.Args.Name),
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
					zap.String("bucket", bucketEmpty.Args.Name),
					zap.String("name", object.Name),
					zapfield.Operation(op),
					zap.Error(err),
				)
				return err
			}
		}

		offset += limit
	}

	err := w.queries.UnlockBucket(ctx, bucketEmpty.Args.Id)
	if err != nil {
		w.logger.Error(
			"failed to unlock bucket from database",
			zap.String("bucket", bucketEmpty.Args.Name),
			zapfield.Operation(op),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func NewBucketEmptyWorker(db *pgxpool.Pool, storage *storage.S3Storage, logger *zap.Logger) *BucketEmptyWorker {
	return &BucketEmptyWorker{
		queries: database.New(db),
		storage: storage,
		logger:  logger,
	}
}
