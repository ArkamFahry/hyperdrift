package bucket

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/entities"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/validators"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/zapfield"
	"github.com/samber/lo"
	"go.uber.org/zap"
)

type BucketService struct {
	bucketRepository *BucketRepository
	logger           *zap.Logger
}

func NewBucketService(bucketRepository *BucketRepository, logger *zap.Logger) *BucketService {
	return &BucketService{
		bucketRepository: bucketRepository,
		logger:           logger,
	}
}

func (bs *BucketService) CreateBucket(ctx context.Context, bucketCreate *dto.BucketCreate) (*dto.Bucket, error) {
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

	createdBucket, err := bs.bucketRepository.CreateBucket(ctx, &entities.BucketCreate{
		Id:                   bucketCreate.Id,
		Name:                 bucketCreate.Name,
		AllowedContentTypes:  bucketCreate.AllowedContentTypes,
		MaxAllowedObjectSize: bucketCreate.MaxAllowedObjectSize,
		Public:               bucketCreate.Public,
		Disabled:             bucketCreate.Disabled,
	})
	if err != nil {
		bs.logger.Error("failed to create bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &dto.Bucket{
		Id:                   createdBucket.Id,
		Version:              createdBucket.Version,
		Name:                 createdBucket.Name,
		AllowedContentTypes:  createdBucket.AllowedContentTypes,
		MaxAllowedObjectSize: createdBucket.MaxAllowedObjectSize,
		Public:               createdBucket.Public,
		Disabled:             createdBucket.Disabled,
		Locked:               createdBucket.Locked,
		LockReason:           createdBucket.LockReason,
		LockedAt:             createdBucket.LockedAt,
		CreatedAt:            createdBucket.CreatedAt,
		UpdatedAt:            createdBucket.UpdatedAt,
	}, nil
}

func (bs *BucketService) EnableBucket(ctx context.Context, id string) (*dto.Bucket, error) {
	const op = "BucketService.EnableBucket"

	bucket, err := bs.bucketRepository.GetBucketById(ctx, id)
	if err != nil {
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	if bucket.Disabled {
		bucket, err = bs.bucketRepository.EnableBucket(ctx, &entities.BucketEnable{
			Id:      bucket.Id,
			Version: bucket.Version,
		})
		if err != nil {
			bs.logger.Error("failed to enable bucket", zap.Error(err), zapfield.Operation(op))
			return nil, err
		}
	} else {
		bs.logger.Error("failed to enable bucket as it is already enabled", zap.Error(err), zapfield.Operation(op))
		return nil, fmt.Errorf("bucket is already enabled")
	}

	return &dto.Bucket{
		Id:                   bucket.Id,
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

func (bs *BucketService) DisableBucket(ctx context.Context, id string) (*dto.Bucket, error) {
	const op = "BucketService.DisableBucket"

	bucket, err := bs.bucketRepository.GetBucketById(ctx, id)
	if err != nil {
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	if !bucket.Disabled {
		bucket, err = bs.bucketRepository.DisableBucket(ctx, &entities.BucketDisable{
			Id:      bucket.Id,
			Version: bucket.Version,
		})
		if err != nil {
			bs.logger.Error("failed to disable bucket", zap.Error(err), zapfield.Operation(op))
			return nil, err
		}
	} else {
		bs.logger.Error("failed to disable bucket as it is already disabled", zap.Error(err), zapfield.Operation(op))
		return nil, fmt.Errorf("bucket is already disabled")
	}

	return &dto.Bucket{
		Id:                   bucket.Id,
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

func (bs *BucketService) AddAllowedContentTypesToBucket(ctx context.Context, bucketAddAllowedContentTypes *dto.BucketAddAllowedContentTypes) (*dto.Bucket, error) {
	const op = "BucketService.AddAllowedContentTypesToBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketAddAllowedContentTypes.Id) {
		bs.logger.Error("bucket id cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty")
	}

	bucket, err := bs.bucketRepository.GetBucketById(ctx, bucketAddAllowedContentTypes.Id)
	if err != nil {
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	if bucket.Disabled {
		bs.logger.Error("failed to update bucket as it is disabled", zap.Error(err), zapfield.Operation(op))
		return nil, fmt.Errorf("bucket is disabled and cannot be updated")
	}

	if bucket.Locked {
		bs.logger.Error(fmt.Sprintf("failed to update bucket as it is locked: %s", *bucket.LockReason), zap.Error(err), zapfield.Operation(op))
		return nil, fmt.Errorf("bucket is locked and cannot be updated")
	}

	if bucketAddAllowedContentTypes.AllowedContentTypes == nil {
		bs.logger.Error("allowed content types cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("allowed content types cannot be empty")
	} else {
		err = validators.ValidateAllowedContentTypes(bucketAddAllowedContentTypes.AllowedContentTypes)
		if err != nil {
			bs.logger.Error("failed to validate mime types", zap.Error(err), zapfield.Operation(op))
			return nil, err
		}
		if lo.Contains[string](bucketAddAllowedContentTypes.AllowedContentTypes, "*/*") {
			bs.logger.Error("allowed content types cannot contain */*", zapfield.Operation(op))
			return nil, fmt.Errorf("allowed content types cannot contain '*/*'")
		}
	}

	if lo.Contains[string](bucket.AllowedContentTypes, "*/*") {
		bucket.AllowedContentTypes = []string{}
	}

	bucket.AllowedContentTypes = lo.Uniq[string](append(bucket.AllowedContentTypes, bucketAddAllowedContentTypes.AllowedContentTypes...))

	contentTypesAddedBucket, err := bs.bucketRepository.UpdateBucketAllowedContentTypes(ctx, &entities.BucketAllowedContentTypesUpdate{
		Id:                  bucket.Id,
		AllowedContentTypes: bucket.AllowedContentTypes,
		Version:             bucket.Version,
	})
	if err != nil {
		bs.logger.Error("failed to add allowed content types to bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &dto.Bucket{
		Id:                   contentTypesAddedBucket.Id,
		Version:              contentTypesAddedBucket.Version,
		Name:                 contentTypesAddedBucket.Name,
		AllowedContentTypes:  contentTypesAddedBucket.AllowedContentTypes,
		MaxAllowedObjectSize: contentTypesAddedBucket.MaxAllowedObjectSize,
		Public:               contentTypesAddedBucket.Public,
		Disabled:             contentTypesAddedBucket.Disabled,
		Locked:               contentTypesAddedBucket.Locked,
		LockReason:           contentTypesAddedBucket.LockReason,
		LockedAt:             contentTypesAddedBucket.LockedAt,
		CreatedAt:            contentTypesAddedBucket.CreatedAt,
		UpdatedAt:            contentTypesAddedBucket.UpdatedAt,
	}, nil
}

func (bs *BucketService) RemoveContentTypesFromBucket(ctx context.Context, bucketRemoveAllowedContentTypes *dto.BucketRemoveAllowedContentTypes) (*dto.Bucket, error) {
	const op = "BucketService.RemoveContentTypesFromBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketRemoveAllowedContentTypes.Id) {
		bs.logger.Error("bucket id cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty")
	}

	bucket, err := bs.bucketRepository.GetBucketById(ctx, bucketRemoveAllowedContentTypes.Id)
	if err != nil {
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	if bucket.Disabled {
		bs.logger.Error("failed to update bucket as it is disabled", zap.Error(err), zapfield.Operation(op))
		return nil, fmt.Errorf("bucket is disabled and cannot be updated")
	}

	if bucket.Locked {
		bs.logger.Error(fmt.Sprintf("failed to update bucket as it is locked: %s", *bucket.LockReason), zap.Error(err), zapfield.Operation(op))
		return nil, fmt.Errorf("bucket is locked and cannot be updated")
	}

	if bucketRemoveAllowedContentTypes.AllowedContentTypes == nil {
		bs.logger.Error("allowed content types cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("allowed content types cannot be empty")
	} else {
		err = validators.ValidateAllowedContentTypes(bucketRemoveAllowedContentTypes.AllowedContentTypes)
		if err != nil {
			bs.logger.Error("failed to validate mime types", zap.Error(err), zapfield.Operation(op))
			return nil, err
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

	contentTypesRemovedBucket, err := bs.bucketRepository.UpdateBucketAllowedContentTypes(ctx, &entities.BucketAllowedContentTypesUpdate{
		Id:                  bucket.Id,
		AllowedContentTypes: bucket.AllowedContentTypes,
		Version:             bucket.Version,
	})
	if err != nil {
		bs.logger.Error("failed to remove allowed content types from bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &dto.Bucket{
		Id:                   contentTypesRemovedBucket.Id,
		Version:              contentTypesRemovedBucket.Version,
		Name:                 contentTypesRemovedBucket.Name,
		AllowedContentTypes:  contentTypesRemovedBucket.AllowedContentTypes,
		MaxAllowedObjectSize: contentTypesRemovedBucket.MaxAllowedObjectSize,
		Public:               contentTypesRemovedBucket.Public,
		Disabled:             contentTypesRemovedBucket.Disabled,
		Locked:               contentTypesRemovedBucket.Locked,
		LockReason:           contentTypesRemovedBucket.LockReason,
		LockedAt:             contentTypesRemovedBucket.LockedAt,
		CreatedAt:            contentTypesRemovedBucket.CreatedAt,
		UpdatedAt:            contentTypesRemovedBucket.UpdatedAt,
	}, nil
}

func (bs *BucketService) UpdateBucket(ctx context.Context, bucketUpdate *dto.BucketUpdate) (*dto.Bucket, error) {
	const op = "BucketService.UpdateBucket"

	if validators.ValidateNotEmptyTrimmedString(bucketUpdate.Id) {
		bs.logger.Error("bucket name cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty")
	}

	bucket, err := bs.bucketRepository.GetBucketById(ctx, bucketUpdate.Id)
	if err != nil {
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	if bucket.Disabled {
		bs.logger.Error("failed to update bucket as it is disabled", zap.Error(err), zapfield.Operation(op))
		return nil, fmt.Errorf("bucket is disabled and cannot be updated")
	}

	if bucket.Locked {
		bs.logger.Error(fmt.Sprintf("failed to update bucket as it is locked: %s", *bucket.LockReason), zap.Error(err), zapfield.Operation(op))
		return nil, fmt.Errorf("bucket is locked and cannot be updated")
	}

	if bucketUpdate.MaxAllowedObjectSize != nil {
		err = validators.ValidateMaxAllowedObjectSize(*bucketUpdate.MaxAllowedObjectSize)
		if err != nil {
			bs.logger.Error("not allowed max object size", zap.Error(err), zapfield.Operation(op))
			return nil, err
		}
		bucket.MaxAllowedObjectSize = bucketUpdate.MaxAllowedObjectSize
	}

	if bucketUpdate.Public != nil {
		bucket.Public = *bucketUpdate.Public
	}

	updatedBucket, err := bs.bucketRepository.UpdateBucket(ctx, &entities.BucketUpdate{
		Id:                   bucket.Id,
		MaxAllowedObjectSize: bucket.MaxAllowedObjectSize,
		Public:               &bucket.Public,
		Version:              bucket.Version,
	})
	if err != nil {
		bs.logger.Error("failed to update bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &dto.Bucket{
		Id:                   updatedBucket.Id,
		Version:              updatedBucket.Version,
		Name:                 updatedBucket.Name,
		AllowedContentTypes:  updatedBucket.AllowedContentTypes,
		MaxAllowedObjectSize: updatedBucket.MaxAllowedObjectSize,
		Public:               updatedBucket.Public,
		Disabled:             updatedBucket.Disabled,
		Locked:               updatedBucket.Locked,
		LockReason:           updatedBucket.LockReason,
		LockedAt:             updatedBucket.LockedAt,
		CreatedAt:            updatedBucket.CreatedAt,
		UpdatedAt:            updatedBucket.UpdatedAt,
	}, nil
}

func (bs *BucketService) DeleteBucket(ctx context.Context, id string) error {
	const op = "BucketService.DeleteBucket"

	if validators.ValidateNotEmptyTrimmedString(id) {
		bs.logger.Error("bucket name cannot be empty", zapfield.Operation(op))
		return fmt.Errorf("bucket id cannot be empty")
	}

	bucket, err := bs.bucketRepository.GetBucketById(ctx, id)
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

	err = bs.bucketRepository.DeleteBucket(ctx, bucket.Id)
	if err != nil {
		bs.logger.Error("failed to delete bucket", zap.Error(err), zapfield.Operation(op))
		return err
	}

	return nil
}

func (bs *BucketService) GetBucket(ctx context.Context, id string) (*dto.Bucket, error) {
	const op = "BucketService.GetBucket"

	bucket, err := bs.bucketRepository.GetBucketById(ctx, id)
	if err != nil {
		bs.logger.Error("failed to get bucket", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &dto.Bucket{
		Id:                   bucket.Id,
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

func (bs *BucketService) GetBucketSize(ctx context.Context, id string) (*dto.BucketSize, error) {
	const op = "BucketService.GetBucketSize"

	if validators.ValidateNotEmptyTrimmedString(id) {
		bs.logger.Error("bucket name cannot be empty", zapfield.Operation(op))
		return nil, fmt.Errorf("bucket id cannot be empty when getting bucket size")
	}

	bucketSize, err := bs.bucketRepository.GetBucketSizeById(ctx, id)
	if err != nil {
		bs.logger.Error("failed to get bucket size", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &dto.BucketSize{
		Id:   bucketSize.Id,
		Name: bucketSize.Name,
		Size: bucketSize.Size,
	}, nil
}

func (bs *BucketService) ListAllBuckets(ctx context.Context) ([]*dto.Bucket, error) {
	const op = "BucketService.ListAllBuckets"

	buckets, err := bs.bucketRepository.ListAllBuckets(ctx)
	if err != nil {
		bs.logger.Error("failed to list all buckets", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	var result []*dto.Bucket

	for _, bucket := range buckets {
		result = append(result, &dto.Bucket{
			Id:                   bucket.Id,
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
