package bucket

import (
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BucketRepository struct {
	query       *database.Queries
	transaction *database.Transaction
}

func NewBucketRepository(db *pgxpool.Pool) *BucketRepository {
	return &BucketRepository{
		query:       database.New(db),
		transaction: database.NewTransaction(db),
	}
}
