package jobs

import (
	"context"
	"github.com/ArkamFahry/storage/server/database"
	"github.com/ArkamFahry/storage/server/models"
	"github.com/ArkamFahry/storage/server/storage"
	"github.com/ArkamFahry/storage/server/zapfield"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type PreSignedObjectUploadCompletion struct {
	BucketName string `json:"bucket_name"`
	ObjectId   string `json:"object_id"`
	ObjectName string `json:"object_name"`
}

func (PreSignedObjectUploadCompletion) Kind() string {
	return "pre.signed.object.upload.completion"
}

type PreSignedObjectUploadCompletionWorker struct {
	queries     *database.Queries
	transaction *database.Transaction
	storage     *storage.S3Storage
	logger      *zap.Logger
	river.WorkerDefaults[PreSignedObjectUploadCompletion]
}

func (w *PreSignedObjectUploadCompletionWorker) Work(ctx context.Context, preSignedObjectUploadCompletion *river.Job[PreSignedObjectUploadCompletion]) error {
	const op = "PreSignedObjectUploadCompletionWorker.Work"

	objectExists, err := w.storage.CheckIfObjectExists(ctx, &storage.ObjectExistsCheck{
		Bucket: preSignedObjectUploadCompletion.Args.BucketName,
		Name:   preSignedObjectUploadCompletion.Args.ObjectName,
	})
	if err != nil {
		w.logger.Error(
			"failed to check if object exists",
			zap.String("bucket_name", preSignedObjectUploadCompletion.Args.BucketName),
			zap.String("object_name", preSignedObjectUploadCompletion.Args.ObjectName),
			zapfield.Operation(op),
			zap.Error(err),
		)
		return err
	}

	if objectExists {
		err = w.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
			object, err := w.queries.WithTx(tx).GetObjectById(ctx, preSignedObjectUploadCompletion.Args.ObjectId)
			if err != nil {
				w.logger.Error(
					"failed to get object",
					zap.String("object_id", preSignedObjectUploadCompletion.Args.ObjectId),
					zapfield.Operation(op),
				)
			}

			if object.UploadStatus == models.ObjectUploadStatusCompleted {
				return nil
			}

			err = w.queries.WithTx(tx).UpdateObjectUploadStatus(ctx, &database.UpdateObjectUploadStatusParams{
				ID:           preSignedObjectUploadCompletion.Args.ObjectId,
				UploadStatus: models.ObjectUploadStatusCompleted,
			})
			if err != nil {
				w.logger.Error(
					"failed to update object upload status to completed",
					zap.String("object_id", preSignedObjectUploadCompletion.Args.ObjectId),
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
	} else {
		err = w.queries.UpdateObjectUploadStatus(ctx, &database.UpdateObjectUploadStatusParams{
			ID:           preSignedObjectUploadCompletion.Args.ObjectId,
			UploadStatus: models.ObjectUploadStatusFailed,
		})
		if err != nil {
			w.logger.Error(
				"failed to update object upload status to failed",
				zap.String("object_id", preSignedObjectUploadCompletion.Args.ObjectId),
				zapfield.Operation(op),
				zap.Error(err),
			)
			return err
		}
	}

	return nil
}

func NewPreSignedObjectUploadCompletionWorker(db *pgxpool.Pool, storage *storage.S3Storage, logger *zap.Logger) *PreSignedObjectUploadCompletionWorker {
	return &PreSignedObjectUploadCompletionWorker{
		queries: database.New(db),
		storage: storage,
		logger:  logger,
	}
}
