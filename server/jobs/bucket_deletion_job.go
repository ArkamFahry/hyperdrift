package jobs

import (
	"context"
	"github.com/driftdev/storage/server/database"
	"github.com/driftdev/storage/server/storage"
	"github.com/driftdev/storage/server/zapfield"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type BucketDeletion struct {
	BucketId string `json:"bucket_id"`
}

func (BucketDeletion) Kind() string {
	return "bucket.deletion"
}

type BucketDeletionWorker struct {
	queries *database.Queries
	storage *storage.Storage
	logger  *zap.Logger
	river.WorkerDefaults[BucketDeletion]
}

func (w *BucketDeletionWorker) Work(ctx context.Context, bucketDeletion *river.Job[BucketDeletion]) error {
	const op = "BucketDeletionWorker.Work"

	bucket, err := w.queries.BucketGetById(ctx, bucketDeletion.Args.BucketId)
	if err != nil {
		w.logger.Error(
			"failed to get bucket",
			zap.String("bucket_id", bucketDeletion.Args.BucketId),
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

	err = w.queries.BucketDelete(ctx, bucket.ID)
	if err != nil {
		w.logger.Error(
			"failed to delete bucket from database",
			zap.String("bucket_id", bucket.ID),
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

func NewBucketDeletionWorker(db *pgxpool.Pool, storage *storage.Storage, logger *zap.Logger) *BucketDeletionWorker {
	return &BucketDeletionWorker{
		queries: database.New(db),
		storage: storage,
		logger:  logger,
	}
}
