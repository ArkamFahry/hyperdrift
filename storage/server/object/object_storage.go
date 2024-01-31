package object

import (
	"context"
	"errors"
	"fmt"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/common/zapfield"
	"github.com/ArkamFahry/hyperdrift/storage/server/object/dto"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type ObjectStorage struct {
	s3Client          *s3.Client
	s3PreSignedClient *s3.PresignClient
	bucketName        string
	config            *config.Config
	logger            *zap.Logger
}

func NewObjectStorage(s3Client *s3.Client, config *config.Config, logger *zap.Logger) *ObjectStorage {
	return &ObjectStorage{
		s3Client:          s3Client,
		s3PreSignedClient: s3.NewPresignClient(s3Client),
		bucketName:        config.S3BucketName,
		config:            config,
		logger:            logger,
	}
}

func (s *ObjectStorage) CreatePreSignedUploadObject(ctx context.Context, preSignedUploadObjectCreate *dto.PreSignedUploadObjectCreate) (*dto.PreSignedObject, error) {
	const op = "ObjectStorage.PreSignedUploadObjectCreate"

	var expiresIn time.Duration

	if preSignedUploadObjectCreate.ExpiresIn != nil {
		expiresIn = time.Duration(*preSignedUploadObjectCreate.ExpiresIn)
	} else {
		expiresIn = time.Duration(s.config.DefaultPreSignedUploadUrlExpiresIn)
	}

	key := createS3Key(preSignedUploadObjectCreate.Bucket, preSignedUploadObjectCreate.Name)

	preSignedPutObject, err := s.s3PreSignedClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(key),
		ContentLength: aws.Int64(preSignedUploadObjectCreate.Size),
		ContentType:   aws.String(preSignedUploadObjectCreate.ContentType),
	},

		func(po *s3.PresignOptions) {
			po.Expires = expiresIn
		},
	)
	if err != nil {
		s.logger.Error("failed to create pre-signed upload url", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &dto.PreSignedObject{
		Url:       preSignedPutObject.URL,
		Method:    "PUT",
		ExpiresAt: time.Now().Add(expiresIn).Unix(),
	}, nil
}

func (s *ObjectStorage) CreatePreSignedDownloadObject(ctx context.Context, preSignedDownloadObjectCreate *dto.PreSignedDownloadObjectCreate) (*dto.PreSignedObject, error) {
	const op = "ObjectStorage.PreSignedDownloadObjectCreate"

	var expiresIn time.Duration

	if preSignedDownloadObjectCreate.ExpiresIn != nil {
		expiresIn = time.Duration(*preSignedDownloadObjectCreate.ExpiresIn)
	} else {
		expiresIn = time.Duration(s.config.DefaultPreSignedDownloadUrlExpiresIn)
	}

	key := createS3Key(preSignedDownloadObjectCreate.Bucket, preSignedDownloadObjectCreate.Name)

	preSignedGetObject, err := s.s3PreSignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	},
		func(po *s3.PresignOptions) {
			po.Expires = expiresIn
		},
	)
	if err != nil {
		s.logger.Error("failed to create pre-signed download url", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &dto.PreSignedObject{
		Url:       preSignedGetObject.URL,
		Method:    "GET",
		ExpiresAt: time.Now().Add(expiresIn).Unix(),
	}, nil
}

func (s *ObjectStorage) CheckIfObjectExists(ctx context.Context, objectExistsCheck *dto.ObjectExistsCheck) (bool, error) {
	const op = "ObjectStorage.ObjectExistsCheck"

	key := createS3Key(objectExistsCheck.Bucket, objectExistsCheck.Name)

	_, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}
		s.logger.Error("failed to check if object exists", zap.Error(err), zapfield.Operation(op))
		return false, err
	}

	return true, nil
}

func (s *ObjectStorage) DeleteObject(ctx context.Context, objectDelete *dto.ObjectDelete) error {
	const op = "ObjectStorage.ObjectDelete"

	key := createS3Key(objectDelete.Bucket, objectDelete.Name)

	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		s.logger.Error("failed to delete object", zap.Error(err), zapfield.Operation(op))
		return err
	}

	return nil
}

func createS3Key(bucket string, name string) string {
	return fmt.Sprintf(`%s/%s`, bucket, name)
}
