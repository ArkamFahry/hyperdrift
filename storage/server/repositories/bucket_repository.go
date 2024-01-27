package repositories

import (
	"context"
	"errors"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/models"
	"github.com/jackc/pgx/v5"
)

type IBucketRepository interface {
	CreateBucket(ctx context.Context, createBucket *models.CreateBucket) error
	UpdateBucket(ctx context.Context, updateBucket *models.UpdateBucket) error
	AddAllowedContentTypeToBucket(ctx context.Context, addAllowedContentTypeToBucket *models.AddAllowedContentTypesToBucket) error
	RemoveAllowedContentTypeFromBucket(ctx context.Context, removeAllowedContentTypeFromBucket *models.RemoveAllowedContentTypesFromBucket) error
	MakeBucketPublic(ctx context.Context, makeBucketPublic *models.MakeBucketPublic) error
	MakeBucketPrivate(ctx context.Context, makeBucketPrivate *models.MakeBucketPrivate) error
	LockBucket(ctx context.Context, lockBucket *models.LockBucket) error
	UnlockBucket(ctx context.Context, unlockBucket *models.UnlockBucket) error
	DeleteBucket(ctx context.Context, deleteBucket *models.DeleteBucket) error
	GetBucketById(ctx context.Context, id string) (*models.Bucket, bool, error)
	GetBucketByName(ctx context.Context, name string) (*models.Bucket, bool, error)
	ListAllBuckets(ctx context.Context) ([]*models.Bucket, bool, error)
	ListBucketsPaginated(ctx context.Context, pagination *models.Pagination) ([]*models.Bucket, *models.PaginationResult, bool, error)
}

type BucketRepository struct {
	db *database.Queries
}

func NewBucketRepository(db *database.Queries) IBucketRepository {
	return &BucketRepository{
		db: db,
	}
}

func (br *BucketRepository) CreateBucket(ctx context.Context, createBucket *models.CreateBucket) error {
	err := br.db.CreateBucket(ctx, &database.CreateBucketParams{
		ID:                   createBucket.Id,
		Name:                 createBucket.Name,
		AllowedContentTypes:  createBucket.AllowedContentTypes,
		MaxAllowedObjectSize: createBucket.MaxAllowedObjectSize,
		Public:               createBucket.Public,
		Disabled:             createBucket.Disabled,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) UpdateBucket(ctx context.Context, updateBucket *models.UpdateBucket) error {
	err := br.db.UpdateBucket(ctx, &database.UpdateBucketParams{
		ID:                   updateBucket.Id,
		MaxAllowedObjectSize: updateBucket.MaxAllowedObjectSize,
		Public:               updateBucket.Public,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) AddAllowedContentTypeToBucket(ctx context.Context, addAllowedContentTypeToBucket *models.AddAllowedContentTypesToBucket) error {
	err := br.db.AddAllowedContentTypesToBucket(ctx, &database.AddAllowedContentTypesToBucketParams{
		ID:                  addAllowedContentTypeToBucket.Id,
		AllowedContentTypes: addAllowedContentTypeToBucket.AllowedContentTypes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) RemoveAllowedContentTypeFromBucket(ctx context.Context, removeAllowedContentTypeFromBucket *models.RemoveAllowedContentTypesFromBucket) error {
	err := br.db.RemoveAllowedContentTypesFromBucket(ctx, &database.RemoveAllowedContentTypesFromBucketParams{
		ID:                  removeAllowedContentTypeFromBucket.Id,
		AllowedContentTypes: removeAllowedContentTypeFromBucket.AllowedContentTypes,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) MakeBucketPublic(ctx context.Context, makeBucketPublic *models.MakeBucketPublic) error {
	err := br.db.MakeBucketPublic(ctx, makeBucketPublic.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) MakeBucketPrivate(ctx context.Context, makeBucketPrivate *models.MakeBucketPrivate) error {
	err := br.db.MakeBucketPrivate(ctx, makeBucketPrivate.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) LockBucket(ctx context.Context, lockBucket *models.LockBucket) error {
	err := br.db.LockBucket(ctx, &database.LockBucketParams{
		ID:         lockBucket.Id,
		LockReason: lockBucket.LockReason,
	})
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) UnlockBucket(ctx context.Context, unlockBucket *models.UnlockBucket) error {
	err := br.db.UnlockBucket(ctx, unlockBucket.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) DeleteBucket(ctx context.Context, deleteBucket *models.DeleteBucket) error {
	err := br.db.DeleteBucket(ctx, deleteBucket.Id)
	if err != nil {
		return err
	}

	return nil
}

func (br *BucketRepository) GetBucketById(ctx context.Context, id string) (*models.Bucket, bool, error) {
	bucket, err := br.db.GetBucketById(ctx, id)
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
	bucket, err := br.db.GetBucketByName(ctx, name)
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
	buckets, err := br.db.ListAllBuckets(ctx)
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

	buckets, err := br.db.ListBucketsPaginated(ctx, &database.ListBucketsPaginatedParams{
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
