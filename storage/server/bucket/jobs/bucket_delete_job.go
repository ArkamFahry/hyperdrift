package jobs

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/storage"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/zapfield"
	"github.com/jackc/pgx/v5"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type BucketDeleteArgs struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func (BucketDeleteArgs) Kind() string {
	return "bucket.delete"
}

type BucketDeleteWorker struct {
	database    *database.Queries
	transaction *database.Transaction
	storage     *storage.S3Storage
	logger      *zap.Logger
	river.WorkerDefaults[BucketDeleteArgs]
}

func (w *BucketDeleteWorker) Work(ctx context.Context, bucketDeleteArgs *river.Job[BucketDeleteArgs]) error {
	const op = "BucketDeleteWorker.Work"

	err := w.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		limit := int32(100)
		offset := int32(0)

		for {
			objects, err := w.database.WithTx(tx).ListObjectsByBucketIdPaged(ctx, &database.ListObjectsByBucketIdPagedParams{
				BucketID: bucketDeleteArgs.Args.Id,
				Offset:   offset,
				Limit:    limit,
			})
			if err != nil {
				if database.IsNotFoundError(err) {
					break
				}
				w.logger.Error(
					"failed to list objects from database",
					zap.String("bucket", bucketDeleteArgs.Args.Name),
					zapfield.Operation(op),
					zap.Error(err),
				)
				return err
			}

			for _, object := range objects {
				err = w.storage.DeleteObject(ctx, &storage.ObjectDelete{
					Bucket: bucketDeleteArgs.Args.Name,
					Name:   object.Name,
				})
				if err != nil {
					w.logger.Error(
						"failed to delete object from storage",
						zap.String("bucket", bucketDeleteArgs.Args.Name),
						zap.String("name", object.Name),
						zapfield.Operation(op),
						zap.Error(err),
					)
					return err
				}
				err = w.database.WithTx(tx).DeleteObject(ctx, object.ID)
				if err != nil {
					w.logger.Error(
						"failed to delete object from database",
						zap.String("bucket", bucketDeleteArgs.Args.Name),
						zap.String("name", object.Name),
						zapfield.Operation(op),
						zap.Error(err),
					)
					return err
				}
			}

			offset += limit
		}

		err := w.database.WithTx(tx).DeleteBucket(ctx, bucketDeleteArgs.Args.Id)
		if err != nil {
			w.logger.Error(
				"failed to delete bucket from database",
				zap.String("bucket", bucketDeleteArgs.Args.Name),
				zapfield.Operation(op),
				zap.Error(err),
			)
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func NewBucketDeleteJob(database *database.Queries, transaction *database.Transaction, storage *storage.S3Storage, logger *zap.Logger) *BucketDeleteWorker {
	return &BucketDeleteWorker{
		database:    database,
		transaction: transaction,
		storage:     storage,
		logger:      logger,
	}
}
