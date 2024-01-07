package models

import (
	"reflect"
	"testing"
	"time"

	"github.com/ArkamFahry/hyperdrift-storage/server/packages/apperr"
)

func TestNewCreateObject(t *testing.T) {
	t.Run("CreateObjectInstance", func(t *testing.T) {
		bucketId := "bucket123"
		bucketName := "myBucket"
		name := "myObject"
		mimeType := "application/json"
		objectSize := int64(1024)
		metadata := map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		}
		createdAt := time.Now()

		expectedObject := &CreateObject{
			Id:           "object_someUniqueId",
			BucketId:     bucketId,
			Name:         "myBucket/myObject",
			MimeType:     mimeType,
			ObjectSize:   objectSize,
			Metadata:     metadata,
			UploadStatus: ObjectUploadStatusPending,
			CreatedAt:    createdAt,
		}

		newObj := NewCreateObject(bucketId, bucketName, name, mimeType, objectSize, metadata)

		if newObj.BucketId != expectedObject.BucketId ||
			newObj.Name != expectedObject.Name ||
			newObj.MimeType != expectedObject.MimeType ||
			newObj.ObjectSize != expectedObject.ObjectSize ||
			!reflect.DeepEqual(newObj.Metadata, expectedObject.Metadata) ||
			newObj.UploadStatus != expectedObject.UploadStatus {
			t.Errorf("Generated object does not match the expected object.\nExpected: %+v\nGot: %+v", expectedObject, newObj)
		}
	})
}

func TestCreateObjectValidation(t *testing.T) {
	t.Run("MissingId", func(t *testing.T) {
		objectMissingID := CreateObject{
			Name:         "bucket/object",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectMissingID.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "id"
		expectedErrorMsg := "id is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("MissingName", func(t *testing.T) {
		objectMissingName := CreateObject{
			Id:           "123",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectMissingName.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "name"
		expectedErrorMsg := "name is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("InvalidNameSpaces", func(t *testing.T) {
		objectInvalidNameSpaces := CreateObject{
			Id:           "123",
			Name:         "bucket/ object",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectInvalidNameSpaces.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "name"
		expectedErrorMsg := "name should not contain any white spaces or tabs"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("InvalidNameParts", func(t *testing.T) {
		objectInvalidNameParts := CreateObject{
			Id:           "123",
			Name:         "bucket",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectInvalidNameParts.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "name"
		expectedErrorMsg := "name should have two parts bucket name and object name"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("EmptyBucketName", func(t *testing.T) {
		objectEmptyBucketName := CreateObject{
			Id:           "123",
			Name:         "/object",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectEmptyBucketName.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "name"
		expectedErrorMsg := "bucket name cannot be empty"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("EmptyObjectName", func(t *testing.T) {
		objectEmptyObjectName := CreateObject{
			Id:           "123",
			Name:         "bucket/",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectEmptyObjectName.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "name"
		expectedErrorMsg := "object name cannot be empty"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("MissingMimeType", func(t *testing.T) {
		objectMissingMimeType := CreateObject{
			Id:           "123",
			Name:         "bucket/object",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectMissingMimeType.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "mime_type"
		expectedErrorMsg := "mime_type is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("InvalidMimeType", func(t *testing.T) {
		objectInvalidMimeType := CreateObject{
			Id:           "123",
			Name:         "bucket/object",
			MimeType:     "image/*jbmp",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectInvalidMimeType.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "mime_type"
		expectedErrorMsg := "invalid mime type"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("NegativeObjectSize", func(t *testing.T) {
		objectNegativeObjectSize := CreateObject{
			Id:           "123",
			Name:         "bucket/object",
			MimeType:     "application/json",
			ObjectSize:   -10,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := objectNegativeObjectSize.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "object_size"
		expectedErrorMsg := "object_size should be greater than 0"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("MissingUploadStatus", func(t *testing.T) {
		objectMissingUploadStatus := CreateObject{
			Id:         "123",
			Name:       "bucket/object",
			MimeType:   "application/json",
			ObjectSize: 100,
			CreatedAt:  time.Now(),
		}

		err := objectMissingUploadStatus.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "upload_status"
		expectedErrorMsg := "upload_status is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("InvalidUploadStatus", func(t *testing.T) {
		objectInvalidUploadStatus := CreateObject{
			Id:           "123",
			Name:         "bucket/object",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: "InvalidStatus",
			CreatedAt:    time.Now(),
		}

		err := objectInvalidUploadStatus.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "upload_status"
		expectedErrorMsg := "invalid upload status"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("EmptyCreatedAt", func(t *testing.T) {
		objectEmptyCreatedAt := CreateObject{
			Id:           "123",
			Name:         "bucket/object",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
		}

		err := objectEmptyCreatedAt.Validate()
		if err == nil {
			t.Error("Expected error, but got nil")
		}

		fieldErr, ok := err.(apperr.MapError)
		if !ok {
			t.Error("Expected a apperr.MapError type")
		}

		expectedField := "created_at"
		expectedErrorMsg := "created_at is required"
		errMsg := fieldErr.Get(expectedField)[0]
		if errMsg != expectedErrorMsg {
			t.Errorf("Expected error message '%s' for field '%s', but got '%s'",
				expectedErrorMsg, expectedField, errMsg)
		}
	})

	t.Run("ValidObject", func(t *testing.T) {
		validObject := CreateObject{
			Id:           "123",
			Name:         "bucket/object",
			MimeType:     "application/json",
			ObjectSize:   100,
			UploadStatus: ObjectUploadStatusCompleted,
			CreatedAt:    time.Now(),
		}

		err := validObject.Validate()
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
	})
}
