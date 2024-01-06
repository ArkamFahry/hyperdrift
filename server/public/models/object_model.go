package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/ArkamFahry/hyperdrift-storage/server/packages/apperr"
	"github.com/ArkamFahry/hyperdrift-storage/server/public/entities"
	"github.com/oklog/ulid/v2"
)

var (
	ObjectMimeTypesValidatorExpr = regexp.MustCompile(`^[a-zA-Z]+\/[a-zA-Z+\-.]+$`)
)

const (
	ObjectUploadStatusPending   = "pending"
	ObjectUploadStatusCompleted = "completed"
	ObjectUploadStatusFailed    = "failed"
)

type CreateObject struct {
	Id           string    `json:"id"`
	BucketId     string    `json:"bucket_id"`
	Name         string    `json:"name"`
	MimeType     string    `json:"mime_type"`
	ObjectSize   int64     `json:"object_size"`
	UploadStatus string    `json:"upload_status"`
	CreatedAt    time.Time `json:"created_at"`
}

func NewCreateObject(bucketId, bucketName, name, mimeType string, objectSize int64) *CreateObject {
	return &CreateObject{
		Id:           fmt.Sprintf(`%s_%s`, "object", ulid.Make().String()),
		BucketId:     bucketId,
		Name:         fmt.Sprintf(`%s/%s`, bucketName, name),
		MimeType:     mimeType,
		ObjectSize:   objectSize,
		UploadStatus: ObjectUploadStatusPending,
		CreatedAt:    time.Now(),
	}
}

func (co *CreateObject) Validate() error {
	var validationErrors apperr.MapError

	if strings.Trim(co.Id, " ") == "" {
		validationErrors.Set("id", "id is required")
	}

	if strings.Trim(co.Name, " ") == "" {
		validationErrors.Set("name", "name is required")
	}

	if strings.ContainsAny(co.Name, " \t\r\n") {
		validationErrors.Set("name", "name should not contain any white spaces or tabs")
	}

	if len(strings.Split(co.Name, "/")) < 2 {
		validationErrors.Set("name", "name should have two parts bucket name and object name")
	}

	if strings.Split(co.Name, "/")[0] == "" {
		validationErrors.Set("name", "bucket name cannot be empty")
	}

	if strings.Split(co.Name, "/")[1] == "" {
		validationErrors.Set("name", "object name cannot be empty")
	}

	if strings.Trim(co.MimeType, " ") == "" {
		validationErrors.Set("mime_type", "mime_type is required")
	}

	if strings.ContainsAny(co.MimeType, " \t\r\n") {
		validationErrors.Set("mime_type", "mime_type should not contain any white spaces or tabs")
	}

	if !ObjectMimeTypesValidatorExpr.MatchString(co.MimeType) {
		validationErrors.Set("mime_type", "invalid mime type")
	}

	if co.ObjectSize < 0 {
		validationErrors.Set("object_size", "object_size should be greater than 0")
	}

	if strings.Trim(co.UploadStatus, " ") == "" {
		validationErrors.Set("upload_status", "upload_status is required")
	}

	switch co.UploadStatus {
	case ObjectUploadStatusPending, ObjectUploadStatusCompleted, ObjectUploadStatusFailed:
	default:
		validationErrors.Set("upload_status", "invalid upload status")
	}

	if co.CreatedAt.IsZero() {
		validationErrors.Set("created_at", "created_at is required")
	}

	if validationErrors != nil {
		return validationErrors
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
		UploadStatus: co.UploadStatus,
		CreatedAt:    co.CreatedAt,
		UpdatedAt:    nil,
	}
}
