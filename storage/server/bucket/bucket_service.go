package bucket

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/entities"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database"
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
			bs.logger.Error("bucket already exists", zap.Error(err), zapfield.Operation(op))
			return nil, fmt.Errorf("bucket with name already exists")
		}
		bs.logger.Error("failed to create bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
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
				return fmt.Errorf("bucket with id '%s' not found", id)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		if bucket.Disabled {
			err := bs.query.WithTx(tx).EnableBucket(ctx, id)
			if err != nil {
				bs.logger.Error("failed to enable bucket", zap.Error(err), zapfield.Operation(op))
				return err
			}
		} else {
			bs.logger.Error("failed to enable bucket as it is already enabled", zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is already enabled")
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.query.GetBucketById(ctx, id)
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

func (bs *BucketService) DisableBucket(ctx context.Context, id string) (*entities.Bucket, error) {
	const op = "BucketService.DisableBucket"

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

	bucket, err := bs.query.GetBucketById(ctx, id)
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

func (bs *BucketService) AddAllowedContentTypesToBucket(ctx context.Context, bucketAddAllowedContentTypes *dto.BucketAddAllowedContentTypes) (*entities.Bucket, error) {
	const op = "BucketService.AddAllowedContentTypesToBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketAddAllowedContentTypes.Id) {
		bs.logger.Error("bucket id cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty")
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, bucketAddAllowedContentTypes.Id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return fmt.Errorf("bucket with id '%s' not found", bucketAddAllowedContentTypes.Id)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		if bucket.Disabled {
			bs.logger.Error("failed to update bucket as it is disabled", zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is disabled and cannot be updated")
		}

		if bucket.Locked {
			bs.logger.Error(fmt.Sprintf("failed to update bucket as it is locked: %s", *bucket.LockReason), zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is locked and cannot be updated")
		}

		if bucketAddAllowedContentTypes.AllowedContentTypes == nil {
			bs.logger.Error("allowed content types cannot be empty", zapfield.Operation(op))
			return fmt.Errorf("allowed content types cannot be empty")
		} else {
			err = validators.ValidateAllowedContentTypes(bucketAddAllowedContentTypes.AllowedContentTypes)
			if err != nil {
				bs.logger.Error("failed to validate mime types", zap.Error(err), zapfield.Operation(op))
				return err
			}
			if lo.Contains[string](bucketAddAllowedContentTypes.AllowedContentTypes, "*/*") {
				bs.logger.Error("allowed content types cannot contain */*", zapfield.Operation(op))
				return fmt.Errorf("allowed content types cannot contain '*/*'")
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
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.query.GetBucketById(ctx, bucketAddAllowedContentTypes.Id)
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

func (bs *BucketService) RemoveContentTypesFromBucket(ctx context.Context, bucketRemoveAllowedContentTypes *dto.BucketRemoveAllowedContentTypes) (*entities.Bucket, error) {
	const op = "BucketService.RemoveContentTypesFromBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketRemoveAllowedContentTypes.Id) {
		bs.logger.Error("bucket id cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty")
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, bucketRemoveAllowedContentTypes.Id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return fmt.Errorf("bucket with id '%s' not found", bucketRemoveAllowedContentTypes.Id)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		if bucket.Disabled {
			bs.logger.Error("failed to update bucket as it is disabled", zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is disabled and cannot be updated")
		}

		if bucket.Locked {
			bs.logger.Error(fmt.Sprintf("failed to update bucket as it is locked: %s", *bucket.LockReason), zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is locked and cannot be updated")
		}

		if bucketRemoveAllowedContentTypes.AllowedContentTypes == nil {
			bs.logger.Error("allowed content types cannot be empty", zapfield.Operation(op))
			return fmt.Errorf("allowed content types cannot be empty")
		} else {
			err = validators.ValidateAllowedContentTypes(bucketRemoveAllowedContentTypes.AllowedContentTypes)
			if err != nil {
				bs.logger.Error("failed to validate mime types", zap.Error(err), zapfield.Operation(op))
				return err
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
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.query.GetBucketById(ctx, bucketRemoveAllowedContentTypes.Id)
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

func (bs *BucketService) UpdateBucket(ctx context.Context, bucketUpdate *dto.BucketUpdate) (*entities.Bucket, error) {
	const op = "BucketService.UpdateBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketUpdate.Id) {
		bs.logger.Error("bucket name cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty")
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.query.WithTx(tx).GetBucketById(ctx, bucketUpdate.Id)
		if err != nil {
			if database.IsNotFoundError(err) {
				return fmt.Errorf("bucket with id '%s' not found", bucketUpdate.Id)
			}
			bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		if bucket.Disabled {
			bs.logger.Error("failed to update bucket as it is disabled", zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is disabled and cannot be updated")
		}

		if bucket.Locked {
			bs.logger.Error(fmt.Sprintf("failed to update bucket as it is locked: %s", *bucket.LockReason), zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is locked and cannot be updated")
		}

		if bucketUpdate.MaxAllowedObjectSize != nil {
			err = validators.ValidateMaxAllowedObjectSize(*bucketUpdate.MaxAllowedObjectSize)
			if err != nil {
				bs.logger.Error("not allowed max object size", zap.Error(err), zapfield.Operation(op))
				return err
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
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	bucket, err := bs.query.GetBucketById(ctx, bucketUpdate.Id)
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

func (bs *BucketService) DeleteBucket(ctx context.Context, id string) error {
	const op = "BucketService.DeleteBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		bs.logger.Error("bucket name cannot be empty", zapfield.Operation(op))
		return fmt.Errorf("bucket id cannot be empty")
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

		if bucket.Disabled {
			bs.logger.Error("failed to delete bucket as it is disabled", zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is disabled and cannot be deleted")
		}

		if bucket.Locked {
			bs.logger.Error(fmt.Sprintf("failed to delete bucket as it is locked: %s", *bucket.LockReason), zap.Error(err), zapfield.Operation(op))
			return fmt.Errorf("bucket is locked and cannot be deleted")
		}

		err = bs.query.WithTx(tx).DeleteBucket(ctx, bucket.ID)
		if err != nil {
			bs.logger.Error("failed to delete bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		return nil
	})
	if err != nil {
		bs.logger.Error("failed to delete bucket", zap.Error(err), zapfield.Operation(op))
		return err
	}

	return nil
}

func (bs *BucketService) GetBucket(ctx context.Context, id string) (*entities.Bucket, error) {
	const op = "BucketService.GetBucket"

	bucket, err := bs.query.GetBucketById(ctx, id)
	if err != nil {
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
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

func (bs *BucketService) GetBucketSize(ctx context.Context, id string) (*entities.BucketSize, error) {
	const op = "BucketService.GetBucketSize"

	if validators.ValidateNotEmptyTrimmedString(id) {
		bs.logger.Error("bucket name cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty when getting bucket size")
	}

	bucketSize, err := bs.query.GetBucketSizeById(ctx, id)
	if err != nil {
		if database.IsNotFoundError(err) {
			return nil, fmt.Errorf("bucket size with id '%s' not found", id)
		}
		bs.logger.Error("failed to get bucket size", zap.Error(err), zapfield.Operation(op))
		return nil, err
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
			return nil, fmt.Errorf("buckets not found")
		}
		bs.logger.Error("failed to list all buckets", zap.Error(err), zapfield.Operation(op))
		return nil, err
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
