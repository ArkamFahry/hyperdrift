package services

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ArkamFahry/storage/server/utils"
	"github.com/samber/lo"
	"github.com/zhooravell/mime"
	"strings"
	"time"

	"github.com/ArkamFahry/storage/server/config"
	"github.com/ArkamFahry/storage/server/database"
	"github.com/ArkamFahry/storage/server/jobs"
	"github.com/ArkamFahry/storage/server/models"
	"github.com/ArkamFahry/storage/server/srverr"
	"github.com/ArkamFahry/storage/server/storage"
	"github.com/ArkamFahry/storage/server/validators"
	"github.com/ArkamFahry/storage/server/zapfield"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

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

func (os *ObjectService) CreatePreSignedUploadSession(ctx context.Context, bucketName string, preSignedUploadSessionCreate *models.PreSignedUploadSessionCreate) (*models.PreSignedUploadSession, error) {
	const op = "ObjectService.CreatePreSignedSession"
	reqId := utils.RequestId(ctx)

	var preSignedObject *storage.PreSignedObject
	var id string

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_name cannot be empty. bucket_name is required to create pre-signed upload url", op, reqId, nil)
	}

	err := os.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := os.getBucketByNameTxn(ctx, tx, bucketName, op)
		if err != nil {
			return err
		}

		if preSignedUploadSessionCreate.ExpiresIn == nil {
			preSignedUploadSessionCreate.ExpiresIn = &os.config.DefaultPreSignedUploadUrlExpiresIn
		} else {
			err = validateExpiration(*preSignedUploadSessionCreate.ExpiresIn)
			if err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
			}
		}

		if lo.Contains[string](bucket.AllowedContentTypes, models.BucketAllowedWildcardContentTypes) {
			defaultContentType := models.ObjectDefaultObjectContentType

			if preSignedUploadSessionCreate.ContentType == nil || (preSignedUploadSessionCreate.ContentType != nil && *preSignedUploadSessionCreate.ContentType == "") {
				fileNameParts := strings.Split(preSignedUploadSessionCreate.Name, ".")
				if len(fileNameParts) == 2 {
					contentType, err := mime.GetMimeTypes(fileNameParts[1])
					if err != nil {
						preSignedUploadSessionCreate.ContentType = &defaultContentType
					} else {
						preSignedUploadSessionCreate.ContentType = &contentType[0]
					}
				} else {
					preSignedUploadSessionCreate.ContentType = &defaultContentType
				}

			} else {
				err = validateContentType(*preSignedUploadSessionCreate.ContentType)
				if err != nil {
					return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
				}
			}
		} else {
			if preSignedUploadSessionCreate.ContentType == nil {
				return srverr.NewServiceError(srverr.InvalidInputError, fmt.Sprintf("content_type cannot be empty. bucket only allows [%s] content types. please specify a allowed content type", strings.Join(bucket.AllowedContentTypes, ", ")), op, reqId, nil)
			} else {
				err = validateContentType(*preSignedUploadSessionCreate.ContentType)
				if err != nil {
					return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
				}
				if !lo.Contains[string](bucket.AllowedContentTypes, *preSignedUploadSessionCreate.ContentType) {
					return srverr.NewServiceError(srverr.InvalidInputError, fmt.Sprintf("content_type '%s' is not allowed. bucket only allows [%s] content types. please specify a allowed content type", *preSignedUploadSessionCreate.ContentType, strings.Join(bucket.AllowedContentTypes, ", ")), op, reqId, nil)
				}
			}
		}

		err = validateContentSize(preSignedUploadSessionCreate.Size)
		if err != nil {
			return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
		}

		if bucket.MaxAllowedObjectSize != nil {
			if preSignedUploadSessionCreate.Size > *bucket.MaxAllowedObjectSize {
				return srverr.NewServiceError(srverr.InvalidInputError, fmt.Sprintf("object size is too large. max allowed object size is %d bytes", *bucket.MaxAllowedObjectSize), op, reqId, nil)
			}
		}

		preSignedObject, err = os.storage.CreatePreSignedUploadObject(ctx, &storage.PreSignedUploadObjectCreate{
			Bucket:      bucket.Name,
			Name:        preSignedUploadSessionCreate.Name,
			ExpiresIn:   preSignedUploadSessionCreate.ExpiresIn,
			ContentType: *preSignedUploadSessionCreate.ContentType,
			Size:        preSignedUploadSessionCreate.Size,
		})
		if err != nil {
			os.logger.Error("failed to create pre-signed upload object", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to create pre-signed upload object", op, reqId, err)
		}

		metadataBytes, err := metadataToBytes(preSignedUploadSessionCreate.Metadata)
		if err != nil {
			os.logger.Error("failed to convert metadata to bytes", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to convert metadata to bytes", op, reqId, err)
		}

		id, err = os.queries.WithTx(tx).CreateObject(ctx, &database.CreateObjectParams{
			BucketID:     bucket.Id,
			Name:         preSignedUploadSessionCreate.Name,
			ContentType:  preSignedUploadSessionCreate.ContentType,
			Size:         preSignedUploadSessionCreate.Size,
			Metadata:     metadataBytes,
			UploadStatus: models.ObjectUploadStatusPending,
		})
		if err != nil {
			if database.IsConflictError(err) {
				return srverr.NewServiceError(srverr.ConflictError, fmt.Sprintf("object with name '%s' already exists", preSignedUploadSessionCreate.Name), op, reqId, err)
			}
			os.logger.Error("failed to create object in database", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to create object in database", op, reqId, err)
		}

		_, err = os.job.InsertTx(ctx, tx, jobs.PreSignedUploadSessionCompletion{
			BucketName: bucket.Name,
			ObjectName: preSignedUploadSessionCreate.Name,
			ObjectId:   id,
		}, &river.InsertOpts{
			ScheduledAt: time.Unix(preSignedObject.ExpiresAt, 0),
		})
		if err != nil {
			os.logger.Error("failed to create pre-signed object upload completion job", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to create pre-signed object upload completion job", op, reqId, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &models.PreSignedUploadSession{
		Id:        id,
		Url:       preSignedObject.Url,
		Method:    preSignedObject.Method,
		ExpiresAt: preSignedObject.ExpiresAt,
	}, nil
}

func (os *ObjectService) CompletePreSignedUploadSession(ctx context.Context, bucketName string, objectId string) error {
	const op = "ObjectService.CompletePreSignedUploadSession"
	reqId := utils.RequestId(ctx)

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket_name cannot be empty. bucket_name is required to complete pre-signed upload session", op, reqId, nil)
	}

	if validators.ValidateNotEmptyTrimmedString(objectId) {
		return srverr.NewServiceError(srverr.InvalidInputError, "object_id cannot be empty. object object_id is required to complete pre-signed upload session", op, reqId, nil)
	}

	bucket, err := os.getBucketByName(ctx, bucketName, op)
	if err != nil {
		return err
	}

	object, err := os.queries.GetObjectById(ctx, objectId)
	if err != nil {
		if database.IsNotFoundError(err) {
			return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", objectId), op, reqId, err)
		}
		os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, reqId, err)
	}

	switch object.UploadStatus {
	case models.ObjectUploadStatusCompleted:
		return srverr.NewServiceError(srverr.BadRequestError, fmt.Sprintf("upload session has already been completed for object '%s'", objectId), op, reqId, nil)
	case models.ObjectUploadStatusFailed:
		return srverr.NewServiceError(srverr.BadRequestError, fmt.Sprintf("upload session has failed for object '%s'", objectId), op, reqId, nil)
	}

	objectExists, err := os.storage.CheckIfObjectExists(ctx, &storage.ObjectExistsCheck{
		Bucket: bucket.Name,
		Name:   object.Name,
	})
	if err != nil {
		os.logger.Error("failed to check if object exists in storage", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return srverr.NewServiceError(srverr.UnknownError, "failed to check if object exists in storage", op, reqId, err)
	}

	if objectExists {
		err = os.queries.UpdateObjectUploadStatus(ctx, &database.UpdateObjectUploadStatusParams{
			ID:           object.ID,
			UploadStatus: models.ObjectUploadStatusCompleted,
		})
		if err != nil {
			os.logger.Error("failed to update object upload status in database to completed", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to update object upload status to completed", op, reqId, err)
		}
	} else {
		return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' has not yet been uploaded to storage", objectId), op, reqId, nil)
	}

	return nil
}

func (os *ObjectService) CreatePreSignedDownloadSession(ctx context.Context, bucketName string, objectId string, expiresIn int64) (*models.PreSignedDownloadSession, error) {
	const op = "ObjectService.CreatePreSignedDownloadSession"
	reqId := utils.RequestId(ctx)

	var preSignedDownloadObject models.PreSignedDownloadSession

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_name cannot be empty. bucket_name is required to create pre-signed download session", op, reqId, nil)
	}

	if validators.ValidateNotEmptyTrimmedString(objectId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "object_id cannot be empty. object_id is required to create pre-signed download session", op, reqId, nil)
	}

	if expiresIn == 0 {
		expiresIn = os.config.DefaultPreSignedUploadUrlExpiresIn
	} else {
		err := validateExpiration(expiresIn)
		if err != nil {
			return nil, srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
		}
	}

	err := os.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := os.getBucketByNameTxn(ctx, tx, bucketName, op)
		if err != nil {
			return err
		}

		object, err := os.queries.WithTx(tx).GetObjectById(ctx, objectId)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", objectId), op, reqId, err)
			}
			os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, reqId, err)
		}

		if object.UploadStatus != models.ObjectUploadStatusCompleted {
			objectExists, err := os.storage.CheckIfObjectExists(ctx, &storage.ObjectExistsCheck{
				Bucket: bucket.Name,
				Name:   object.Name,
			})
			if err != nil {
				os.logger.Error("failed to check if object exists in storage", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
				return srverr.NewServiceError(srverr.UnknownError, "failed to check if object exists in storage", op, reqId, err)
			}

			if objectExists {
				err = os.queries.WithTx(tx).UpdateObjectUploadStatus(ctx, &database.UpdateObjectUploadStatusParams{
					ID:           object.ID,
					UploadStatus: models.ObjectUploadStatusCompleted,
				})
				if err != nil {
					os.logger.Error("failed to update object upload status in database to completed", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
					return srverr.NewServiceError(srverr.UnknownError, "failed to update object upload status in database to completed", op, reqId, err)
				}
			} else {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' upload has not been completed", object.ID), op, reqId, nil)
			}
		}

		preSignedObject, err := os.storage.CreatePreSignedDownloadObject(ctx, &storage.PreSignedDownloadObjectCreate{
			Bucket: bucket.Name,
			Name:   object.Name,
		})
		if err != nil {
			os.logger.Error("failed to create pre-signed download url", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to create pre-signed download url", op, reqId, err)
		}

		err = os.queries.WithTx(tx).UpdateObjectLastAccessedAt(ctx, object.ID)
		if err != nil {
			os.logger.Error("failed to update object last accessed at", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to update object last accessed at", op, reqId, err)
		}

		preSignedDownloadObject = models.PreSignedDownloadSession{
			Url:       preSignedObject.Url,
			Method:    preSignedObject.Method,
			ExpiresAt: preSignedObject.ExpiresAt,
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return &preSignedDownloadObject, nil
}

func (os *ObjectService) DeleteObject(ctx context.Context, bucketName string, objectId string) error {
	const op = "ObjectService.DeleteObject"
	reqId := utils.RequestId(ctx)

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket_name cannot be empty. bucket_name is required to delete object", op, reqId, nil)
	}

	if validators.ValidateNotEmptyTrimmedString(objectId) {
		return srverr.NewServiceError(srverr.InvalidInputError, "object_id cannot be empty. object_id is required to delete object", op, reqId, nil)
	}

	err := os.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := os.getBucketByNameTxn(ctx, tx, bucketName, op)
		if err != nil {
			return err
		}

		object, err := os.queries.WithTx(tx).GetObjectById(ctx, objectId)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", objectId), op, reqId, err)
			}
			os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, reqId, err)
		}

		if object.UploadStatus != models.ObjectUploadStatusCompleted {
			return srverr.NewServiceError(srverr.InvalidInputError, fmt.Sprintf("upload has not yet been completed for object '%s'. delete operation can only be performed on objects that have been uploaded", object.ID), op, reqId, nil)
		}

		err = os.storage.DeleteObject(ctx, &storage.ObjectDelete{
			Bucket: bucket.Name,
			Name:   object.Name,
		})
		if err != nil {
			os.logger.Error("failed to delete object from storage", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to delete object from storage", op, reqId, err)
		}

		err = os.queries.WithTx(tx).DeleteObject(ctx, object.ID)
		if err != nil {
			os.logger.Error("failed to delete object from database", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to delete object from database", op, reqId, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (os *ObjectService) GetObject(ctx context.Context, bucketName string, objectId string) (*models.Object, error) {
	const op = "ObjectService.GetObject"
	reqId := utils.RequestId(ctx)

	if validators.ValidateNotEmptyTrimmedString(objectId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "object id cannot be empty. object id is required to get object", op, reqId, nil)
	}

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_name cannot be empty. bucket_name is required to get object", op, reqId, nil)
	}

	bucket, err := os.getBucketByName(ctx, bucketName, op)
	if err != nil {
		return nil, err
	}

	object, err := os.queries.GetObjectByBucketIdAndName(ctx, &database.GetObjectByBucketIdAndNameParams{
		BucketID: bucket.Id,
		Name:     objectId,
	})
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", objectId), op, reqId, err)
		}
		os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, reqId, err)
	}

	metadataMap, err := bytesToMetadata(object.Metadata)
	if err != nil {
		os.logger.Error("failed to convert metadata from bytes", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to convert metadata from bytes", op, reqId, err)
	}

	return &models.Object{
		Id:           object.ID,
		BucketId:     object.BucketID,
		Name:         object.Name,
		ContentType:  object.ContentType,
		Size:         object.Size,
		Metadata:     metadataMap,
		UploadStatus: object.UploadStatus,
		CreatedAt:    object.CreatedAt,
		UpdatedAt:    object.UpdatedAt,
	}, nil
}

func (os *ObjectService) SearchObjects(ctx context.Context, bucketName string, objectPath string, level int32, limit int32, offset int32) ([]*models.Object, error) {
	const op = "ObjectService.SearchObjects"
	reqId := utils.RequestId(ctx)

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required to search objects", op, reqId, nil)
	}

	if validators.ValidateNotEmptyTrimmedString(objectPath) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "object name cannot be empty. object name is required to search objects", op, reqId, nil)
	}

	if level < 0 {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "levels cannot be less than 0", op, reqId, nil)
	}

	if limit < 0 {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "limit cannot be less than 0", op, reqId, nil)
	}

	if offset < 0 {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "offset cannot be less than 0", op, reqId, nil)
	}

	if level == 0 {
		level = 1
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
		os.logger.Error("failed to search objects", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to search objects", op, reqId, err)
	}
	if len(objects) == 0 {
		return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("no objects found for bucket '%s' with path '%s'", bucketName, objectPath), op, reqId, nil)
	}

	var result []*models.Object

	for _, object := range objects {
		metadataMap, _ := bytesToMetadata(object.Metadata)
		result = append(result, &models.Object{
			Id:           object.ID,
			Version:      object.Version,
			BucketId:     object.BucketID,
			Name:         object.Name,
			ContentType:  object.ContentType,
			Size:         object.Size,
			Metadata:     metadataMap,
			UploadStatus: object.UploadStatus,
			CreatedAt:    object.CreatedAt,
			UpdatedAt:    &object.UpdatedAt,
		})
	}

	return result, nil
}

func (os *ObjectService) getBucketByNameTxn(ctx context.Context, tx pgx.Tx, bucketName string, op string) (*models.Bucket, error) {
	reqId := utils.RequestId(ctx)

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required", op, reqId, nil)
	}

	if validateBucketName(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name is not valid. it must start and end with an alphanumeric character, and can include alphanumeric characters, hyphens, and dots. The total length must be between 3 and 63 characters", op, reqId, nil)
	}

	bucket, err := os.queries.WithTx(tx).GetBucketByName(ctx, bucketName)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket '%s' not found", bucketName), op, reqId, err)
		}
		os.logger.Error("failed to get bucket by name", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket by name", op, "", err)
	}

	if bucket.Disabled {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is disabled", bucket.Name), op, reqId, err)
	}

	if bucket.Locked {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is locked for '%s'", bucket.Name, *bucket.LockReason), op, reqId, err)
	}

	return &models.Bucket{
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

func (os *ObjectService) getBucketByName(ctx context.Context, bucketName string, op string) (*models.Bucket, error) {
	reqId := utils.RequestId(ctx)

	if validators.ValidateNotEmptyTrimmedString(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required", op, reqId, nil)
	}

	if validateBucketName(bucketName) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name is not valid. it must start and end with an alphanumeric character, and can include alphanumeric characters, hyphens, and dots. The total length must be between 3 and 63 characters", op, reqId, nil)
	}

	bucket, err := os.queries.GetBucketByName(ctx, bucketName)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket '%s' not found", bucketName), op, reqId, err)
		}
		os.logger.Error("failed to get bucket by name", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket by name", op, reqId, err)
	}

	if bucket.Disabled {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is disabled", bucket.Name), op, reqId, err)
	}

	if bucket.Locked {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is locked for '%s'", bucket.Name, *bucket.LockReason), op, reqId, err)
	}

	return &models.Bucket{
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
	if !validators.ValidateContentType(contentType) {
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
