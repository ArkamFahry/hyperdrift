package bucket

import (
	"context"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/bucket/dto"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/zapfield"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type BucketStorage struct {
	s3Client   *s3.Client
	bucketName string
	config     *config.Config
	logger     *zap.Logger
}

func NewBucketStorage(s3Client *s3.Client, config *config.Config, logger *zap.Logger) *BucketStorage {
	return &BucketStorage{
		s3Client:   s3Client,
		bucketName: config.S3BucketName,
		config:     config,
		logger:     logger,
	}
}

func (s *BucketStorage) EmptyBucket(ctx context.Context, emptyBucket *dto.BucketEmpty) error {
	const op = "BucketStorage.EmptyBucket"

	key := createS3BucketPathKey(emptyBucket.Id)

	_, err := s.s3Client.ListObjects(ctx, &s3.ListObjectsInput{
		Bucket: aws.String(s.bucketName),
		Prefix: aws.String(key),
	})
	if err != nil {
		s.logger.Error(
			"failed to list objects",
			zap.Error(err),
			zap.String("bucket", s.bucketName),
			zap.String("key", key),
			zapfield.Operation(op),
		)
		return err
	}

	return nil
}

func createS3BucketPathKey(bucket string) string {
	return fmt.Sprintf(`%s/`, bucket)
}
