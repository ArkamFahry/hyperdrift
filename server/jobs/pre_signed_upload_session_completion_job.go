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

type PreSignedUploadSessionCompletion struct {
	BucketName string `json:"bucket_name"`
	ObjectId   string `json:"object_id"`
	ObjectName string `json:"object_name"`
}

func (PreSignedUploadSessionCompletion) Kind() string {
	return "pre.signed.upload.session.completion"
}

type PreSignedUploadSessionCompletionWorker struct {
	queries     *database.Queries
	transaction *database.Transaction
	storage     *storage.S3Storage
	logger      *zap.Logger
	river.WorkerDefaults[PreSignedUploadSessionCompletion]
}

func (w *PreSignedUploadSessionCompletionWorker) Work(ctx context.Context, preSignedUploadSessionCompletion *river.Job[PreSignedUploadSessionCompletion]) error {
	const op = "PreSignedUploadSessionCompletionWorker.Work"

	objectExists, err := w.storage.CheckIfObjectExists(ctx, &storage.ObjectExistsCheck{
		Bucket: preSignedUploadSessionCompletion.Args.BucketName,
		Name:   preSignedUploadSessionCompletion.Args.ObjectName,
	})
	if err != nil {
		w.logger.Error(
			"failed to check if object exists",
			zap.Error(err),
			zapfield.Operation(op),
			zap.String("bucket_name", preSignedUploadSessionCompletion.Args.BucketName),
			zap.String("object_name", preSignedUploadSessionCompletion.Args.ObjectName),
		)
		return err
	}

	if objectExists {
		err = w.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
			object, err := w.queries.WithTx(tx).GetObjectById(ctx, preSignedUploadSessionCompletion.Args.ObjectId)
			if err != nil {
				w.logger.Error(
					"failed to get object",
					zap.Error(err),
					zapfield.Operation(op),
					zap.String("object_id", preSignedUploadSessionCompletion.Args.ObjectId),
				)
			}

			if object.UploadStatus == models.ObjectUploadStatusCompleted {
				return nil
			}

			err = w.queries.WithTx(tx).UpdateObjectUploadStatus(ctx, &database.UpdateObjectUploadStatusParams{
				ID:           preSignedUploadSessionCompletion.Args.ObjectId,
				UploadStatus: models.ObjectUploadStatusCompleted,
			})
			if err != nil {
				w.logger.Error(
					"failed to update object upload status to completed",
					zap.Error(err),
					zapfield.Operation(op),
					zap.String("object_id", preSignedUploadSessionCompletion.Args.ObjectId),
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
			ID:           preSignedUploadSessionCompletion.Args.ObjectId,
			UploadStatus: models.ObjectUploadStatusFailed,
		})
		if err != nil {
			w.logger.Error(
				"failed to update object upload status to failed",
				zap.Error(err),
				zapfield.Operation(op),
				zap.String("object_id", preSignedUploadSessionCompletion.Args.ObjectId),
			)
			return err
		}
	}

	return nil
}

func NewPreSignedUploadSessionCompletionWorker(db *pgxpool.Pool, storage *storage.S3Storage, logger *zap.Logger) *PreSignedUploadSessionCompletionWorker {
	return &PreSignedUploadSessionCompletionWorker{
		queries: database.New(db),
		storage: storage,
		logger:  logger,
	}
}
