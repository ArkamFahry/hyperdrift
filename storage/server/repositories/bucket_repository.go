package repositories

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/models"
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
	GetBucketById(ctx context.Context, id string) (*models.Bucket, error)
	GetBucketByName(ctx context.Context, name string) (*models.Bucket, error)
	ListAllBuckets(ctx context.Context) ([]*models.Bucket, error)
	ListBucketsPaginated(ctx context.Context, pagination *models.Pagination) ([]*models.Bucket, *models.PaginationResult, error)
}
