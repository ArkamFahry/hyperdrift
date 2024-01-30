package bucket

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type IStorage interface {
	EmptyBucket(ctx context.Context, emptyBucket *dto.BucketEmpty) error
}

type S3Storage struct {
	s3Client   *s3.Client
	bucketName string
	config     *config.Config
	logger     *zap.Logger
}

var _ IStorage = (*S3Storage)(nil)

func NewS3Storage(s3Client *s3.Client, config *config.Config, logger *zap.Logger) IStorage {
	return &S3Storage{
		s3Client:   s3Client,
		bucketName: config.S3BucketName,
		config:     config,
		logger:     logger,
	}
}

func (s *S3Storage) EmptyBucket(ctx context.Context, emptyBucket *dto.BucketEmpty) error {
	const op = "bucket_storage.EmptyBucket"

	key := createS3Key(emptyBucket.Id)

	_, err := s.s3Client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(key),
	})
	if err != nil {
		s.logger.Error("failed to list objects", zap.Error(err), zap.String("bucket", s.bucketName), zap.String("key", key), zap.String("operation", op))
		return err
	}

	return nil
}

func createS3Key(bucket string) string {
	return fmt.Sprintf(`%s/`, bucket)
}
