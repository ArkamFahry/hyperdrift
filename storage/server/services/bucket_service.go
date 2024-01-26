package services

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IBucketService interface {
}

type BucketService struct {
	database    *database.Queries
	transaction *database.Transaction
}

func NewBucketService(db *pgxpool.Pool) *BucketService {
	return &BucketService{
		database:    database.New(db),
		transaction: database.NewTransaction(db),
	}
}

func (s *BucketService) CreateBucket(ctx context.Context, createBucket models.CreateBucket) error {
	err := s.transaction.WithTransaction(ctx, func(q *database.Queries) error {
		err := q.CreateBucket(ctx, &database.CreateBucketParams{
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
	})

	return err
}
