package models

import (
	"reflect"
	"testing"
	"time"

	"github.com/ArkamFahry/hyperdrift-storage/server/packages/apperr"
)

func TestNewCreateBucket(t *testing.T) {
	t.Run("CreateBucketInstance", func(t *testing.T) {
		name := "myBucket"
		allowedMimeTypes := []string{"image/png", "image/jpeg"}
		allowedObjectSize := int64(1024)
		createdAt := time.Now()

		expectedBucket := &CreateBucket{
			Id:                "bucket_someUniqueId",
			Name:              name,
			AllowedMimeTypes:  allowedMimeTypes,
			AllowedObjectSize: allowedObjectSize,
			CreatedAt:         createdAt,
		}

		newBucket := NewCreateBucket(name, allowedMimeTypes, allowedObjectSize)

		if newBucket.Name != expectedBucket.Name ||
			!reflect.DeepEqual(newBucket.AllowedMimeTypes, expectedBucket.AllowedMimeTypes) ||
			newBucket.AllowedObjectSize != expectedBucket.AllowedObjectSize {
			t.Errorf("Generated bucket does not match the expected bucket.\nExpected: %+v\nGot: %+v", expectedBucket, newBucket)
		}
	})
}

func TestCreateBucketValidation(t *testing.T) {
	t.Run("MissingId", func(t *testing.T) {
		bucketMissingID := CreateBucket{
			Name:              "myBucket",
			AllowedMimeTypes:  []string{"image/png", "image/jpeg"},
			AllowedObjectSize: 1024,
			CreatedAt:         time.Now(),
		}

		err := bucketMissingID.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(*apperr.FieldError)
		if !ok {
			t.Error("Expected a *apperr.FieldError type")
		}

		expectedField := "id"
		expectedErrorMsg := "id is required"
		if fieldErr.Field != expectedField || fieldErr.Message != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s' for field '%s'",
				expectedErrorMsg, expectedField, fieldErr.Message, fieldErr.Field)
		}
	})

	t.Run("MissingName", func(t *testing.T) {
		bucketMissingName := CreateBucket{
			Id:                "bucket123",
			AllowedMimeTypes:  []string{"image/png", "image/jpeg"},
			AllowedObjectSize: 1024,
			CreatedAt:         time.Now(),
		}

		err := bucketMissingName.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(*apperr.FieldError)
		if !ok {
			t.Error("Expected a *apperr.FieldError type")
		}

		expectedField := "name"
		expectedErrorMsg := "name is required"
		if fieldErr.Field != expectedField || fieldErr.Message != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s' for field '%s'",
				expectedErrorMsg, expectedField, fieldErr.Message, fieldErr.Field)
		}
	})

	t.Run("NameContainsWhiteSpace", func(t *testing.T) {
		bucketNameWithWhiteSpace := CreateBucket{
			Id:                "bucket123",
			Name:              "my Bucket",
			AllowedMimeTypes:  []string{"image/png", "image/jpeg"},
			AllowedObjectSize: 1024,
			CreatedAt:         time.Now(),
		}

		err := bucketNameWithWhiteSpace.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(*apperr.FieldError)
		if !ok {
			t.Error("Expected a *apperr.FieldError type")
		}

		expectedField := "name"
		expectedErrorMsg := "name should not contain any white spaces or tabs"
		if fieldErr.Field != expectedField || fieldErr.Message != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s' for field '%s'",
				expectedErrorMsg, expectedField, fieldErr.Message, fieldErr.Field)
		}
	})

	t.Run("InvalidNameFormat", func(t *testing.T) {
		bucketInvalidNameFormat := CreateBucket{
			Id:                "bucket123",
			Name:              "my@Bucket",
			AllowedMimeTypes:  []string{"image/png", "image/jpeg"},
			AllowedObjectSize: 1024,
			CreatedAt:         time.Now(),
		}

		err := bucketInvalidNameFormat.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(*apperr.FieldError)
		if !ok {
			t.Error("Expected a *apperr.FieldError type")
		}

		expectedField := "name"
		expectedErrorMsg := "name should only contain letters, numbers, hyphens and underscores"
		if fieldErr.Field != expectedField || fieldErr.Message != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s' for field '%s'",
				expectedErrorMsg, expectedField, fieldErr.Message, fieldErr.Field)
		}
	})

	t.Run("InvalidMimeTypes", func(t *testing.T) {
		bucketInvalidMimeTypes := CreateBucket{
			Id:                "bucket123",
			Name:              "myBucket",
			AllowedMimeTypes:  []string{"application/executable*", "image/jpeg??/"},
			AllowedObjectSize: 1024,
			CreatedAt:         time.Now(),
		}

		err := bucketInvalidMimeTypes.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(*apperr.FieldError)
		if !ok {
			t.Error("Expected a *apperr.FieldError type")
		}

		expectedField := "allowed_mime_types"
		expectedErrorMsg := `not allowed mime type "application/executable*"`
		if fieldErr.Field != expectedField || fieldErr.Message != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s' for field '%s'",
				expectedErrorMsg, expectedField, fieldErr.Message, fieldErr.Field)
		}
	})

	t.Run("ValidBucket", func(t *testing.T) {
		validBucket := CreateBucket{
			Id:                "bucket123",
			Name:              "myBucket",
			AllowedMimeTypes:  []string{"image/png", "image/jpeg"},
			AllowedObjectSize: 1024,
			CreatedAt:         time.Now(),
		}

		err := validBucket.Validate()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
