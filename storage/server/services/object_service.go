package services

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/entities"
	"github.com/ArkamFahry/hyperdrift/storage/server/srverr"
	"github.com/ArkamFahry/hyperdrift/storage/server/zapfield"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type ObjectService struct {
	query       *database.Queries
	transaction *database.Transaction
	logger      *zap.Logger
	job         *river.Client[pgx.Tx]
}

func NewObjectService(db *pgxpool.Pool, logger *zap.Logger, job *river.Client[pgx.Tx]) *ObjectService {
	return &ObjectService{
		query:       database.New(db),
		transaction: database.NewTransaction(db),
		logger:      logger,
		job:         job,
	}
}

func (os *ObjectService) CreatePreSignedUploadObject(ctx context.Context, objectCreate *dto.PreSignedUploadObjectCreate) (*dto.PreSignedObject, error) {
	const op = "ObjectService.CreatePreSignedUploadUrl"

	return nil, nil
}

func (os *ObjectService) getBucketByName(ctx context.Context, tx pgx.Tx, bucketName string, op string) (*entities.Bucket, error) {
	bucket, err := os.query.WithTx(tx).GetBucketByName(ctx, bucketName)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with name '%s' not found", bucketName), op, "", err)
		}
		os.logger.Error("failed to get bucket by name", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket by name", op, "", err)
	}

	if bucket.Disabled {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with name '%s' is disabled", bucket.Name), op, "", err)
	}

	if bucket.Locked {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with name '%s' is locked for '%s'", bucket.Name, *bucket.LockReason), op, "", err)
	}

	return &entities.Bucket{
		Id:                   bucket.ID,
		Version:              bucket.Version,
		Name:                 bucket.Name,
		AllowedContentTypes:  bucket.AllowedContentTypes,
		MaxAllowedObjectSize: bucket.MaxAllowedObjectSize,
		Public:               bucket.Public,
		Disabled:             bucket.Disabled,
		Locked:               bucket.Locked,
		LockReason:           bucket.LockReason,
		LockedAt:             bucket.LockedAt,
		CreatedAt:            bucket.CreatedAt,
		UpdatedAt:            bucket.UpdatedAt,
	}, nil
}
