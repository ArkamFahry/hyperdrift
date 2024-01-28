package services

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/database"
	"github.com/ArkamFahry/hyperdrift/storage/server/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type IBucketService interface {
}

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

func (bs *BucketService) CreateBucket(ctx context.Context, bucketCreate *models.BucketCreate) error {
	const op = "services.BucketService.CreateBucket"

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
			bs.logger.Error("failed to create bucket", zap.Error(err), zap.String("operation", op))
			return err
		}

		bucketCreateEvent := bucketCreate.ToEvent()

		bucketCreateEventContent, err := bucketCreateEvent.ContentToByte()
		if err != nil {
			bs.logger.Error("failed to convert bucket create event content to byte", zap.Error(err), zap.String("operation", op))
			return err
		}

		err = bs.database.WithTx(tx).CreateEvent(ctx, &database.CreateEventParams{
			ID:        bucketCreateEvent.Id,
			Name:      bucketCreateEvent.Name,
			Content:   bucketCreateEventContent,
			Status:    bucketCreateEvent.Status,
			Retries:   bucketCreateEvent.Retries,
			ExpiresAt: bucketCreateEvent.ExpiresAt,
			CreatedAt: bucketCreateEvent.CreatedAt,
		})
		if err != nil {
			bs.logger.Error("failed to create bucket create event", zap.Error(err), zap.String("operation", op))
			return err
		}

		return nil
	})
	if err != nil {
		bs.logger.Error("failed to create bucket", zap.Error(err))
		return err
	}

	return nil
}
