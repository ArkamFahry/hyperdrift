package bucket

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/entities"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/srverr"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/validators"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/zapfield"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type BucketService struct {
	query       *database.Queries
	transaction *database.Transaction
	logger      *zap.Logger
}

func NewBucketService(db *pgxpool.Pool, logger *zap.Logger) *BucketService {
	return &BucketService{
		query:       database.New(db),
		transaction: database.NewTransaction(db),
		logger:      logger,
	}
}

func (bs *BucketService) CreateBucket(ctx context.Context, bucketCreate *dto.BucketCreate) (*entities.Bucket, error) {
	const op = "BucketService.CreateBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketCreate.Name) {
		bs.logger.Error("bucket name cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket name cannot be empty")
	}

	if bucketCreate.AllowedContentTypes != nil {
		err := validators.ValidateAllowedContentTypes(bucketCreate.AllowedContentTypes)
		if err != nil {
			bs.logger.Error("failed to validate mime types", zap.Error(err), zapfield.Operation(op))
			return nil, err
		}
	} else {
		bucketCreate.AllowedContentTypes = []string{"*/*"}
	}

	if bucketCreate.MaxAllowedObjectSize != nil {
		err := validators.ValidateMaxAllowedObjectSize(*bucketCreate.MaxAllowedObjectSize)
		if err != nil {
			bs.logger.Error("failed to validate max allowed object size", zap.Error(err), zapfield.Operation(op))
			return nil, err
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

	bucket, err := bs.query.GetBucketById(ctx, bucketCreate.Id)
	if err != nil {
		return nil, err
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

func (bs *BucketService) EnableBucket(ctx context.Context, id string) (*entities.Bucket, error) {
	const op = "BucketService.EnableBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		bs.logger.Error("bucket id cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty")
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with id '%s' not found", id), op, "", err)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to get bucket", op, "", err)
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
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return fmt.Errorf("bucket with id '%s' not found", id)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		if !bucket.Disabled {
			err = bs.query.WithTx(tx).DisableBucket(ctx, id)
			if err != nil {
				bs.logger.Error("failed to disable bucket", zap.Error(err), zapfield.Operation(op))
				return err
			}
		} else {
			bs.logger.Error("failed to disable bucket as it is already disabled", zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is already disabled")
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

func (bs *BucketService) AddAllowedContentTypesToBucket(ctx context.Context, bucketAddAllowedContentTypes *dto.BucketAddAllowedContentTypes) (*entities.Bucket, error) {
	const op = "BucketService.AddAllowedContentTypesToBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketAddAllowedContentTypes.Id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, bucketAddAllowedContentTypes.Id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with id '%s' not found", bucketAddAllowedContentTypes.Id), op, "", nil)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to get bucket", op, "", err)
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is disabled", bucket.ID), op, "", nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is locked for %s", bucket.ID, *bucket.LockReason), op, "", nil)
		}

		if bucketAddAllowedContentTypes.AllowedContentTypes == nil {
			return srverr.NewServiceError(srverr.InvalidInputError, "allowed content types cannot be empty", op, "", nil)
		} else {
			if lo.Contains[string](bucketAddAllowedContentTypes.AllowedContentTypes, "*/*") {
				return srverr.NewServiceError(srverr.InvalidInputError, "wildcard '*/*' cannot be used as an allowed content type", op, "", nil)
			}
			err = validators.ValidateAllowedContentTypes(bucketAddAllowedContentTypes.AllowedContentTypes)
			if err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", nil)
			}
		}

		if lo.Contains[string](bucket.AllowedContentTypes, "*/*") {
			bucket.AllowedContentTypes = []string{}
		}

		bucket.AllowedContentTypes = lo.Uniq[string](append(bucket.AllowedContentTypes, bucketAddAllowedContentTypes.AllowedContentTypes...))

		err = bs.query.WithTx(tx).UpdateBucketAllowedContentTypes(ctx, &database.UpdateBucketAllowedContentTypesParams{
			ID:                  bucket.ID,
			AllowedContentTypes: bucket.AllowedContentTypes,
		})
		if err != nil {
			bs.logger.Error("failed to add allowed content types to bucket", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to add allowed content types to bucket", op, "", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.GetBucketById(ctx, bucketAddAllowedContentTypes.Id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) RemoveContentTypesFromBucket(ctx context.Context, bucketRemoveAllowedContentTypes *dto.BucketRemoveAllowedContentTypes) (*entities.Bucket, error) {
	const op = "BucketService.RemoveContentTypesFromBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketRemoveAllowedContentTypes.Id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, bucketRemoveAllowedContentTypes.Id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with id '%s' not found", bucketRemoveAllowedContentTypes.Id), op, "", nil)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is disabled", bucket.ID), op, "", nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is locked for %s", bucket.ID, *bucket.LockReason), op, "", nil)
		}

		if bucketRemoveAllowedContentTypes.AllowedContentTypes == nil {
			return srverr.NewServiceError(srverr.InvalidInputError, "allowed content types cannot be empty", op, "", nil)
		} else {
			err = validators.ValidateAllowedContentTypes(bucketRemoveAllowedContentTypes.AllowedContentTypes)
			if err != nil {
				return srverr.NewServiceError(srverr.InvalidInputError, err.Error(), op, "", err)
			}
		}

		if lo.Contains[string](bucketRemoveAllowedContentTypes.AllowedContentTypes, "*/*") {
			bucketRemoveAllowedContentTypes.AllowedContentTypes = []string{"*/*"}
			bucket.AllowedContentTypes = []string{}
		} else {
			bucket.AllowedContentTypes = lo.Filter[string](bucket.AllowedContentTypes, func(contentType string, _ int) bool {
				return !lo.Contains[string](bucketRemoveAllowedContentTypes.AllowedContentTypes, contentType)
			})
		}

		err = bs.query.WithTx(tx).UpdateBucketAllowedContentTypes(ctx, &database.UpdateBucketAllowedContentTypesParams{
			ID:                  bucket.ID,
			AllowedContentTypes: bucket.AllowedContentTypes,
		})
		if err != nil {
			bs.logger.Error("failed to remove allowed content types from bucket", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to remove allowed content types from bucket", op, "", err)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.GetBucketById(ctx, bucketRemoveAllowedContentTypes.Id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) UpdateBucket(ctx context.Context, bucketUpdate *dto.BucketUpdate) (*entities.Bucket, error) {
	const op = "BucketService.UpdateBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketUpdate.Id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, bucketUpdate.Id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with id '%s' not found", bucketUpdate.Id), op, "", err)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to get bucket", op, "", err)
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is disabled", bucketUpdate.Id), op, "", nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is locked for %s", bucketUpdate.Id, *bucket.LockReason), op, "", nil)
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

	bucket, err := bs.GetBucketById(ctx, bucketUpdate.Id)
	if err != nil {
		return nil, err
	}

	return bucket, nil
}

func (bs *BucketService) DeleteBucket(ctx context.Context, id string) error {
	const op = "BucketService.DeleteBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty", op, "", nil)
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return srverr.NewServiceError(srverr.NotFoundError, fmt.Sprintf("bucket with id '%s' not found", id), op, "", err)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return srverr.NewServiceError(srverr.UnknownError, "failed to get bucket", op, "", err)
		}

		if bucket.Disabled {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is disabled and cannot be deleted", id), op, "", nil)
		}

		if bucket.Locked {
			return srverr.NewServiceError(srverr.ForbiddenError, fmt.Sprintf("bucket with id '%s' is locked and cannot be deleted", id), op, "", nil)
		}

		err = bs.query.WithTx(tx).DeleteBucket(ctx, bucket.ID)
		if err != nil {
			return srverr.NewServiceError(srverr.UnknownError, "failed to delete bucket", op, "", err)
		}

		return nil
	})
	if err != nil {
		bs.logger.Error("failed to delete bucket", zap.Error(err), zapfield.Operation(op))
		return err
	}

	return nil
}

func (bs *BucketService) GetBucketById(ctx context.Context, id string) (*entities.Bucket, error) {
	const op = "BucketService.GetBucketById"

	if validators.ValidateNotEmptyTrimmedString(id) {
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty", op, "", nil)
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
		return nil, srverr.NewServiceError(srverr.InvalidInputError, "bucket id cannot be empty", op, "", nil)
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
