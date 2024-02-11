package jobs

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/storage"
	"github.com/ArkamFahry/hyperdrift/storage/server/zapfield"
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
			zap.String("bucket", preSignedObjectUploadCompletion.Args.BucketName),
			zapfield.Operation(op),
			zap.Error(err),
		)
		return err
	}

	err = w.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		if objectExists {
			err = w.queries.WithTx(tx).UpdateObjectUploadStatus(ctx, &database.UpdateObjectUploadStatusParams{
				ID:           preSignedObjectUploadCompletion.Args.ObjectId,
				UploadStatus: dto.ObjectUploadStatusCompleted,
			})
			if err != nil {
				w.logger.Error(
					"failed to update object upload status to completed",
					zap.String("bucket", preSignedObjectUploadCompletion.Args.BucketName),
					zapfield.Operation(op),
					zap.Error(err),
				)
				return err
			}
		} else {
			err = w.queries.WithTx(tx).DeleteObject(ctx, preSignedObjectUploadCompletion.Args.ObjectId)
			if err != nil {
				w.logger.Error(
					"failed to delete object",
					zap.String("bucket", preSignedObjectUploadCompletion.Args.BucketName),
					zapfield.Operation(op),
					zap.Error(err),
				)
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func NewPreSignedObjectUploadCompletionWorker(db *pgxpool.Pool, storage *storage.S3Storage, logger *zap.Logger) *PreSignedObjectUploadCompletionWorker {
	return &PreSignedObjectUploadCompletionWorker{
		queries:     database.New(db),
		transaction: database.NewTransaction(db),
		storage:     storage,
		logger:      logger,
	}
}
