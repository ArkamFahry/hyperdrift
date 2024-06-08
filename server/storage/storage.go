package storage

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/driftdev/storage/server/config"
	"github.com/driftdev/storage/server/zapfield"
	"go.uber.org/zap"
)

type Storage struct {
	s3Client          *s3.Client
	s3PreSignedClient *s3.PresignClient
	bucket            string
	config            *config.Config
	logger            *zap.Logger
}

func NewStorage(s3Client *s3.Client, config *config.Config, logger *zap.Logger) *Storage {
	return &Storage{
		s3Client:          s3Client,
		s3PreSignedClient: s3.NewPresignClient(s3Client),
		bucket:            config.S3Bucket,
		config:            config,
		logger:            logger,
	}
}

func (s *Storage) UploadObject(ctx context.Context, objectUpload *ObjectUpload) error {
	const op = "Storage.UploadObject"

	key := createS3Key(objectUpload.Bucket, objectUpload.Name)

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(objectUpload.ContentType),
		Body:        objectUpload.Content,
	})
	if err != nil {
		s.logger.Error("failed to put object", zap.Error(err), zapfield.Operation(op))
		return err
	}

	return nil
}

func (s *Storage) CreatePreSignedUploadObject(ctx context.Context, preSignedUploadObjectCreate *PreSignedUploadObjectCreate) (*PreSignedObject, error) {
	const op = "Storage.CreatePreSignedUploadObject"

	var expiresIn time.Duration

	if preSignedUploadObjectCreate.ExpiresIn != nil {
		expiresIn = time.Duration(*preSignedUploadObjectCreate.ExpiresIn) * time.Second
	} else {
		expiresIn = time.Duration(s.config.DefaultPreSignedUploadUrlExpiry) * time.Second
	}

	key := createS3Key(preSignedUploadObjectCreate.Bucket, preSignedUploadObjectCreate.Name)

	preSignedPutObject, err := s.s3PreSignedClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(preSignedUploadObjectCreate.ContentType),
		ContentLength: aws.Int64(preSignedUploadObjectCreate.ContentLength),
	},
		s3.WithPresignExpires(expiresIn),
	)
	if err != nil {
		s.logger.Error("failed to create pre-signed put object", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &PreSignedObject{
		Url:       preSignedPutObject.URL,
		Method:    preSignedPutObject.Method,
		ExpiresAt: time.Now().Unix() + int64(expiresIn.Seconds()),
	}, nil
}

func (s *Storage) CreatePreSignedDownloadObject(ctx context.Context, preSignedDownloadObjectCreate *PreSignedDownloadObjectCreate) (*PreSignedObject, error) {
	const op = "Storage.CreatePreSignedDownloadObject"

	var expiresIn time.Duration

	if preSignedDownloadObjectCreate.ExpiresIn != nil {
		expiresIn = time.Duration(*preSignedDownloadObjectCreate.ExpiresIn) * time.Second
	} else {
		expiresIn = time.Duration(s.config.DefaultPreSignedDownloadUrlExpiry) * time.Second
	}

	key := createS3Key(preSignedDownloadObjectCreate.Bucket, preSignedDownloadObjectCreate.Name)

	preSignedGetObject, err := s.s3PreSignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	},
		s3.WithPresignExpires(expiresIn),
	)
	if err != nil {
		s.logger.Error("failed to create pre-signed get object", zap.Error(err), zapfield.Operation(op))
		return nil, err
	}

	return &PreSignedObject{
		Url:       preSignedGetObject.URL,
		Method:    preSignedGetObject.Method,
		ExpiresAt: time.Now().Unix() + int64(expiresIn.Seconds()),
	}, nil
}

func (s *Storage) CheckIfObjectExists(ctx context.Context, objectExistsCheck *ObjectExistsCheck) (bool, error) {
	const op = "Storage.ObjectExistsCheck"

	key := createS3Key(objectExistsCheck.Bucket, objectExistsCheck.Name)

	_, err := s.s3Client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var responseError *awshttp.ResponseError
		if errors.As(err, &responseError) && responseError.ResponseError.HTTPStatusCode() == http.StatusNotFound {
			return false, nil
		}
		s.logger.Error("failed to head object", zap.Error(err), zapfield.Operation(op))
		return false, err
	}

	return true, nil
}

func (s *Storage) DeleteObject(ctx context.Context, objectDelete *ObjectDelete) error {
	const op = "Storage.DeleteObject"

	key := createS3Key(objectDelete.Bucket, objectDelete.Name)

	_, err := s.s3Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
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
