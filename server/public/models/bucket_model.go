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
	BucketNameValidatorExpr             = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)
	BucketAllowedMimeTypesValidatorExpr = regexp.MustCompile(`^[a-zA-Z]+\/[a-zA-Z+\-.]+$`)
)

type CreateBucket struct {
	Id                string    `json:"id"`
	Name              string    `json:"name"`
	AllowedMimeTypes  []string  `json:"allowed_mime_types"`
	AllowedObjectSize int64     `json:"allowed_object_size"`
	CreatedAt         time.Time `json:"created_at"`
}

func NewCreateBucket(name string, allowedMimeTypes []string, allowedObjectSize int64) *CreateBucket {
	if allowedMimeTypes == nil {
		allowedMimeTypes = []string{"*/*"}
	}

	return &CreateBucket{
		Id:                fmt.Sprintf(`%s_%s`, "bucket", ulid.Make().String()),
		Name:              name,
		AllowedMimeTypes:  allowedMimeTypes,
		AllowedObjectSize: allowedObjectSize,
		CreatedAt:         time.Now(),
	}
}

func (cb *CreateBucket) Validate() error {
	var validationErrors apperr.MapError

	if strings.Trim(cb.Id, " ") == "" {
		validationErrors.Set("id", "id is required")
	}

	if strings.Trim(cb.Name, " ") == "" {
		validationErrors.Set("name", "name is required")
	}

	if strings.ContainsAny(cb.Name, " \t\r\n") {
		validationErrors.Set("name", "name should not contain any white spaces or tabs")
	}

	if !BucketNameValidatorExpr.MatchString(cb.Name) {
		validationErrors.Set("name", "name should only contain letters, numbers, hyphens and underscores")
	}

	if len(cb.AllowedMimeTypes) > 0 {
		for _, allowedMimeType := range cb.AllowedMimeTypes {
			if !BucketAllowedMimeTypesValidatorExpr.MatchString(allowedMimeType) {
				validationErrors.Set("allowed_mime_types", fmt.Sprintf(`not allowed mime type "%s"`, allowedMimeType))
				break
			}
		}
	}

	if cb.CreatedAt.IsZero() {
		validationErrors.Set("created_at", "created_at is required")
	}

	if validationErrors != nil {
		return validationErrors
	}

	return nil
}

func (cb *CreateBucket) ConvertToEntity() *entities.Bucket {
	return &entities.Bucket{
		Id:                cb.Id,
		Name:              cb.Name,
		AllowedMimeTypes:  cb.AllowedMimeTypes,
		AllowedObjectSize: cb.AllowedObjectSize,
		CreatedAt:         cb.CreatedAt,
		UpdatedAt:         nil,
	}
}
