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
	"go.uber.org/zap"
)

type BucketService struct {
	database    *database.Queries
	transaction *database.Transaction
	logger      *zap.Logger
}

func NewBucketService(db *pgxpool.Pool, logger *zap.Logger) *BucketService {
	return &BucketService{
		database:    database.New(db),
		transaction: database.NewTransaction(db),
		logger:      logger,
	}
}

func (bs *BucketService) CreateBucket(ctx context.Context, bucketCreate *dto.BucketCreate) error {
	const op = "BucketService.CreateBucket"

	if bucketCreate.AllowedContentTypes != nil {
		err := validators.ValidateAllowedContentTypes(bucketCreate.AllowedContentTypes)
		if err != nil {
			bs.logger.Error("failed to validate mime types", zap.Error(err), zapfield.Operation(op))
			return err
		}
	}

	if bucketCreate.MaxAllowedObjectSize != nil {
		err := validators.ValidateMaxAllowedObjectSize(*bucketCreate.MaxAllowedObjectSize)
		if err != nil {
			bs.logger.Error("failed to validate max allowed object size", zap.Error(err), zapfield.Operation(op))
			return err
		}
	}

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		err := bs.database.WithTx(tx).CreateBucket(ctx, &database.CreateBucketParams{
			ID:                   bucketCreate.Id,
			Name:                 bucketCreate.Name,
			AllowedContentTypes:  bucketCreate.AllowedContentTypes,
			MaxAllowedObjectSize: bucketCreate.MaxAllowedObjectSize,
			Public:               bucketCreate.Public,
			Disabled:             bucketCreate.Disabled,
		})
		if err != nil {
			bs.logger.Error("failed to create bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		return nil
	})
	if err != nil {
		bs.logger.Error("failed to create bucket", zap.Error(err), zapfield.Operation(op))
		return err
	}

	return nil
}

func (bs *BucketService) UpdateBucket(ctx context.Context, bucketUpdate *dto.BucketUpdate) (*entities.Bucket, error) {
	const op = "BucketService.UpdateBucket"
	var updatedBucket *entities.Bucket

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.database.WithTx(tx).GetBucketById(ctx, bucketUpdate.Id)
		if err != nil {
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

		err = bs.database.WithTx(tx).UpdateBucket(ctx, &database.UpdateBucketParams{
			ID:                   bucket.ID,
			MaxAllowedObjectSize: bucket.MaxAllowedObjectSize,
			Public:               &bucket.Public,
		})
		if err != nil {
			bs.logger.Error("failed to update bucket", zap.Error(err), zapfield.Operation(op))
			return err
		}

		bucket, err = bs.database.WithTx(tx).GetBucketById(ctx, bucket.ID)
		if err != nil {
			bs.logger.Error("failed to get bucket after update", zap.Error(err), zapfield.Operation(op))
			return err
		}

		updatedBucket = &entities.Bucket{
			Id:                   bucket.ID,
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
		}

		return nil
	})
	if err != nil {
		bs.logger.Error("failed to update bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return updatedBucket, nil
}

func (bs *BucketService) DeleteBucket(ctx context.Context, deleteBucket *dto.BucketDelete) error {
	const op = "BucketService.DeleteBucket"

	err := bs.transaction.WithTransaction(ctx, func(tx pgx.Tx) error {
		bucket, err := bs.database.WithTx(tx).GetBucketById(ctx, deleteBucket.Id)
		if err != nil {
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

		err = bs.database.WithTx(tx).DeleteBucket(ctx, bucket.ID)
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

	bucket, err := bs.database.GetBucketById(ctx, id)
	if err != nil {
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &entities.Bucket{
		Id:                   bucket.ID,
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
