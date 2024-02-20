package jobs

import (
	"context"
	"github.com/ArkamFahry/storage/server/database"
	"github.com/ArkamFahry/storage/server/models"
	"github.com/ArkamFahry/storage/server/storage"
	"github.com/ArkamFahry/storage/server/zapfield"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type PreSignedUploadSessionCompletion struct {
	ObjectId string `json:"object_id"`
}

func (PreSignedUploadSessionCompletion) Kind() string {
	return "pre.signed.upload.session.completion"
}

type PreSignedUploadSessionCompletionWorker struct {
	queries *database.Queries
	storage *storage.S3Storage
	logger  *zap.Logger
	river.WorkerDefaults[PreSignedUploadSessionCompletion]
}

func (w *PreSignedUploadSessionCompletionWorker) Work(ctx context.Context, preSignedUploadSessionCompletion *river.Job[PreSignedUploadSessionCompletion]) error {
	const op = "PreSignedUploadSessionCompletionWorker.Work"

	object, err := w.queries.ObjectGetByIdWithBucketName(ctx, preSignedUploadSessionCompletion.Args.ObjectId)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil
		}
		w.logger.Error(
			"failed to get object",
			zap.Error(err),
			zapfield.Operation(op),
			zap.String("object_id", preSignedUploadSessionCompletion.Args.ObjectId),
		)
		return err
	}
	objectExists, err := w.storage.CheckIfObjectExists(ctx, &storage.ObjectExistsCheck{
		Bucket: object.BucketName,
		Name:   object.Name,
	})
	if err != nil {
		w.logger.Error(
			"failed to check if object exists",
			zap.Error(err),
			zapfield.Operation(op),
			zap.String("bucket_name", object.BucketName),
			zap.String("object_name", object.Name),
		)
		return err
	}

	if objectExists {
		if object.UploadStatus != models.ObjectUploadStatusCompleted {
			err = w.queries.ObjectUpdateUploadStatus(ctx, &database.ObjectUpdateUploadStatusParams{
				ID:           object.ID,
				UploadStatus: models.ObjectUploadStatusCompleted,
			})
			if err != nil {
				w.logger.Error(
					"failed to update object upload status to completed",
					zap.Error(err),
					zapfield.Operation(op),
					zap.String("object_id", object.ID),
				)
				return err
			}
		}
	} else {
		err = w.storage.DeleteObject(ctx, &storage.ObjectDelete{
			Bucket: object.BucketName,
			Name:   object.Name,
		})
		if err != nil {
			w.logger.Error(
				"failed to delete object from storage",
				zap.Error(err),
				zapfield.Operation(op),
				zap.String("bucket_name", object.BucketName),
				zap.String("object_name", object.Name),
			)
			return err
		}
		err = w.queries.ObjectDelete(ctx, object.ID)
		if err != nil {
			w.logger.Error(
				"failed to delete object from database",
				zap.Error(err),
				zapfield.Operation(op),
				zap.String("object_id", object.ID),
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
