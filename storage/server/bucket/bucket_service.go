package bucket

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"regexp"
	"strings"
)

type Service struct {
	database    *database.Queries
	transaction *database.Transaction
	logger      *zap.Logger
}

func NewService(db *pgxpool.Pool, logger *zap.Logger) *Service {
	return &Service{
		database:    database.New(db),
		transaction: database.NewTransaction(db),
		logger:      logger,
	}
}

func (bs *Service) CreateBucket(ctx context.Context, bucketCreate *dto.BucketCreate) error {
	const op = "services.BucketService.CreateBucket"

	if bucketCreate.AllowedContentTypes != nil {
		err := validateAllowedContentTypes(bucketCreate.AllowedContentTypes)
		if err != nil {
			bs.logger.Error("failed to validate mime types", zap.Error(err), zap.String("operation", op))
			return err
		}
	}

	if bucketCreate.MaxAllowedObjectSize != nil {
		err := validateMaxAllowedObjectSize(*bucketCreate.MaxAllowedObjectSize)
		if err != nil {
			bs.logger.Error("failed to validate max allowed object size", zap.Error(err), zap.String("operation", op))
			return err
		}
	}

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

		return nil
	})
	if err != nil {
		bs.logger.Error("failed to create bucket", zap.Error(err), zap.String("operation", op))
		return err
	}

	return nil
}

func validateAllowedContentTypes(mimeTypes []string) error {
	var invalidContentTypes []string
	for _, mimeType := range mimeTypes {
		if !validateContentType(mimeType) {
			invalidContentTypes = append(invalidContentTypes, mimeType)
		}
	}

	if len(invalidContentTypes) > 0 {
		return fmt.Errorf("invalid content types: [%s]", strings.Join(invalidContentTypes, ", "))
	}

	return nil
}

func validateMaxAllowedObjectSize(maxAllowedObjectSize int64) error {
	if maxAllowedObjectSize < 0 {
		return fmt.Errorf("max allowed object size must be 0 or greater than 0")
	}

	return nil
}

func validateContentType(mimeType string) bool {
	mimeTypePattern := `^[a-zA-Z]+/[a-zA-Z0-9\-\.\+]+$`

	re := regexp.MustCompile(mimeTypePattern)

	return re.MatchString(mimeType)
}
