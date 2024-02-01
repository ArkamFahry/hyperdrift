package bucket

import (
	"context"
	"errors"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/entities"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/pagiantor"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type BucketRepository struct {
	query       *database.Queries
	transaction *database.Transaction
	logger      *zap.Logger
}

func NewBucketRepository(db *pgxpool.Pool) *BucketRepository {
	return &BucketRepository{
		query:       database.New(db),
		transaction: database.NewTransaction(db),
	}
}

func (r *BucketRepository) CreateBucket(ctx context.Context, bucketCreate *entities.BucketCreate) (*entities.Bucket, error) {
	err := r.query.CreateBucket(ctx, &database.CreateBucketParams{
		ID:                   bucketCreate.Id,
		Name:                 bucketCreate.Name,
		AllowedContentTypes:  bucketCreate.AllowedContentTypes,
		MaxAllowedObjectSize: bucketCreate.MaxAllowedObjectSize,
		Public:               bucketCreate.Public,
		Disabled:             bucketCreate.Disabled,
	})
	if err != nil {
		return nil, err
	}

	createdBucket, err := r.query.GetBucketById(ctx, bucketCreate.Id)
	if err != nil {
		return nil, err
	}

	return &entities.Bucket{
		Id:                   createdBucket.ID,
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

func (r *BucketRepository) EnableBucket(ctx context.Context, bucketEnable *entities.BucketEnable) (*entities.Bucket, error) {
	err := r.query.EnableBucket(ctx, &database.EnableBucketParams{
		ID:      bucketEnable.Id,
		Version: bucketEnable.Version,
	})
	if err != nil {
		return nil, err
	}

	enabledBucket, err := r.query.GetBucketById(ctx, bucketEnable.Id)
	if err != nil {
		return nil, err
	}

	return &entities.Bucket{
		Id:                   enabledBucket.ID,
		Version:              enabledBucket.Version,
		Name:                 enabledBucket.Name,
		AllowedContentTypes:  enabledBucket.AllowedContentTypes,
		MaxAllowedObjectSize: enabledBucket.MaxAllowedObjectSize,
		Public:               enabledBucket.Public,
		Disabled:             enabledBucket.Disabled,
		Locked:               enabledBucket.Locked,
		LockReason:           enabledBucket.LockReason,
		LockedAt:             enabledBucket.LockedAt,
		CreatedAt:            enabledBucket.CreatedAt,
		UpdatedAt:            enabledBucket.UpdatedAt,
	}, nil
}

func (r *BucketRepository) DisableBucket(ctx context.Context, bucketDisable *entities.BucketDisable) (*entities.Bucket, error) {
	err := r.query.DisableBucket(ctx, &database.DisableBucketParams{
		ID:      bucketDisable.Id,
		Version: bucketDisable.Version,
	})
	if err != nil {
		return nil, err
	}

	disabledBucket, err := r.query.GetBucketById(ctx, bucketDisable.Id)
	if err != nil {
		return nil, err
	}

	return &entities.Bucket{
		Id:                   disabledBucket.ID,
		Version:              disabledBucket.Version,
		Name:                 disabledBucket.Name,
		AllowedContentTypes:  disabledBucket.AllowedContentTypes,
		MaxAllowedObjectSize: disabledBucket.MaxAllowedObjectSize,
		Public:               disabledBucket.Public,
		Disabled:             disabledBucket.Disabled,
		Locked:               disabledBucket.Locked,
		LockReason:           disabledBucket.LockReason,
		LockedAt:             disabledBucket.LockedAt,
		CreatedAt:            disabledBucket.CreatedAt,
		UpdatedAt:            disabledBucket.UpdatedAt,
	}, nil
}

func (r *BucketRepository) UpdateBucketAllowedContentTypes(ctx context.Context, bucketAllowedContentTypesUpdate *entities.BucketAllowedContentTypesUpdate) (*entities.Bucket, error) {
	err := r.query.UpdateBucketAllowedContentTypes(ctx, &database.UpdateBucketAllowedContentTypesParams{
		ID:                  bucketAllowedContentTypesUpdate.Id,
		AllowedContentTypes: bucketAllowedContentTypesUpdate.AllowedContentTypes,
		Version:             bucketAllowedContentTypesUpdate.Version,
	})
	if err != nil {
		return nil, err
	}

	contentTypeRemovedBucket, err := r.query.GetBucketById(ctx, bucketAllowedContentTypesUpdate.Id)
	if err != nil {
		return nil, err
	}

	return &entities.Bucket{
		Id:                   contentTypeRemovedBucket.ID,
		Version:              contentTypeRemovedBucket.Version,
		Name:                 contentTypeRemovedBucket.Name,
		AllowedContentTypes:  contentTypeRemovedBucket.AllowedContentTypes,
		MaxAllowedObjectSize: contentTypeRemovedBucket.MaxAllowedObjectSize,
		Public:               contentTypeRemovedBucket.Public,
		Disabled:             contentTypeRemovedBucket.Disabled,
		Locked:               contentTypeRemovedBucket.Locked,
		LockReason:           contentTypeRemovedBucket.LockReason,
		LockedAt:             contentTypeRemovedBucket.LockedAt,
		CreatedAt:            contentTypeRemovedBucket.CreatedAt,
		UpdatedAt:            contentTypeRemovedBucket.UpdatedAt,
	}, nil
}

func (r *BucketRepository) UpdateBucket(ctx context.Context, bucketUpdate *entities.BucketUpdate) (*entities.Bucket, error) {
	err := r.query.UpdateBucket(ctx, &database.UpdateBucketParams{
		ID:                   bucketUpdate.Id,
		MaxAllowedObjectSize: bucketUpdate.MaxAllowedObjectSize,
		Public:               bucketUpdate.Public,
		Version:              bucketUpdate.Version,
	})
	if err != nil {
		return nil, err
	}

	updatedBucket, err := r.query.GetBucketById(ctx, bucketUpdate.Id)
	if err != nil {
		return nil, err
	}

	return &entities.Bucket{
		Id:                   updatedBucket.ID,
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

func (r *BucketRepository) DeleteBucket(ctx context.Context, id string) error {
	err := r.query.DeleteBucket(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func (r *BucketRepository) GetBucketById(ctx context.Context, id string) (*entities.Bucket, error) {
	bucket, err := r.query.GetBucketById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("bucket with id '%s' not found", id)
		}
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

func (r *BucketRepository) GetBucketByName(ctx context.Context, name string) (*entities.Bucket, error) {
	bucket, err := r.query.GetBucketByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("bucket with name '%s' not found", name)
		}
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

func (r *BucketRepository) GetBucketSizeById(ctx context.Context, id string) (*entities.BucketSize, error) {
	bucketSize, err := r.query.GetBucketSizeById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("size of bucket with id '%s' not found", id)
		}
		return nil, err
	}

	return &entities.BucketSize{
		Id:   bucketSize.ID,
		Name: bucketSize.Name,
		Size: bucketSize.Size,
	}, nil
}

func (r *BucketRepository) ListAllBuckets(ctx context.Context) ([]*entities.Bucket, error) {
	bucketList, err := r.query.ListAllBuckets(ctx)
	if err != nil {
		return nil, err
	}

	var result []*entities.Bucket

	for _, bucket := range bucketList {
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

func (r *BucketRepository) ListBucketsPaginated(ctx context.Context, paginationInput pagiantor.PaginationInput) ([]*entities.Bucket, error) {
	paginationInput.SetDefaults()

	bucketList, err := r.query.ListBucketsPaginated(ctx, &database.ListBucketsPaginatedParams{
		Cursor: paginationInput.Cursor,
		Limit:  paginationInput.Limit,
	})
	if err != nil {
		return nil, err
	}

	var result []*entities.Bucket

	for _, bucket := range bucketList {
		result = append(result, &entities.Bucket{
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
		})
	}

	return result, nil
}
