package models

import (
	"strings"
	"testing"
)

func TestNewCreateObject(t *testing.T) {
	bucketId := "bucket123"
	bucketName := "testBucket"
	objectName := "testObject"
	mimeType := "image/jpeg"
	objectSize := int64(1024)

	createdObject := NewCreateObject(bucketId, bucketName, objectName, mimeType, objectSize)

	if createdObject.Id == "" {
		t.Error("ID not generated")
	}

	expectedObjectName := bucketName + "/" + objectName
	if createdObject.Name != expectedObjectName {
		t.Errorf("Expected name %s, got %s", expectedObjectName, createdObject.Name)
	}

	if createdObject.MimeType != mimeType {
		t.Errorf("Expected MIME type %s, got %s", mimeType, createdObject.MimeType)
	}

	if createdObject.ObjectSize != objectSize {
		t.Errorf("Expected object size %d, got %d", objectSize, createdObject.ObjectSize)
	}

	if createdObject.UploadStatus != ObjectUploadStatusPending {
		t.Errorf("Expected upload status %s, got %s", ObjectUploadStatusPending, createdObject.UploadStatus)
	}

	if createdObject.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
}

func TestCreateObjectValidation(t *testing.T) {
	validObject := NewCreateObject("bucket123", "testBucket", "testObject", "image/jpeg", 1024)
	if err := validObject.Validate(); err != nil {
		t.Errorf("Expected validation to pass for valid object, but got error: %v", err)
	}

	emptyBucketNameObject := NewCreateObject("bucket123", "", "testObject", "image/jpeg", 1024)
	expectedError := "bucket name cannot be empty"
	if err := emptyBucketNameObject.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for empty name, but got: %v", expectedError, err)
	}

	emptyNameObject := NewCreateObject("bucket123", "testBucket", "", "image/jpeg", 1024)
	expectedError = "object name cannot be empty"
	if err := emptyNameObject.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for empty name, but got: %v", expectedError, err)
	}

	whiteSpaceNameObject := NewCreateObject("bucket123", "testBucket", "test Object", "image/jpeg", 1024)
	expectedError = "name should not contain any white spaces or tabs"
	if err := whiteSpaceNameObject.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for name with white spaces, but got: %v", expectedError, err)
	}

	emptyMimeTypeObject := NewCreateObject("bucket123", "testBucket", "testObject", "", 1024)
	expectedError = "mime_type is required"
	if err := emptyMimeTypeObject.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for empty MIME type, but got: %v", expectedError, err)
	}

	invalidMimeTypeObject := NewCreateObject("bucket123", "testBucket", "testObject", "invalid/mime&type", 1024)
	expectedError = "invalid mime type"
	if err := invalidMimeTypeObject.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for invalid MIME type, but got: %v", expectedError, err)
	}

	negativeSizeObject := NewCreateObject("bucket123", "testBucket", "testObject", "image/jpeg", -1024)
	expectedError = "object_size should be greater than 0"
	if err := negativeSizeObject.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for negative object size, but got: %v", expectedError, err)
	}
}
