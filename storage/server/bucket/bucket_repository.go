package bucket

import (
	"context"
	"errors"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/dtos"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/jackc/pgx/v5"
)

type IBucketRepository interface {
	CreateBucket(ctx context.Context, createBucket *dto.BucketCreate) error
	UpdateBucket(ctx context.Context, updateBucket *dto.BucketUpdate) error
	AddAllowedContentTypeToBucket(ctx context.Context, addAllowedContentTypeToBucket *dto.BucketAddAllowedContentTypes) error
	RemoveAllowedContentTypeFromBucket(ctx context.Context, removeAllowedContentTypeFromBucket *dto.BucketRemoveAllowedContentTypes) error
	MakeBucketPublic(ctx context.Context, makeBucketPublic *dto.BucketMakePublic) error
	MakeBucketPrivate(ctx context.Context, makeBucketPrivate *dto.BucketMakePrivate) error
	LockBucket(ctx context.Context, lockBucket *dto.BucketLock) error
	UnlockBucket(ctx context.Context, unlockBucket *dto.BucketUnlock) error
	DeleteBucket(ctx context.Context, deleteBucket *dto.BucketDelete) error
	GetBucketById(ctx context.Context, id string) (*Bucket, bool, error)
	GetBucketByName(ctx context.Context, name string) (*Bucket, bool, error)
	ListAllBuckets(ctx context.Context) ([]*Bucket, bool, error)
	ListBucketsPaginated(ctx context.Context, pagination *dtos.Pagination) ([]*Bucket, *dtos.PaginationResult, bool, error)
}

type Repository struct {
	queries     *database.Queries
	transaction *database.Transaction
}

func NewBucketRepository(db *database.Queries) IBucketRepository {
	return &Repository{
		queries: db,
	}
}

func (br *Repository) CreateBucket(ctx context.Context, bucketCreate *dto.BucketCreate) error {
	err := br.queries.CreateBucket(ctx, &database.CreateBucketParams{
		ID:                   bucketCreate.Id,
		Name:                 bucketCreate.Name,
		AllowedContentTypes:  bucketCreate.AllowedContentTypes,
		MaxAllowedObjectSize: bucketCreate.MaxAllowedObjectSize,
		Public:               bucketCreate.Public,
		Disabled:             bucketCreate.Disabled,
	})
	if err != nil {
		return err
	}
	return nil
}

func (br *Repository) UpdateBucket(ctx context.Context, bucketUpdate *dto.BucketUpdate) error {
	err := br.queries.UpdateBucket(ctx, &database.UpdateBucketParams{
		ID:                   bucketUpdate.Id,
		MaxAllowedObjectSize: bucketUpdate.MaxAllowedObjectSize,
		Public:               bucketUpdate.Public,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *Repository) AddAllowedContentTypeToBucket(ctx context.Context, bucketAddAllowedContentType *dto.BucketAddAllowedContentTypes) error {
	err := br.queries.AddAllowedContentTypesToBucket(ctx, &database.AddAllowedContentTypesToBucketParams{
		ID:                  bucketAddAllowedContentType.Id,
		AllowedContentTypes: bucketAddAllowedContentType.AllowedContentTypes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *Repository) RemoveAllowedContentTypeFromBucket(ctx context.Context, bucketRemoveAllowedContentType *dto.BucketRemoveAllowedContentTypes) error {
	err := br.queries.RemoveAllowedContentTypesFromBucket(ctx, &database.RemoveAllowedContentTypesFromBucketParams{
		ID:                  bucketRemoveAllowedContentType.Id,
		AllowedContentTypes: bucketRemoveAllowedContentType.AllowedContentTypes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *Repository) MakeBucketPublic(ctx context.Context, bucketMakePublic *dto.BucketMakePublic) error {
	err := br.queries.MakeBucketPublic(ctx, bucketMakePublic.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *Repository) MakeBucketPrivate(ctx context.Context, bucketMakePrivate *dto.BucketMakePrivate) error {
	err := br.queries.MakeBucketPrivate(ctx, bucketMakePrivate.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *Repository) LockBucket(ctx context.Context, bucketLock *dto.BucketLock) error {
	err := br.queries.LockBucket(ctx, &database.LockBucketParams{
		ID:         bucketLock.Id,
		LockReason: bucketLock.LockReason,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *Repository) UnlockBucket(ctx context.Context, bucketUnlock *dto.BucketUnlock) error {
	err := br.queries.UnlockBucket(ctx, bucketUnlock.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *Repository) DeleteBucket(ctx context.Context, bucketDelete *dto.BucketDelete) error {
	err := br.queries.DeleteBucket(ctx, bucketDelete.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *Repository) GetBucketById(ctx context.Context, id string) (*Bucket, bool, error) {
	bucket, err := br.queries.GetBucketById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &Bucket{
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
	}, true, nil
}

func (br *Repository) GetBucketByName(ctx context.Context, name string) (*Bucket, bool, error) {
	bucket, err := br.queries.GetBucketByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &Bucket{
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
	}, true, nil
}

func (br *Repository) ListAllBuckets(ctx context.Context) ([]*Bucket, bool, error) {
	buckets, err := br.queries.ListAllBuckets(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var bucketModels []*Bucket

	for _, bucket := range buckets {
		bucketModels = append(bucketModels, &Bucket{
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

	return bucketModels, true, nil
}

func (br *Repository) ListBucketsPaginated(ctx context.Context, pagination *dtos.Pagination) ([]*Bucket, *dtos.PaginationResult, bool, error) {
	pagination.SetDefaults()

	buckets, err := br.queries.ListBucketsPaginated(ctx, &database.ListBucketsPaginatedParams{
		Cursor: pagination.Cursor,
		Limit:  &pagination.Limit,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil, false, nil
		}
		return nil, nil, false, err
	}

	var bucketModels []*Bucket

	for _, bucket := range buckets {
		bucketModels = append(bucketModels, &Bucket{
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

	return bucketModels, nil, true, nil
}
