package models

import (
	"fmt"
	"strings"
	"time"

	"github.com/ArkamFahry/hyperdrift-storage/server/packages/apperr"
	"github.com/ArkamFahry/hyperdrift-storage/server/packages/utils"
	"github.com/ArkamFahry/hyperdrift-storage/server/packages/validators"
	"github.com/ArkamFahry/hyperdrift-storage/server/public/entities"
)

const (
	ObjectUploadStatusPending   = "pending"
	ObjectUploadStatusCompleted = "completed"
	ObjectUploadStatusFailed    = "failed"
)

type CreateObject struct {
	Id           string         `json:"id"`
	BucketId     string         `json:"bucket_id"`
	Name         string         `json:"name"`
	MimeType     string         `json:"mime_type"`
	ObjectSize   int64          `json:"object_size"`
	Metadata     map[string]any `json:"metadata"`
	UploadStatus string         `json:"upload_status"`
	CreatedAt    time.Time      `json:"created_at"`
}

func NewObjectId() string {
	return fmt.Sprintf(`%s_%s`, "object", utils.NewId())
}

func NewObjectName(bucketName, name string) string {
	return fmt.Sprintf(`%s/%s`, bucketName, name)
}

func NewCreateObject(bucketId, bucketName, name, mimeType string, objectSize int64, metadata map[string]any) *CreateObject {
	return &CreateObject{
		Id:           NewObjectId(),
		BucketId:     bucketId,
		Name:         NewObjectName(bucketName, name),
		MimeType:     mimeType,
		ObjectSize:   objectSize,
		Metadata:     metadata,
		UploadStatus: ObjectUploadStatusPending,
		CreatedAt:    time.Now(),
	}
}

func (co *CreateObject) Validate() error {
	if validators.IsEmptyString(co.Id) {
		return apperr.NewFieldError("id", "id is required")
	}

	if validators.IsEmptyString(co.Name) {
		return apperr.NewFieldError("name", "name is required")
	}

	if validators.ContainsAnyWhiteSpaces(co.Name) {
		return apperr.NewFieldError("name", "name should not contain any white spaces or tabs")
	}

	if len(strings.Split(co.Name, "/")) < 2 {
		return apperr.NewFieldError("name", "name should have two parts bucket name and object name")
	}

	if strings.Split(co.Name, "/")[0] == "" {
		return apperr.NewFieldError("name", "bucket name cannot be empty")
	}

	if strings.Split(co.Name, "/")[1] == "" {
		return apperr.NewFieldError("name", "object name cannot be empty")
	}

	if validators.IsEmptyString(co.MimeType) {
		return apperr.NewFieldError("mime_type", "mime_type is required")
	}

	if validators.ContainsAnyWhiteSpaces(co.MimeType) {
		return apperr.NewFieldError("mime_type", "mime_type should not contain any white spaces or tabs")
	}

	if validators.IsInvalidMimeTypeValid(co.MimeType) {
		return apperr.NewFieldError("mime_type", "invalid mime type")
	}

	if co.ObjectSize < 0 {
		return apperr.NewFieldError("object_size", "object_size should be greater than 0")
	}

	if validators.IsEmptyString(co.UploadStatus) {
		return apperr.NewFieldError("upload_status", "upload_status is required")
	}

	switch co.UploadStatus {
	case ObjectUploadStatusPending, ObjectUploadStatusCompleted, ObjectUploadStatusFailed:
	default:
		return apperr.NewFieldError("upload_status", "invalid upload status")
	}

	if co.CreatedAt.IsZero() {
		return apperr.NewFieldError("created_at", "created_at is required")
	}

	return nil
}

func (co *CreateObject) ConvertToEntity() *entities.Object {
	return &entities.Object{
		Id:           co.Id,
		BucketId:     co.BucketId,
		Name:         co.Name,
		MimeType:     co.MimeType,
		ObjectSize:   co.ObjectSize,
		Metadata:     co.Metadata,
		UploadStatus: co.UploadStatus,
		CreatedAt:    co.CreatedAt,
		UpdatedAt:    nil,
	}
}
