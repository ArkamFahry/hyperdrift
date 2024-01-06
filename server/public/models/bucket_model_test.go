package models

import (
	"reflect"
	"strings"
	"testing"
)

func TestNewCreateBucket(t *testing.T) {
	name := "TestBucket"
	allowedObjectSize := int64(100)

	bucketWithoutMimetype := NewCreateBucket(name, nil, allowedObjectSize)

	if bucketWithoutMimetype.Id == "" {
		t.Error("ID not generated")
	}

	if bucketWithoutMimetype.Name != name {
		t.Errorf("Expected name %s, got %s", name, bucketWithoutMimetype.Name)
	}

	if !reflect.DeepEqual(bucketWithoutMimetype.AllowedMimeTypes, []string{"*/*"}) {
		t.Errorf("Expected default MIME types, got %v", bucketWithoutMimetype.AllowedMimeTypes)
	}

	if bucketWithoutMimetype.AllowedObjectSize != allowedObjectSize {
		t.Errorf("Expected allowedObjectSize %d, got %d", allowedObjectSize, bucketWithoutMimetype.AllowedObjectSize)
	}

	if bucketWithoutMimetype.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}

	bucketWithMimetype := NewCreateBucket(name, []string{"image/jpeg"}, allowedObjectSize)

	if bucketWithMimetype.Id == "" {
		t.Error("ID not generated")
	}

	if bucketWithMimetype.Name != name {
		t.Errorf("Expected name %s, got %s", name, bucketWithMimetype.Name)
	}

	if !reflect.DeepEqual(bucketWithMimetype.AllowedMimeTypes, []string{"image/jpeg"}) {
		t.Errorf("Expected default MIME types, got %v", bucketWithMimetype.AllowedMimeTypes)
	}

	if bucketWithMimetype.AllowedObjectSize != allowedObjectSize {
		t.Errorf("Expected allowedObjectSize %d, got %d", allowedObjectSize, bucketWithMimetype.AllowedObjectSize)
	}

	if bucketWithMimetype.CreatedAt.IsZero() {
		t.Error("CreatedAt not set")
	}
}

func TestCreateBucketValidation(t *testing.T) {
	validBucket := NewCreateBucket("valid_bucket", []string{"image/jpeg"}, 1024)
	if err := validBucket.Validate(); err != nil {
		t.Errorf("Expected validation to pass for valid bucket, but got error: %v", err)
	}

	emptyNameBucket := NewCreateBucket("", []string{"image/jpeg"}, 1024)
	expectedError := "name is required"
	if err := emptyNameBucket.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for empty name, but got: %v", expectedError, err)
	}

	whiteSpaceNameBucket := NewCreateBucket("bucket with space", []string{"image/jpeg"}, 1024)
	expectedError = "name should not contain any white spaces or tabs"
	if err := whiteSpaceNameBucket.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for name with white spaces, but got: %v", expectedError, err)
	}

	invalidNameCharactersBucket := NewCreateBucket("bucket@invalid", []string{"image/jpeg"}, 1024)
	expectedError = "name should only contain letters, numbers, hyphens and underscores"
	if err := invalidNameCharactersBucket.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for name with invalid characters, but got: %v", expectedError, err)
	}

	invalidMimeTypesBucket := NewCreateBucket("valid_bucket", []string{"invalid/mime&type"}, 1024)
	expectedError = `not allowed mime type "invalid/mime&type"`
	if err := invalidMimeTypesBucket.Validate(); err == nil || !strings.Contains(err.Error(), expectedError) {
		t.Errorf("Expected '%s' error for invalid MIME type, but got: %v", expectedError, err)
	}
}
