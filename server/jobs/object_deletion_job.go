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

type ObjectDeletion struct {
	ObjectId string `json:"object_id"`
}

func (ObjectDeletion) Kind() string {
	return "object.deletion"
}

type ObjectDeletionWorker struct {
	queries *database.Queries
	storage *storage.Storage
	logger  *zap.Logger
	river.WorkerDefaults[ObjectDeletion]
}

func (w *ObjectDeletionWorker) Work(ctx context.Context, objectDeletion *river.Job[ObjectDeletion]) error {
	const op = "ObjectDeletionWorker.Work"

	object, err := w.queries.ObjectGetByIdWithBucketName(ctx, objectDeletion.Args.ObjectId)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil
		}
		w.logger.Error(
			"failed to get object",
			zap.Error(err),
			zapfield.Operation(op),
			zap.String("object_id", objectDeletion.Args.ObjectId),
		)
		return err
	}

	err = w.storage.DeleteObject(ctx, &storage.ObjectDelete{
		Bucket: object.BucketName,
		Name:   object.Name,
	})
	if err != nil {
		w.logger.Error(
			"failed to delete object",
			zap.Error(err),
			zapfield.Operation(op),
			zap.String("bucket_name", object.BucketName),
			zap.String("object_name", object.Name),
		)
		return err
	}

	err = w.queries.ObjectDelete(ctx, objectDeletion.Args.ObjectId)
	if err != nil {
		w.logger.Error(
			"failed to delete object",
			zap.Error(err),
			zapfield.Operation(op),
			zap.String("object_id", objectDeletion.Args.ObjectId),
		)
		return err
	}

	return nil
}

func NewObjectDeletionWorker(db *pgxpool.Pool, storage *storage.Storage, logger *zap.Logger) *ObjectDeletionWorker {
	return &ObjectDeletionWorker{
		queries: database.New(db),
		storage: storage,
		logger:  logger,
	}
}
