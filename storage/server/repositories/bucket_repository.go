package repositories

import (
	"context"
	"errors"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/models"
	"github.com/jackc/pgx/v5"
)

type IBucketRepository interface {
	CreateBucket(ctx context.Context, createBucket *models.BucketCreate) error
	UpdateBucket(ctx context.Context, updateBucket *models.BucketUpdate) error
	AddAllowedContentTypeToBucket(ctx context.Context, addAllowedContentTypeToBucket *models.BucketAddAllowedContentTypes) error
	RemoveAllowedContentTypeFromBucket(ctx context.Context, removeAllowedContentTypeFromBucket *models.BucketRemoveAllowedContentTypes) error
	MakeBucketPublic(ctx context.Context, makeBucketPublic *models.BucketMakePublic) error
	MakeBucketPrivate(ctx context.Context, makeBucketPrivate *models.BucketMakePrivate) error
	LockBucket(ctx context.Context, lockBucket *models.BucketLock) error
	UnlockBucket(ctx context.Context, unlockBucket *models.BucketUnlock) error
	DeleteBucket(ctx context.Context, deleteBucket *models.BucketDelete) error
	GetBucketById(ctx context.Context, id string) (*models.Bucket, bool, error)
	GetBucketByName(ctx context.Context, name string) (*models.Bucket, bool, error)
	ListAllBuckets(ctx context.Context) ([]*models.Bucket, bool, error)
	ListBucketsPaginated(ctx context.Context, pagination *models.Pagination) ([]*models.Bucket, *models.PaginationResult, bool, error)
}

type BucketRepository struct {
	queries     *database.Queries
	transaction *database.Transaction
}

func NewBucketRepository(db *database.Queries) IBucketRepository {
	return &BucketRepository{
		queries: db,
	}
}

func (br *BucketRepository) CreateBucket(ctx context.Context, bucketCreate *models.BucketCreate) error {
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

func (br *BucketRepository) UpdateBucket(ctx context.Context, bucketUpdate *models.BucketUpdate) error {
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

func (br *BucketRepository) AddAllowedContentTypeToBucket(ctx context.Context, bucketAddAllowedContentType *models.BucketAddAllowedContentTypes) error {
	err := br.queries.AddAllowedContentTypesToBucket(ctx, &database.AddAllowedContentTypesToBucketParams{
		ID:                  bucketAddAllowedContentType.Id,
		AllowedContentTypes: bucketAddAllowedContentType.AllowedContentTypes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) RemoveAllowedContentTypeFromBucket(ctx context.Context, bucketRemoveAllowedContentType *models.BucketRemoveAllowedContentTypes) error {
	err := br.queries.RemoveAllowedContentTypesFromBucket(ctx, &database.RemoveAllowedContentTypesFromBucketParams{
		ID:                  bucketRemoveAllowedContentType.Id,
		AllowedContentTypes: bucketRemoveAllowedContentType.AllowedContentTypes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) MakeBucketPublic(ctx context.Context, bucketMakePublic *models.BucketMakePublic) error {
	err := br.queries.MakeBucketPublic(ctx, bucketMakePublic.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) MakeBucketPrivate(ctx context.Context, bucketMakePrivate *models.BucketMakePrivate) error {
	err := br.queries.MakeBucketPrivate(ctx, bucketMakePrivate.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) LockBucket(ctx context.Context, bucketLock *models.BucketLock) error {
	err := br.queries.LockBucket(ctx, &database.LockBucketParams{
		ID:         bucketLock.Id,
		LockReason: bucketLock.LockReason,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) UnlockBucket(ctx context.Context, bucketUnlock *models.BucketUnlock) error {
	err := br.queries.UnlockBucket(ctx, bucketUnlock.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) DeleteBucket(ctx context.Context, bucketDelete *models.BucketDelete) error {
	err := br.queries.DeleteBucket(ctx, bucketDelete.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) GetBucketById(ctx context.Context, id string) (*models.Bucket, bool, error) {
	bucket, err := br.queries.GetBucketById(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &models.Bucket{
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

func (br *BucketRepository) GetBucketByName(ctx context.Context, name string) (*models.Bucket, bool, error) {
	bucket, err := br.queries.GetBucketByName(ctx, name)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	return &models.Bucket{
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

func (br *BucketRepository) ListAllBuckets(ctx context.Context) ([]*models.Bucket, bool, error) {
	buckets, err := br.queries.ListAllBuckets(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, err
	}

	var bucketModels []*models.Bucket

	for _, bucket := range buckets {
		bucketModels = append(bucketModels, &models.Bucket{
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

func (br *BucketRepository) ListBucketsPaginated(ctx context.Context, pagination *models.Pagination) ([]*models.Bucket, *models.PaginationResult, bool, error) {
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

	var bucketModels []*models.Bucket

	for _, bucket := range buckets {
		bucketModels = append(bucketModels, &models.Bucket{
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
