package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ArkamFahry/hyperdrift/storage/server/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/entities"
	"github.com/ArkamFahry/hyperdrift/storage/server/jobs"
	"github.com/ArkamFahry/hyperdrift/storage/server/srverr"
	"github.com/ArkamFahry/hyperdrift/storage/server/storage"
	"github.com/ArkamFahry/hyperdrift/storage/server/validators"
	"github.com/ArkamFahry/hyperdrift/storage/server/zapfield"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

const DefaultContentType = "application/octet-stream"

type ObjectService struct {
	queries     *database.Queries
	transaction *database.Transaction
	storage     *storage.S3Storage
	job         *river.Client[pgx.Tx]
	config      *config.Config
	logger      *zap.Logger
}

func NewObjectService(db *pgxpool.Pool, storage *storage.S3Storage, job *river.Client[pgx.Tx], config *config.Config, logger *zap.Logger) *ObjectService {
	return &ObjectService{
		queries:     database.New(db),
		transaction: database.NewTransaction(db),
		storage:     storage,
		job:         job,
		config:      config,
		logger:      logger,
	}
}

func (os *ObjectService) CreatePreSignedUploadObject(ctx context.Context, preSignedUploadObjectCreate *dto.PreSignedUploadObjectCreate) (*dto.PreSignedUploadObject, error) {
	const op = "ObjectService.CreatePreSignedUploadUrl"

	var preSignedObject *dto.PreSignedUploadObject
	var id string

	err := os.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := os.getBucketByNameTxn(ctx, tx, preSignedUploadObjectCreate.Bucket, op)
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

		if preSignedUploadObjectCreate.ContentType == nil {
			contentType := DefaultContentType
			preSignedUploadObjectCreate.ContentType = &contentType
		} else {
			err = validateContentType(*preSignedUploadObjectCreate.ContentType)
			if err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
			}
		}

		err = validateContentSize(preSignedUploadObjectCreate.Size)
		if err != nil {
			return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
		}

		preSignedObject, err = os.storage.CreatePreSignedUploadObject(ctx, &storage.PreSignedUploadObjectCreate{
			Bucket:      bucket.Name,
			Name:        preSignedUploadObjectCreate.Name,
			ExpiresIn:   preSignedUploadObjectCreate.ExpiresIn,
			ContentType: *preSignedUploadObjectCreate.ContentType,
			Size:        preSignedUploadObjectCreate.Size,
		})
		if err != nil {
			return srverr.NewServiceError(srverr.UnknownError, "failed to create pre-signed upload object", op, "", err)
		}

		metadataBytes, err := metadataToBytes(preSignedUploadObjectCreate.Metadata)
		if err != nil {
			os.logger.Error("failed to convert metadata to bytes", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to convert metadata to bytes", op, "", err)
		}

		id, err = os.queries.WithTx(tx).CreateObject(ctx, &database.CreateObjectParams{
			BucketID:     bucket.Id,
			Name:         preSignedUploadObjectCreate.Name,
			ContentType:  preSignedUploadObjectCreate.ContentType,
			Size:         preSignedUploadObjectCreate.Size,
			Public:       preSignedUploadObjectCreate.Public,
			Metadata:     metadataBytes,
			UploadStatus: dto.ObjectUploadStatusPending,
		})
		if err != nil {
			if database.IsConflictError(err) {
				return srverr.NewServiceError(srverr.ConflictError, fmt.Sprintf("object with name '%s' already exists", preSignedUploadObjectCreate.Name), op, "", err)
			}
			os.logger.Error("failed to create object in database", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to create object in database", op, "", err)
		}

		_, err = os.job.InsertTx(ctx, tx, jobs.PreSignedObjectUploadCompletion{
			BucketName: bucket.Name,
			ObjectName: preSignedUploadObjectCreate.Name,
			ObjectId:   id,
		}, &river.InsertOpts{
			ScheduledAt: time.Unix(preSignedObject.ExpiresAt, 0),
		})
		if err != nil {
			os.logger.Error("failed to insert pre-signed object upload completion job", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to create pre-signed object upload completion job", op, "", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.PreSignedUploadObject{
		Id:        id,
		Url:       preSignedObject.Url,
		Method:    preSignedObject.Method,
		ExpiresAt: preSignedObject.ExpiresAt,
	}, nil
}

func (os *ObjectService) CompletePreSignedObjectUpload(ctx context.Context, id string) error {
	const op = "ObjectService.CompletePreSignedObjectUpload"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return srverr.NewServiceError(srverr.InvalidInputError, "object id cannot be empty. object id is required to complete pre-signed upload", op, "", nil)
	}

	object, err := os.queries.GetObjectByIdWithBucketName(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", id), op, "", err)
		}
		os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op))
		return srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, "", err)
	}

	switch object.UploadStatus {
	case dto.ObjectUploadStatusCompleted:
		return srverr.NewServiceError(srverr.InvalidInputError, fmt.Sprintf("upload has already been completed for object '%s'", object.Name), op, "", nil)
	case dto.ObjectUploadStatusFailed:
		return srverr.NewServiceError(srverr.InvalidInputError, fmt.Sprintf("upload has failed for object '%s'", object.Name), op, "", nil)
	}

	objectExists, err := os.storage.CheckIfObjectExists(ctx, &storage.ObjectExistsCheck{
		Bucket: object.BucketName,
		Name:   object.Name,
	})
	if err != nil {
		os.logger.Error("failed to check if object exists in storage", zap.Error(err), zapfield.Operation(op))
		return srverr.NewServiceError(srverr.UnknownError, "failed to check if object exists in storage", op, "", err)
	}

	if objectExists {
		err = os.queries.UpdateObjectUploadStatus(ctx, &database.UpdateObjectUploadStatusParams{
			ID:           object.ID,
			UploadStatus: dto.ObjectUploadStatusCompleted,
		})
		if err != nil {
			os.logger.Error("failed to update object upload status in database to completed", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to update object upload status in database to completed", op, "", err)
		}
	} else {
		return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' has not yet been uploaded to storage", object.Name), op, "", nil)
	}

	return nil
}

func (os *ObjectService) GetObjectById(ctx context.Context, id string) (*entities.Object, error) {
	const op = "ObjectService.GetObjectById"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "object id cannot be empty. object id is required to get object", op, "", nil)
	}

	object, err := os.queries.GetObjectById(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", id), op, "", err)
		}
		os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, "", err)
	}

	metadataMap, err := bytesToMetadata(object.Metadata)
	if err != nil {
		os.logger.Error("failed to convert metadata from bytes", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to convert metadata from bytes", op, "", err)
	}

	return &entities.Object{
		Id:           object.ID,
		BucketId:     object.BucketID,
		Name:         object.Name,
		ContentType:  object.ContentType,
		Size:         object.Size,
		Public:       object.Public,
		Metadata:     metadataMap,
		UploadStatus: object.UploadStatus,
		CreatedAt:    object.CreatedAt,
		UpdatedAt:    object.UpdatedAt,
	}, nil
}

func (os *ObjectService) SearchObjectsByBucketNameAndObjectPath(ctx context.Context, bucketName string, objectPath string, level int32, limit int32, offset int32) ([]*entities.Object, error) {
	const op = "ObjectService.SearchObjectsByBucketNameAndObjectPath"

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required to search objects", op, "", nil)
	}

	if validators.ValidateNotEmptyTrimmedString(objectPath) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "object name cannot be empty. object name is required to search objects", op, "", nil)
	}

	if level < 0 {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "levels cannot be less than 0", op, "", nil)
	}

	if limit < 0 {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "limit cannot be less than 0", op, "", nil)
	}

	if offset < 0 {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "offset cannot be less than 0", op, "", nil)
	}

	if limit == 0 {
		limit = 100
	}

	objects, err := os.queries.SearchObjectsByPath(ctx, &database.SearchObjectsByPathParams{
		BucketName: bucketName,
		ObjectPath: objectPath,
		Level:      &level,
		Limit:      &limit,
		Offset:     &offset,
	})
	if err != nil {
		os.logger.Error("failed to search objects from database", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to search objects from database", op, "", err)
	}
	if len(objects) == 0 {
		return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("no objects found for bucket '%s' with path '%s'", bucketName, objectPath), op, "", nil)
	}

	var result []*entities.Object

	for _, object := range objects {
		metadataMap, _ := bytesToMetadata(object.Metadata)
		result = append(result, &entities.Object{
			Id:           object.ID,
			Version:      object.Version,
			BucketId:     object.BucketID,
			Name:         object.Name,
			ContentType:  object.ContentType,
			Size:         object.Size,
			Public:       object.Public,
			Metadata:     metadataMap,
			UploadStatus: object.UploadStatus,
			CreatedAt:    object.CreatedAt,
			UpdatedAt:    &object.UpdatedAt,
		})
	}

	return result, nil
}

func (os *ObjectService) getBucketByNameTxn(ctx context.Context, tx pgx.Tx, bucketName string, op string) (*entities.Bucket, error) {
	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required", op, "", nil)
	}

	if validateBucketName(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name is not valid. it must start and end with an alphanumeric character, and can include alphanumeric characters, hyphens, and dots. The total length must be between 3 and 63 characters", op, "", nil)
	}

	bucket, err := os.queries.WithTx(tx).GetBucketByName(ctx, bucketName)
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

func bytesToMetadata(metadataBytes []byte) (map[string]any, error) {
	var metadata map[string]any
	err := json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata from bytes: %w", err)
	}
	return metadata, nil
}
