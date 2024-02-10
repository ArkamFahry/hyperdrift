package services

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/entities"
	"github.com/ArkamFahry/hyperdrift/storage/server/jobs"
	"github.com/ArkamFahry/hyperdrift/storage/server/srverr"
	"github.com/ArkamFahry/hyperdrift/storage/server/validators"
	"github.com/ArkamFahry/hyperdrift/storage/server/zapfield"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/samber/lo"
	"go.uber.org/zap"
	"regexp"
)

type BucketService struct {
	query       *database.Queries
	transaction *database.Transaction
	logger      *zap.Logger
	job         *river.Client[pgx.Tx]
}

func NewBucketService(db *pgxpool.Pool, logger *zap.Logger, job *river.Client[pgx.Tx]) *BucketService {
	return &BucketService{
		query:       database.New(db),
		transaction: database.NewTransaction(db),
		logger:      logger,
		job:         job,
	}
}

func (bs *BucketService) CreateBucket(ctx context.Context, bucketCreate *dto.BucketCreate) (*entities.Bucket, error) {
	const op = "BucketService.CreateBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketCreate.Name) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name cannot be empty. bucket name is required to create bucket", op, "", nil)
	}

	if validateBucketName(bucketCreate.Name) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket name is not valid. it must start and end with an alphanumeric character, and can include alphanumeric characters, hyphens, and dots. The total length must be between 3 and 63 characters.", op, "", nil)
	}

	if bucketCreate.AllowedContentTypes != nil {
		if len(bucketCreate.AllowedContentTypes) > 1 {
			if lo.Contains[string](bucketCreate.AllowedContentTypes, "*/*") {
				return nil, srverr.NewServiceError(srverr.InvalidInputError, "wildcard '*/*' is not allowed to be included with other content types. if you want to allow all content types use  '*/*'", op, "", nil)
			}
		}

		err := validators.ValidateAllowedContentTypes(bucketCreate.AllowedContentTypes)
		if err != nil {
			return nil, srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
		}
	} else {
		bucketCreate.AllowedContentTypes = []string{"*/*"}
	}

	if bucketCreate.MaxAllowedObjectSize != nil {
		err := validators.ValidateMaxAllowedObjectSize(*bucketCreate.MaxAllowedObjectSize)
		if err != nil {
			return nil, srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
		}
	}

	err := bs.query.CreateBucket(ctx, &database.CreateBucketParams{
		ID:                   bucketCreate.Id,
		Name:                 bucketCreate.Name,
		AllowedContentTypes:  bucketCreate.AllowedContentTypes,
		MaxAllowedObjectSize: bucketCreate.MaxAllowedObjectSize,
		Public:               bucketCreate.Public,
		Disabled:             bucketCreate.Disabled,
	})
	if err != nil {
		if database.IsConflictError(err) {
			return nil, srverr.NewServiceError(srverr.ConflictError, fmt.Sprintf("bucket with name '%s' already exists", bucketCreate.Name), op, "", err)
		}
		bs.logger.Error("failed to create bucket", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to create bucket", op, "", err)
	}

	bucket, err := bs.GetBucketById(ctx, bucketCreate.Id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) EnableBucket(ctx context.Context, id string) (*entities.Bucket, error) {
	const op = "BucketService.EnableBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to enable bucket", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Disabled {
			err = bs.query.WithTx(tx).EnableBucket(ctx, id)
			if err != nil {
				bs.logger.Error("failed to enable bucket", zap.Error(err), zapfield.Operation(op))
				return srverr.NewServiceError(srverr.UnknownError, "failed to enable bucket", op, "", err)
			}
		} else {
			return srverr.NewServiceError(srverr.BadRequestError, "bucket is already enabled", op, "", nil)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.GetBucketById(ctx, id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) DisableBucket(ctx context.Context, id string) (*entities.Bucket, error) {
	const op = "BucketService.DisableBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to disable bucket", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if !bucket.Disabled {
			err = bs.query.WithTx(tx).DisableBucket(ctx, id)
			if err != nil {
				bs.logger.Error("failed to disable bucket", zap.Error(err), zapfield.Operation(op))
				return srverr.NewServiceError(srverr.UnknownError, "failed to disable bucket", op, "", err)
			}
		} else {
			return srverr.NewServiceError(srverr.BadRequestError, "bucket is already disabled", op, "", nil)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.GetBucketById(ctx, id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) UpdateBucket(ctx context.Context, id string, bucketUpdate *dto.BucketUpdate) (*entities.Bucket, error) {
	const op = "BucketService.UpdateBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to update bucket", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is disabled and cannot be updated", bucket.ID), op, "", nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is locked for '%s' and cannot be updated", bucket.ID, *bucket.LockReason), op, "", nil)
		}

		if bucketUpdate.AllowedContentTypes != nil {
			if len(bucketUpdate.AllowedContentTypes) > 1 {
				if lo.Contains[string](bucketUpdate.AllowedContentTypes, "*/*") {
					return srverr.NewServiceError(srverr.InvalidInputError, "wildcard '*/*' is not allowed to be included with other content types. if you want to allow all content types use  '*/*'", op, "", nil)
				}
			}

			err = validators.ValidateAllowedContentTypes(bucketUpdate.AllowedContentTypes)
			if err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
			}

			bucket.AllowedContentTypes = bucketUpdate.AllowedContentTypes
		}

		if bucketUpdate.MaxAllowedObjectSize != nil {
			err = validators.ValidateMaxAllowedObjectSize(*bucketUpdate.MaxAllowedObjectSize)
			if err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
			}
			bucket.MaxAllowedObjectSize = bucketUpdate.MaxAllowedObjectSize
		}

		if bucketUpdate.Public != nil {
			bucket.Public = *bucketUpdate.Public
		}

		err = bs.query.WithTx(tx).UpdateBucket(ctx, &database.UpdateBucketParams{
			ID:                   bucket.ID,
			AllowedContentTypes:  bucket.AllowedContentTypes,
			MaxAllowedObjectSize: bucket.MaxAllowedObjectSize,
			Public:               &bucket.Public,
		})
		if err != nil {
			bs.logger.Error("failed to update bucket", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to update bucket", op, "", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.GetBucketById(ctx, id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) EmptyBucket(ctx context.Context, id string) error {
	const op = "BucketService.EmptyBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to empty bucket", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is disabled and cannot be emptied", bucket.ID), op, "", nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is locked for '%s' and cannot be emptied", bucket.ID, *bucket.LockReason), op, "", nil)
		}

		err = bs.query.WithTx(tx).LockBucket(ctx, &database.LockBucketParams{
			ID:         bucket.ID,
			LockReason: "bucket.empty",
		})
		if err != nil {
			return srverr.NewServiceError(srverr.UnknownError, "failed to lock bucket for emptying", op, "", err)
		}

		_, err = bs.job.InsertTx(ctx, tx, &jobs.BucketEmpty{
			Id:   bucket.ID,
			Name: bucket.Name,
		}, nil)

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (bs *BucketService) DeleteBucket(ctx context.Context, id string) error {
	const op = "BucketService.DeleteBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to delete bucket", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.getBucketByIdTxn(ctx, tx, id, op)
		if err != nil {
			return err
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is disabled and cannot be deleted", bucket.ID), op, "", nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is locked for '%s' and cannot be deleted", bucket.ID, *bucket.LockReason), op, "", nil)
		}

		err = bs.query.WithTx(tx).LockBucket(ctx, &database.LockBucketParams{
			ID:         bucket.ID,
			LockReason: "bucket.delete",
		})
		if err != nil {
			return srverr.NewServiceError(srverr.UnknownError, "failed to lock bucket for deletion", op, "", err)
		}

		_, err = bs.job.InsertTx(ctx, tx, jobs.BucketDelete{
			Id:   bucket.ID,
			Name: bucket.Name,
		}, nil)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (bs *BucketService) GetBucketById(ctx context.Context, id string) (*entities.Bucket, error) {
	const op = "BucketService.GetBucketById"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to get bucket", op, "", nil)
	}

	bucket, err := bs.query.GetBucketById(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with id '%s' not found", id), op, "", err)
		}
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket", op, "", err)
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

func (bs *BucketService) GetBucketSize(ctx context.Context, id string) (*entities.BucketSize, error) {
	const op = "BucketService.GetBucketSize"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty. bucket id is required to get bucket size", op, "", nil)
	}

	bucketSize, err := bs.query.GetBucketSizeById(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with id '%s' not found", id), op, "", err)
		}
		bs.logger.Error("failed to get bucket size", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket size", op, "", err)
	}

	return &entities.BucketSize{
		Id:   bucketSize.ID,
		Name: bucketSize.Name,
		Size: bucketSize.Size,
	}, nil
}

func (bs *BucketService) ListAllBuckets(ctx context.Context) ([]*entities.Bucket, error) {
	const op = "BucketService.ListAllBuckets"

	buckets, err := bs.query.ListAllBuckets(ctx)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, "no buckets found", op, "", err)
		}
		bs.logger.Error("failed to list all buckets", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to list all buckets", op, "", err)
	}

	var result []*entities.Bucket

	for _, bucket := range buckets {
		result = append(result, &entities.Bucket{
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
		})
	}

	return result, nil
}

func (bs *BucketService) getBucketByIdTxn(ctx context.Context, tx pgx.Tx, id string, op string) (*database.StorageBucket, error) {
	bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with id '%s' not found", id), op, "", err)
		}
		bs.logger.Error("failed to get bucket by id", zap.Error(err), zapfield.Operation(op))
		return nil, srverr.NewServiceError(srverr.UnknownError, "failed to get bucket by id", op, "", err)
	}
	return bucket, nil
}

func validateBucketName(name string) bool {
	regexPattern := `^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`

	regex := regexp.MustCompile(regexPattern)

	if len(name) < 3 || len(name) > 63 {
		return true
	}

	if regex.MatchString(name) {
		return false
	} else {
		return true
	}
}
