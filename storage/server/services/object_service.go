package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/entities"
	"github.com/ArkamFahry/hyperdrift/storage/server/srverr"
	"github.com/ArkamFahry/hyperdrift/storage/server/storage"
	"github.com/ArkamFahry/hyperdrift/storage/server/validators"
	"github.com/ArkamFahry/hyperdrift/storage/server/zapfield"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

type ObjectService struct {
	query       *database.Queries
	transaction *database.Transaction
	storage     *storage.S3Storage
	job         *river.Client[pgx.Tx]
	config      *config.Config
	logger      *zap.Logger
}

func NewObjectService(db *pgxpool.Pool, job *river.Client[pgx.Tx], config *config.Config, logger *zap.Logger) *ObjectService {
	return &ObjectService{
		query:       database.New(db),
		transaction: database.NewTransaction(db),
		job:         job,
		config:      config,
		logger:      logger,
	}
}

func (os *ObjectService) CreatePreSignedUploadObject(ctx context.Context, preSignedUploadObjectCreate *dto.PreSignedUploadObjectCreate) (*dto.PreSignedObject, error) {
	const op = "ObjectService.CreatePreSignedUploadUrl"

	var presignedUploadUrl *dto.PreSignedObject

	err := os.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := os.getBucketByName(ctx, tx, preSignedUploadObjectCreate.Bucket, op)
		if err != nil {
			return err
		}

		if preSignedUploadObjectCreate.ExpiresIn == nil {
			preSignedUploadObjectCreate.ExpiresIn = &os.config.DefaultPreSignedUploadUrlExpiresIn
		} else {
			err = validateExpiration(*preSignedUploadObjectCreate.ExpiresIn)
			if err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
			}
		}

		err = validateContentType(preSignedUploadObjectCreate.ContentType)
		if err != nil {
			return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
		}

		err = validateContentSize(preSignedUploadObjectCreate.Size)
		if err != nil {
			return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
		}

		presignedUploadUrl, err = os.storage.CreatePreSignedUploadObject(ctx, &storage.PreSignedUploadObjectCreate{
			Bucket:      bucket.Name,
			Name:        preSignedUploadObjectCreate.Name,
			ExpiresIn:   preSignedUploadObjectCreate.ExpiresIn,
			ContentType: preSignedUploadObjectCreate.ContentType,
			Size:        preSignedUploadObjectCreate.Size,
		})
		if err != nil {
			return srverr.NewServiceError(srverr.UnknownError, "failed to create pre-signed upload object", op, "", err)
		}

		metadataBytes, err := metadataToBytes(preSignedUploadObjectCreate.Metadata)
		if err != nil {
			return srverr.NewServiceError(srverr.UnknownError, "failed to convert metadata to bytes", op, "", err)
		}

		err = os.query.WithTx(tx).CreateObject(ctx, &database.CreateObjectParams{
			ID:          newObjectId(),
			BucketID:    bucket.Id,
			Name:        preSignedUploadObjectCreate.Name,
			ContentType: preSignedUploadObjectCreate.ContentType,
			Size:        preSignedUploadObjectCreate.Size,
			Public:      preSignedUploadObjectCreate.Public,
			Metadata:    metadataBytes,
		})
		if database.IsConflictError(err) {
			return srverr.NewServiceError(srverr.ConflictError, fmt.Sprintf("object with name '%s' already exists", preSignedUploadObjectCreate.Name), op, "", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.PreSignedObject{
		Url:       presignedUploadUrl.Url,
		Method:    presignedUploadUrl.Method,
		ExpiresAt: presignedUploadUrl.ExpiresAt,
	}, nil
}

func (os *ObjectService) getBucketByName(ctx context.Context, tx pgx.Tx, bucketName string, op string) (*entities.Bucket, error) {
	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required", op, "", nil)
	}

	if validateBucketName(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name is not valid. it must start and end with an alphanumeric character, and can include alphanumeric characters, hyphens, and dots. The total length must be between 3 and 63 characters", op, "", nil)
	}

	bucket, err := os.query.WithTx(tx).GetBucketByName(ctx, bucketName)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket '%s' not found", bucketName), op, "", err)
		}
		os.logger.Error("failed to get bucket by name", zap.Error(err), zapfield.Operation(op), zap.String("bucket", bucketName))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket by name", op, "", err)
	}

	if bucket.Disabled {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is disabled", bucket.Name), op, "", err)
	}

	if bucket.Locked {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is locked for '%s'", bucket.Name, *bucket.LockReason), op, "", err)
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

func validateExpiration(expiresIn int64) error {
	if expiresIn <= 0 {
		return fmt.Errorf("expires in must be greater than 0")
	}

	return nil
}

func validateContentType(contentType string) error {
	if validators.ValidateContentType(contentType) {
		return fmt.Errorf("invalid content type '%s'", contentType)
	}

	return nil
}

func validateContentSize(size int64) error {
	if size <= 0 {
		return fmt.Errorf("content size must be greater than 0")
	}

	return nil
}

func metadataToBytes(metadata map[string]any) ([]byte, error) {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata to bytes: %w", err)
	}
	return metadataBytes, nil
}

func newObjectId() string {
	return fmt.Sprintf("objects_%s", uuid.New().String())
}
