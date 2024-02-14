package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/ArkamFahry/storage/server/config"
	"github.com/ArkamFahry/storage/server/zapfield"
	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
	"net/http"
	"time"
)

type S3Storage struct {
	s3Client          *s3.Client
	s3PreSignedClient *s3.PresignClient
	bucketName        string
	config            *config.Config
	logger            *zap.Logger
}

func NewS3Storage(s3Client *s3.Client, config *config.Config, logger *zap.Logger) *S3Storage {
	return &S3Storage{
		s3Client:          s3Client,
		s3PreSignedClient: s3.NewPresignClient(s3Client),
		bucketName:        config.S3BucketName,
		config:            config,
		logger:            logger,
	}
}

func (s *S3Storage) UploadObject(ctx context.Context, objectUpload *ObjectUpload) error {

	key := createS3Key(objectUpload.Bucket, objectUpload.Name)

	_, err := s.s3Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(key),
		ContentType: aws.String(objectUpload.ContentType),
		Body:        objectUpload.Content,
	})
	if err != nil {
		return err
	}

	return nil
}

func (s *S3Storage) CreatePreSignedUploadObject(ctx context.Context, preSignedUploadObjectCreate *PreSignedUploadObjectCreate) (*PreSignedObject, error) {
	const op = "S3Storage.CreatePreSignedUploadObject"

	var expiresIn time.Duration

	if preSignedUploadObjectCreate.ExpiresIn != nil {
		expiresIn = time.Duration(*preSignedUploadObjectCreate.ExpiresIn) * time.Second
	} else {
		expiresIn = time.Duration(s.config.DefaultPreSignedUploadUrlExpiresIn) * time.Second
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

	return &PreSignedObject{
		Url:       preSignedPutObject.URL,
		Method:    preSignedPutObject.Method,
		ExpiresAt: time.Now().Unix() + int64(expiresIn.Seconds()),
	}, nil
}

func (s *S3Storage) CreatePreSignedDownloadObject(ctx context.Context, preSignedDownloadObjectCreate *PreSignedDownloadObjectCreate) (*PreSignedObject, error) {
	const op = "S3Storage.PreSignedDownloadObjectCreate"

	var expiresIn time.Duration

	if preSignedDownloadObjectCreate.ExpiresIn != nil {
		expiresIn = time.Duration(*preSignedDownloadObjectCreate.ExpiresIn) * time.Second
	} else {
		expiresIn = time.Duration(s.config.DefaultPreSignedDownloadUrlExpiresIn) * time.Second
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

	return &PreSignedObject{
		Url:       preSignedGetObject.URL,
		Method:    preSignedGetObject.Method,
		ExpiresAt: time.Now().Unix() + int64(expiresIn.Seconds()),
	}, nil
}

func (s *S3Storage) CheckIfObjectExists(ctx context.Context, objectExistsCheck *ObjectExistsCheck) (bool, error) {
	const op = "S3Storage.ObjectExistsCheck"

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

func (s *S3Storage) DeleteObject(ctx context.Context, objectDelete *ObjectDelete) error {
	const op = "S3Storage.ObjectDelete"

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
