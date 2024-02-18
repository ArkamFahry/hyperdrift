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

type BucketEmptying struct {
	BucketId string `json:"bucket_id"`
}

func (BucketEmptying) Kind() string {
	return "bucket.emptying"
}

type BucketEmptyingWorker struct {
	queries *database.Queries
	storage *storage.S3Storage
	logger  *zap.Logger
	river.WorkerDefaults[BucketEmptying]
}

func (w *BucketEmptyingWorker) Work(ctx context.Context, bucketEmpty *river.Job[BucketEmptying]) error {
	const op = "BucketEmptyingWorker.Work"

	bucket, err := w.queries.BucketGetById(ctx, bucketEmpty.Args.BucketId)
	if err != nil {
		w.logger.Error(
			"failed to get bucket",
			zap.String("bucket_id", bucketEmpty.Args.BucketId),
			zapfield.Operation(op),
			zap.Error(err),
		)
		return err
	}

	limit := int32(100)
	offset := int32(0)

	for {
		objects, err := w.queries.ObjectsListBucketIdPaged(ctx, &database.ObjectsListBucketIdPagedParams{
			BucketID: bucket.ID,
			Offset:   offset,
			Limit:    limit,
		})
		if err != nil {
			w.logger.Error(
				"failed to list objects",
				zap.String("bucket_id", bucket.ID),
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
				Bucket: bucket.Name,
				Name:   object.Name,
			})
			if err != nil {
				w.logger.Error(
					"failed to delete object from storage",
					zap.String("bucket_name", bucket.Name),
					zap.String("object_name", object.Name),
					zapfield.Operation(op),
					zap.Error(err),
				)
				return err
			}
			err = w.queries.ObjectDelete(ctx, object.ID)
			if err != nil {
				w.logger.Error(
					"failed to delete object from database",
					zap.String("object_id", object.ID),
					zapfield.Operation(op),
					zap.Error(err),
				)
				return err
			}
		}

		offset += limit
	}

	err = w.queries.BucketUnlock(ctx, bucket.ID)
	if err != nil {
		w.logger.Error(
			"failed to unlock bucket from database",
			zap.String("bucket_id", bucket.ID),
			zapfield.Operation(op),
			zap.Error(err),
		)
		return err
	}

	return nil
}

func NewBucketEmptyingWorker(db *pgxpool.Pool, storage *storage.S3Storage, logger *zap.Logger) *BucketEmptyingWorker {
	return &BucketEmptyingWorker{
		queries: database.New(db),
		storage: storage,
		logger:  logger,
	}
}
