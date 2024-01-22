package storage

import (
	"context"
	"github.com/ArkamFahry/hyperdrift/storage/server/config"
	"github.com/ArkamFahry/hyperdrift/storage/server/models"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
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

func (s *S3Storage) CreatePreSignedUploadUrl(ctx context.Context, createPreSignedUploadUrl *models.CreatePreSignedUploadUrl) (*models.PreSignedUrl, error) {
	var expiresIn time.Duration

	if createPreSignedUploadUrl.ExpiresIn != nil {
		expiresIn = time.Duration(*createPreSignedUploadUrl.ExpiresIn)
	} else {
		expiresIn = time.Duration(s.config.DefaultPreSignedUploadUrlExpiresIn)
	}

	preSignedUploadUrl, err := s.s3PreSignedClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucketName),
		Key:           aws.String(createPreSignedUploadUrl.Path),
		ContentLength: aws.Int64(createPreSignedUploadUrl.Size),
		ContentType:   aws.String(createPreSignedUploadUrl.MimeType),
	},

		func(po *s3.PresignOptions) {
			po.Expires = expiresIn
		},
	)
	if err != nil {
		return nil, err
	}

	return &models.PreSignedUrl{
		Url:       preSignedUploadUrl.URL,
		Method:    "PUT",
		ExpiresAt: time.Now().Add(expiresIn).Unix(),
	}, nil
}

func (s *S3Storage) CreatePreSignedDownloadUrl(ctx context.Context, createPreSignedDownloadUrl *models.CreatePreSignedDownloadUrl) (*models.PreSignedUrl, error) {
	var expiresIn time.Duration

	if createPreSignedDownloadUrl.ExpiresIn != nil {
		expiresIn = time.Duration(*createPreSignedDownloadUrl.ExpiresIn)
	} else {
		expiresIn = time.Duration(s.config.DefaultPreSignedDownloadUrlExpiresIn)
	}

	preSignedDownloadUrl, err := s.s3PreSignedClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(createPreSignedDownloadUrl.Path),
	},
		func(po *s3.PresignOptions) {
			po.Expires = expiresIn
		},
	)
	if err != nil {
		return nil, err
	}

	return &models.PreSignedUrl{
		Url:       preSignedDownloadUrl.URL,
		Method:    "GET",
		ExpiresAt: time.Now().Add(expiresIn).Unix(),
	}, nil
}
