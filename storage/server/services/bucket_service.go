package services

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
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
