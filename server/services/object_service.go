package services

import (
	"context"
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

func (os *ObjectService) CreatePreSignedUploadSession(ctx context.Context, bucketId string, preSignedUploadSessionCreate *models.PreSignedUploadSessionCreate) (*models.PreSignedUploadSession, error) {
	const op = "ObjectService.CreatePreSignedUploadSession"
	reqId := utils.RequestId(ctx)

	var preSignedObject *storage.PreSignedObject
	var id string

	if validateNotEmptyTrimmedString(bucketId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_id cannot be empty. bucket_id is required to create pre-signed upload session", op, reqId, nil)
	}

	if validateNotEmptyTrimmedString(preSignedUploadSessionCreate.Name) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "name cannot be empty. name is required to create pre-signed upload session", op, reqId, nil)
	}

	if err := validateObjectName(preSignedUploadSessionCreate.Name); err != nil {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, nil)
	}

	err := os.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := os.getBucketByIdTxn(ctx, tx, bucketId, op)
		if err != nil {
			return err
		}

		if preSignedUploadSessionCreate.ExpiresIn == nil {
			preSignedUploadSessionCreate.ExpiresIn = &os.config.DefaultPreSignedUploadUrlExpiry
		} else {
			if err = validateExpiration(*preSignedUploadSessionCreate.ExpiresIn); err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
			}
		}

		if lo.Contains[string](bucket.AllowedMimeTypes, models.BucketAllowedMimeTypesWildcard) {
			defaultMimeType := models.ObjectDefaultMimeType
			if preSignedUploadSessionCreate.MimeType == nil || (preSignedUploadSessionCreate.MimeType != nil && *preSignedUploadSessionCreate.MimeType == "") {
				objectNameParts := strings.Split(preSignedUploadSessionCreate.Name, ".")
				if len(objectNameParts) > 1 {
					objectExtension := objectNameParts[len(objectNameParts)-1]
					mimeType, err := mime.GetMimeTypes(objectExtension)
					if err != nil {
						preSignedUploadSessionCreate.MimeType = &defaultMimeType
					} else {
						preSignedUploadSessionCreate.MimeType = &mimeType[0]
					}
				} else {
					preSignedUploadSessionCreate.MimeType = &defaultMimeType
				}
			} else {
				if err = validateMimeType(*preSignedUploadSessionCreate.MimeType); err != nil {
					return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
				}
			}
		} else {
			if preSignedUploadSessionCreate.MimeType == nil {
				return srverr.NewServiceError(srverr.InvalidInputError, fmt.Sprintf("mime_type cannot be empty. bucket only allows [%s] mime types. please specify a allowed mime type", strings.Join(bucket.AllowedMimeTypes, ", ")), op, reqId, nil)
			} else {
				if err = validateMimeType(*preSignedUploadSessionCreate.MimeType); err != nil {
					return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
				}
				if !lo.Contains[string](bucket.AllowedMimeTypes, *preSignedUploadSessionCreate.MimeType) {
					return srverr.NewServiceError(srverr.InvalidInputError, fmt.Sprintf("mime_type '%s' is not allowed. bucket only allows [%s] mime types. please specify a allowed mime type", *preSignedUploadSessionCreate.MimeType, strings.Join(bucket.AllowedMimeTypes, ", ")), op, reqId, nil)
				}
			}
		}

		if err = validateObjectSize(preSignedUploadSessionCreate.Size); err != nil {
			return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
		}

		if bucket.MaxAllowedObjectSize != nil {
			if preSignedUploadSessionCreate.Size > *bucket.MaxAllowedObjectSize {
				return srverr.NewServiceError(srverr.BadRequestError, fmt.Sprintf("object size is too large. max allowed object size is %d bytes", *bucket.MaxAllowedObjectSize), op, reqId, nil)
			}
		}

		preSignedObject, err = os.storage.CreatePreSignedUploadObject(ctx, &storage.PreSignedUploadObjectCreate{
			Bucket:        bucket.Name,
			Name:          preSignedUploadSessionCreate.Name,
			ExpiresIn:     preSignedUploadSessionCreate.ExpiresIn,
			ContentType:   *preSignedUploadSessionCreate.MimeType,
			ContentLength: preSignedUploadSessionCreate.Size,
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

		id, err = os.queries.WithTx(tx).ObjectCreate(ctx, &database.ObjectCreateParams{
			BucketID:     bucket.Id,
			Name:         preSignedUploadSessionCreate.Name,
			ContentType:  preSignedUploadSessionCreate.MimeType,
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
			ObjectId: id,
		}, &river.InsertOpts{
			ScheduledAt: time.Unix(preSignedObject.ExpiresAt, 0).Add(time.Second * 10),
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

func (os *ObjectService) CompletePreSignedUploadSession(ctx context.Context, bucketId string, objectId string) error {
	const op = "ObjectService.CompletePreSignedUploadSession"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(bucketId) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket_id cannot be empty. bucket_id is required to complete pre-signed upload session", op, reqId, nil)
	}

	if validateNotEmptyTrimmedString(objectId) {
		return srverr.NewServiceError(srverr.InvalidInputError, "object_id cannot be empty. object object_id is required to complete pre-signed upload session", op, reqId, nil)
	}

	bucket, err := os.getBucketById(ctx, bucketId, op)
	if err != nil {
		return err
	}

	object, err := os.queries.ObjectGetById(ctx, objectId)
	if err != nil {
		if database.IsNotFoundError(err) {
			return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", objectId), op, reqId, err)
		}
		os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, reqId, err)
	}

	if object.UploadStatus == models.ObjectUploadStatusCompleted {
		return srverr.NewServiceError(srverr.BadRequestError, fmt.Sprintf("upload session has already been completed for object '%s'", objectId), op, reqId, nil)
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
		err = os.queries.ObjectUpdateUploadStatus(ctx, &database.ObjectUpdateUploadStatusParams{
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

func (os *ObjectService) CreatePreSignedDownloadSession(ctx context.Context, bucketId string, objectId string, expiresIn int64) (*models.PreSignedDownloadSession, error) {
	const op = "ObjectService.CreatePreSignedDownloadSession"
	reqId := utils.RequestId(ctx)

	var preSignedDownloadObject models.PreSignedDownloadSession

	if validateNotEmptyTrimmedString(bucketId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_id cannot be empty. bucket_id is required to create pre-signed download session", op, reqId, nil)
	}

	if validateNotEmptyTrimmedString(objectId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "object_id cannot be empty. object_id is required to create pre-signed download session", op, reqId, nil)
	}

	if expiresIn == 0 {
		expiresIn = os.config.DefaultPreSignedUploadUrlExpiry
	} else {
		err := validateExpiration(expiresIn)
		if err != nil {
			return nil, srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
		}
	}

	err := os.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := os.getBucketByIdTxn(ctx, tx, bucketId, op)
		if err != nil {
			return err
		}

		object, err := os.queries.WithTx(tx).ObjectGetById(ctx, objectId)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", objectId), op, reqId, err)
			}
			os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, reqId, err)
		}

		if object.UploadStatus == models.ObjectUploadStatusPending {
			objectExists, err := os.storage.CheckIfObjectExists(ctx, &storage.ObjectExistsCheck{
				Bucket: bucket.Name,
				Name:   object.Name,
			})
			if err != nil {
				os.logger.Error("failed to check if object exists in storage", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
				return srverr.NewServiceError(srverr.UnknownError, "failed to check if object exists in storage", op, reqId, err)
			}

			if objectExists {
				err = os.queries.WithTx(tx).ObjectUpdateUploadStatus(ctx, &database.ObjectUpdateUploadStatusParams{
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
			return srverr.NewServiceError(srverr.UnknownError, "failed to create pre-signed download session", op, reqId, err)
		}

		err = os.queries.WithTx(tx).ObjectUpdateLastAccessedAt(ctx, object.ID)
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

func (os *ObjectService) DeleteObject(ctx context.Context, bucketId string, objectId string) error {
	const op = "ObjectService.DeleteObject"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(bucketId) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket_id cannot be empty. bucket_id is required to delete object", op, reqId, nil)
	}

	if validateNotEmptyTrimmedString(objectId) {
		return srverr.NewServiceError(srverr.InvalidInputError, "object_id cannot be empty. object_id is required to delete object", op, reqId, nil)
	}

	err := os.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := os.getBucketByIdTxn(ctx, tx, bucketId, op)
		if err != nil {
			return err
		}

		object, err := os.queries.WithTx(tx).ObjectGetById(ctx, objectId)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("object '%s' not found", objectId), op, reqId, err)
			}
			os.logger.Error("failed to get object from database", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to get object from database", op, reqId, err)
		}

		if object.UploadStatus == models.ObjectUploadStatusPending {
			return srverr.NewServiceError(srverr.BadRequestError, fmt.Sprintf("upload has not yet been completed for object '%s'. delete operation can only be performed on objects that have been uploaded", object.ID), op, reqId, nil)
		}

		err = os.storage.DeleteObject(ctx, &storage.ObjectDelete{
			Bucket: bucket.Name,
			Name:   object.Name,
		})
		if err != nil {
			os.logger.Error("failed to delete object from storage", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to delete object from storage", op, reqId, err)
		}

		err = os.queries.WithTx(tx).ObjectDelete(ctx, object.ID)
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

func (os *ObjectService) GetObject(ctx context.Context, bucketId string, objectId string) (*models.Object, error) {
	const op = "ObjectService.GetObject"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(bucketId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_id cannot be empty. bucket_id is required to get object", op, reqId, nil)
	}

	if validateNotEmptyTrimmedString(objectId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "object_id cannot be empty. object_id is required to get object", op, reqId, nil)
	}

	bucket, err := os.getBucketById(ctx, bucketId, op)
	if err != nil {
		return nil, err
	}

	object, err := os.queries.ObjectGetByBucketIdAndName(ctx, &database.ObjectGetByBucketIdAndNameParams{
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
		MimeType:     object.MimeType,
		Size:         object.Size,
		Metadata:     metadataMap,
		UploadStatus: object.UploadStatus,
		CreatedAt:    object.CreatedAt,
		UpdatedAt:    object.UpdatedAt,
	}, nil
}

func (os *ObjectService) SearchObjects(ctx context.Context, bucketId string, objectPath string, limit int32, offset int32) ([]*models.Object, error) {
	const op = "ObjectService.SearchObjects"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(bucketId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_id cannot be empty. bucket_id is required to search objects", op, reqId, nil)
	}

	if validateNotEmptyTrimmedString(objectPath) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "object_path cannot be empty. object_path is required to search objects", op, reqId, nil)
	}

	if limit < 0 {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "limit cannot be less than 0", op, reqId, nil)
	}

	if offset < 0 {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "offset cannot be less than 0", op, reqId, nil)
	}

	if limit == 0 {
		limit = 100
	}

	objects, err := os.queries.ObjectSearchByBucketIdAndObjectPath(ctx, &database.ObjectSearchByBucketIdAndObjectPathParams{
		BucketID:   bucketId,
		ObjectPath: objectPath,
		Limit:      limit,
		Offset:     offset,
	})
	if err != nil {
		os.logger.Error("failed to search objects", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to search objects", op, reqId, err)
	}
	if len(objects) == 0 {
		return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("no objects found for bucket '%s' with path '%s'", bucketId, objectPath), op, reqId, nil)
	}

	var result []*models.Object

	for _, object := range objects {
		metadataMap, _ := bytesToMetadata(object.Metadata)
		result = append(result, &models.Object{
			Id:           object.ID,
			Version:      object.Version,
			BucketId:     object.BucketID,
			Name:         object.Name,
			MimeType:     object.MimeType,
			Size:         object.Size,
			Metadata:     metadataMap,
			UploadStatus: object.UploadStatus,
			CreatedAt:    object.CreatedAt,
			UpdatedAt:    object.UpdatedAt,
		})
	}

	return result, nil
}

func (os *ObjectService) getBucketByIdTxn(ctx context.Context, tx pgx.Tx, bucketId string, op string) (*models.Bucket, error) {
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(bucketId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_id cannot be empty. bucket_id is required", op, reqId, nil)
	}

	bucket, err := os.queries.WithTx(tx).BucketGetById(ctx, bucketId)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket '%s' not found", bucketId), op, reqId, err)
		}
		os.logger.Error("failed to get bucket by id", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket by id", op, "", err)
	}

	if bucket.Disabled {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is disabled", bucket.ID), op, reqId, err)
	}

	if bucket.Locked {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is locked for '%s'", bucket.ID, *bucket.LockReason), op, reqId, err)
	}

	return &models.Bucket{
		Id:                   bucket.ID,
		Version:              bucket.Version,
		Name:                 bucket.Name,
		AllowedMimeTypes:     bucket.AllowedMimeTypes,
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

func (os *ObjectService) getBucketById(ctx context.Context, bucketId string, op string) (*models.Bucket, error) {
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(bucketId) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket_id cannot be empty. bucket_id is required", op, reqId, nil)
	}

	bucket, err := os.queries.BucketGetById(ctx, bucketId)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket '%s' not found", bucketId), op, reqId, err)
		}
		os.logger.Error("failed to get bucket by id", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket by id", op, reqId, err)
	}

	if bucket.Disabled {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is disabled", bucket.ID), op, reqId, err)
	}

	if bucket.Locked {
		return nil, srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is locked for '%s'", bucket.ID, *bucket.LockReason), op, reqId, err)
	}

	return &models.Bucket{
		Id:                   bucket.ID,
		Version:              bucket.Version,
		Name:                 bucket.Name,
		AllowedMimeTypes:     bucket.AllowedMimeTypes,
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
