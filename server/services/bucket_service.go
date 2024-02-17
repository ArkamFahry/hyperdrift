package services

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/storage/server/database"
	"github.com/ArkamFahry/storage/server/jobs"
	"github.com/ArkamFahry/storage/server/models"
	"github.com/ArkamFahry/storage/server/srverr"
	"github.com/ArkamFahry/storage/server/utils"
	"github.com/ArkamFahry/storage/server/zapfield"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type BucketService struct {
	query       *database.Queries
	transaction *database.Transaction
	job         *river.Client[pgx.Tx]
	logger      *zap.Logger
}

func NewBucketService(db *pgxpool.Pool, job *river.Client[pgx.Tx], logger *zap.Logger) *BucketService {
	return &BucketService{
		query:       database.New(db),
		transaction: database.NewTransaction(db),
		job:         job,
		logger:      logger,
	}
}

func (bs *BucketService) CreateBucket(ctx context.Context, bucketCreate *models.BucketCreate) (*models.Bucket, error) {
	const op = "BucketService.CreateBucket"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(bucketCreate.Name) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required to create bucket", op, reqId, nil)
	}

	if validateBucketName(bucketCreate.Name) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name is not valid. it must start and end with an alphanumeric character, and can include alphanumeric characters, hyphens, and dots. The total length must be between 3 and 63 characters.", op, reqId, nil)
	}

	if bucketCreate.AllowedMimeTypes != nil {
		if len(bucketCreate.AllowedMimeTypes) > 1 {
			if lo.Contains[string](bucketCreate.AllowedMimeTypes, models.BucketAllowedWildcardMimeType) {
				return nil, srverr.NewServiceError(srverr.InvalidInputError, "wildcard '*/*' is not allowed to be included with other content types. if you want to allow all content types only use '*/*'", op, reqId, nil)
			}
		}

		if err := validateAllowedMimeTypes(bucketCreate.AllowedMimeTypes); err != nil {
			return nil, srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
		}
	} else {
		bucketCreate.AllowedMimeTypes = []string{models.BucketAllowedWildcardMimeType}
	}

	if bucketCreate.MaxAllowedObjectSize != nil {
		if err := validateMaxAllowedObjectSize(*bucketCreate.MaxAllowedObjectSize); err != nil {
			return nil, srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
		}
	}

	id, err := bs.query.BucketCreate(ctx, &database.BucketCreateParams{
		Name:                 bucketCreate.Name,
		AllowedMimeTypes:     bucketCreate.AllowedMimeTypes,
		MaxAllowedObjectSize: bucketCreate.MaxAllowedObjectSize,
		Public:               bucketCreate.Public,
	})
	if err != nil {
		if database.IsConflictError(err) {
			return nil, srverr.NewServiceError(srverr.ConflictError, fmt.Sprintf("bucket with name '%s' already exists", bucketCreate.Name), op, reqId, err)
		}
		bs.logger.Error("failed to create bucket", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to create bucket", op, reqId, err)
	}

	bucket, err := bs.GetBucket(ctx, id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) UpdateBucket(ctx context.Context, id string, bucketUpdate *models.BucketUpdate) (*models.Bucket, error) {
	const op = "BucketService.UpdateBucket"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to update bucket", op, reqId, nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is disabled and cannot be updated", bucket.ID), op, reqId, nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is locked for '%s' and cannot be updated", bucket.ID, *bucket.LockReason), op, reqId, nil)
		}

		if bucketUpdate.AllowedMimeTypes != nil {
			if len(bucketUpdate.AllowedMimeTypes) > 1 {
				if lo.Contains[string](bucketUpdate.AllowedMimeTypes, models.BucketAllowedWildcardMimeType) {
					return srverr.NewServiceError(srverr.InvalidInputError, "wildcard '*/*' is not allowed to be included with other content types. if you want to allow all content types only add '*/*' in allowed content types", op, reqId, nil)
				}
			}

			if err = validateAllowedMimeTypes(bucketUpdate.AllowedMimeTypes); err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
			}

			bucket.AllowedMimeTypes = bucketUpdate.AllowedMimeTypes
		}

		if bucketUpdate.MaxAllowedObjectSize != nil {
			if err = validateMaxAllowedObjectSize(*bucketUpdate.MaxAllowedObjectSize); err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, reqId, err)
			}
			bucket.MaxAllowedObjectSize = bucketUpdate.MaxAllowedObjectSize
		}

		if bucketUpdate.Public != nil {
			bucket.Public = *bucketUpdate.Public
		}

		err = bs.query.WithTx(tx).BucketUpdate(ctx, &database.BucketUpdateParams{
			ID:                   bucket.ID,
			AllowedMimeTypes:     bucket.AllowedMimeTypes,
			MaxAllowedObjectSize: bucket.MaxAllowedObjectSize,
			Public:               &bucket.Public,
		})
		if err != nil {
			bs.logger.Error("failed to update bucket", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to update bucket", op, reqId, err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.GetBucket(ctx, id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) EnableBucket(ctx context.Context, id string) (*models.Bucket, error) {
	const op = "BucketService.EnableBucket"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to enable bucket", op, reqId, nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.BadRequestError, fmt.Sprintf("bucket '%s' is locked for '%s' and cannot be enabled", bucket.ID, *bucket.LockReason), op, reqId, nil)
		}

		if bucket.Disabled {
			err = bs.query.WithTx(tx).BucketEnable(ctx, id)
			if err != nil {
				bs.logger.Error("failed to enable bucket", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
				return srverr.NewServiceError(srverr.UnknownError, "failed to enable bucket", op, reqId, err)
			}
		} else {
			return srverr.NewServiceError(srverr.BadRequestError, "bucket is already enabled", op, reqId, nil)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.GetBucket(ctx, id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) DisableBucket(ctx context.Context, id string) (*models.Bucket, error) {
	const op = "BucketService.DisableBucket"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to disable bucket", op, reqId, nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.BadRequestError, fmt.Sprintf("bucket '%s' is locked for '%s' and cannot be disabled", bucket.ID, *bucket.LockReason), op, reqId, nil)
		}

		if !bucket.Disabled {
			err = bs.query.WithTx(tx).BucketDisable(ctx, id)
			if err != nil {
				bs.logger.Error("failed to disable bucket", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
				return srverr.NewServiceError(srverr.UnknownError, "failed to disable bucket", op, reqId, err)
			}
		} else {
			return srverr.NewServiceError(srverr.BadRequestError, "bucket is already disabled", op, reqId, nil)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.GetBucket(ctx, id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) EmptyBucket(ctx context.Context, id string) error {
	const op = "BucketService.EmptyBucket"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(id) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to empty bucket", op, reqId, nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is disabled and cannot be emptied", bucket.ID), op, reqId, nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is locked for '%s' and cannot be emptied", bucket.ID, *bucket.LockReason), op, reqId, nil)
		}

		err = bs.query.WithTx(tx).BucketLock(ctx, &database.BucketLockParams{
			ID:         bucket.ID,
			LockReason: models.BucketLockedReasonBucketEmptying,
		})
		if err != nil {
			bs.logger.Error("failed to lock bucket for emptying", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to lock bucket for emptying", op, reqId, err)
		}

		_, err = bs.job.InsertTx(ctx, tx, &jobs.BucketEmptying{
			BucketId: bucket.ID,
		}, nil)
		if err != nil {
			bs.logger.Error("failed to create bucket empty job", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to create bucket empty job", op, reqId, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (bs *BucketService) DeleteBucket(ctx context.Context, id string) error {
	const op = "BucketService.DeleteBucket"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(id) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to delete bucket", op, reqId, nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is disabled and cannot be deleted", bucket.ID), op, reqId, nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket '%s' is locked for '%s' and cannot be deleted", bucket.ID, *bucket.LockReason), op, reqId, nil)
		}

		err = bs.query.WithTx(tx).BucketLock(ctx, &database.BucketLockParams{
			ID:         bucket.ID,
			LockReason: models.BucketLockedReasonBucketDeletion,
		})
		if err != nil {
			bs.logger.Error("failed to lock bucket for deletion", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to lock bucket for deletion", op, reqId, err)
		}

		_, err = bs.job.InsertTx(ctx, tx, jobs.BucketDeletion{
			BucketId: bucket.ID,
		}, nil)
		if err != nil {
			bs.logger.Error("failed to create bucket delete job", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
			return srverr.NewServiceError(srverr.UnknownError, "failed to create bucket delete job", op, reqId, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (bs *BucketService) GetBucket(ctx context.Context, id string) (*models.Bucket, error) {
	const op = "BucketService.GetBucket"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to get bucket", op, reqId, nil)
	}

	bucket, err := bs.query.BucketGetById(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket '%s' not found", id), op, reqId, err)
		}
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket", op, reqId, err)
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

func (bs *BucketService) GetBucketSize(ctx context.Context, id string) (*models.BucketSize, error) {
	const op = "BucketService.GetBucketSize"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to get bucket size", op, reqId, nil)
	}

	bucketSize, err := bs.query.BucketGetSizeById(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket '%s' not found", id), op, reqId, err)
		}
		bs.logger.Error("failed to get bucket size", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket size", op, reqId, err)
	}

	return &models.BucketSize{
		Id:   bucketSize.ID,
		Name: bucketSize.Name,
		Size: bucketSize.Size,
	}, nil
}

func (bs *BucketService) ListAllBuckets(ctx context.Context) ([]*models.Bucket, error) {
	const op = "BucketService.ListAllBuckets"
	reqId := utils.RequestId(ctx)

	buckets, err := bs.query.BucketListAll(ctx)
	if err != nil {
		bs.logger.Error("failed to list all buckets", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to list all buckets", op, reqId, err)
	}
	if len(buckets) == 0 {
		return nil, srverr.NewServiceError(srverr.NotFoundError, "no buckets found", op, reqId, nil)
	}

	var result []*models.Bucket

	for _, bucket := range buckets {
		result = append(result, &models.Bucket{
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
		})
	}

	return result, nil
}

func (bs *BucketService) SearchBuckets(ctx context.Context, name string) ([]*models.Bucket, error) {
	const op = "BucketService.SearchBuckets"
	reqId := utils.RequestId(ctx)

	if validateNotEmptyTrimmedString(name) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required to search buckets", op, reqId, nil)
	}

	buckets, err := bs.query.BucketSearch(ctx, name)
	if err != nil {
		bs.logger.Error("failed to search buckets", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to search buckets", op, reqId, err)
	}
	if len(buckets) == 0 {
		return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("no buckets found with name '%s'", name), op, reqId, nil)
	}

	var result []*models.Bucket

	for _, bucket := range buckets {
		result = append(result, &models.Bucket{
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
		})
	}

	return result, nil
}

func (bs *BucketService) getBucketByIdTxn(ctx context.Context, tx pgx.Tx, id string, op string) (*database.StorageBucket, error) {
	reqId := utils.RequestId(ctx)

	bucket, err := bs.query.WithTx(tx).BucketGetById(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket '%s' not found", id), op, reqId, err)
		}
		bs.logger.Error("failed to get bucket by id", zap.Error(err), zapfield.Operation(op), zapfield.RequestId(reqId))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket by id", op, reqId, err)
	}
	return bucket, nil
}
