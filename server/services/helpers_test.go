package services

import (
	"fmt"
	"github.com/ArkamFahry/storage/server/models"
	"github.com/samber/lo"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetadataConversion(t *testing.T) {
	metadata := map[string]any{
		"user_id": "user_123456789",
		"user_name": map[string]any{
			"first_name": "John",
			"last_name":  "Doe",
		},
		"access_only_to": []any{"admin", "user"},
	}

	metadataBytes := metadataToBytes(metadata)
	assert.NotNil(t, metadataBytes, "Metadata to bytes conversion failed")

	deserializedMetadata := bytesToMetadata(metadataBytes)
	assert.NotNil(t, deserializedMetadata, "Bytes to metadata conversion failed")
	assert.Equal(t, metadata, deserializedMetadata, "Deserialized metadata does not match original")
}

func TestInvalidBytesToMetadata(t *testing.T) {
	invalidBytes := []byte("invalid JSON")
	deserializedMetadata := bytesToMetadata(invalidBytes)
	assert.Nil(t, deserializedMetadata, "Bytes to metadata conversion should fail with invalid JSON")
}

func TestDetermineMimeType(t *testing.T) {
	type test struct {
		name                         string
		bucket                       *models.Bucket
		preSignedUploadSessionCreate *models.PreSignedUploadSessionCreate
		expectedMimeType             *string
		expectedError                error
	}

	mimeTypeInferredByExtensionTest := test{
		name: "Default mime type",
		bucket: &models.Bucket{
			AllowedMimeTypes: []string{"*/*"},
		},
		preSignedUploadSessionCreate: &models.PreSignedUploadSessionCreate{
			BucketId: "bucket1",
			Name:     "user/david/avatar.jpg",
			MimeType: nil,
			Size:     1218077,
			Metadata: nil,
		},
		expectedMimeType: lo.ToPtr[string]("image/jpeg"),
		expectedError:    nil,
	}

	inferredMimeType, err := determineMimeType(mimeTypeInferredByExtensionTest.bucket, mimeTypeInferredByExtensionTest.preSignedUploadSessionCreate)
	assert.Equal(t, mimeTypeInferredByExtensionTest.expectedMimeType, inferredMimeType)
	assert.Equal(t, mimeTypeInferredByExtensionTest.expectedError, err)

	emptyMimeTypeTest := test{
		name: "Empty mime type",
		bucket: &models.Bucket{
			AllowedMimeTypes: []string{"image/jpeg"},
		},
		preSignedUploadSessionCreate: &models.PreSignedUploadSessionCreate{
			BucketId: "bucket1",
			Name:     "user/david/avatar.jpg",
			MimeType: nil,
			Size:     1218077,
			Metadata: nil,
		},
		expectedMimeType: nil,
		expectedError:    fmt.Errorf("mime_type cannot be empty. bucket only allows [image/jpeg] mime types. please specify an allowed mime type"),
	}

	_, err = determineMimeType(emptyMimeTypeTest.bucket, emptyMimeTypeTest.preSignedUploadSessionCreate)
	assert.Equal(t, emptyMimeTypeTest.expectedError, err)

	mimeTypeNotAllowedTest := test{
		name: "Invalid mime type",
		bucket: &models.Bucket{
			AllowedMimeTypes: []string{"image/jpeg"},
		},
		preSignedUploadSessionCreate: &models.PreSignedUploadSessionCreate{
			BucketId: "bucket1",
			Name:     "user/david/avatar.jpg",
			MimeType: lo.ToPtr[string]("image/png"),
			Size:     1218077,
			Metadata: nil,
		},
		expectedMimeType: nil,
		expectedError:    fmt.Errorf("mime_type 'image/png' is not allowed. bucket only allows [image/jpeg] mime types. please specify an allowed mime type"),
	}

	_, err = determineMimeType(mimeTypeNotAllowedTest.bucket, mimeTypeNotAllowedTest.preSignedUploadSessionCreate)
	assert.Equal(t, mimeTypeNotAllowedTest.expectedError, err)
}
