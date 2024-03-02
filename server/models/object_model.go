package models

import (
	"fmt"
	"time"
)

const (
	ObjectUploadStatusPending   = "pending"
	ObjectUploadStatusCompleted = "completed"

	ObjectDefaultMimeType = "application/octet-stream"
)

type Object struct {
	Id             string         `json:"id" example:"object_01HPG4GN5JY2Z6S0638ERSG375"`
	Version        int32          `json:"version" example:"0"`
	BucketId       string         `json:"bucket_id" example:"bucket_01HPG4GN5JY2Z6S0638ERSG375"`
	Name           string         `json:"name" example:"user/david/avatar.jpg"`
	MimeType       string         `json:"mime_type" example:"image/jpeg"`
	Size           int64          `json:"size" example:"1218077"`
	Metadata       map[string]any `json:"metadata" extensions:"x-nullable"`
	UploadStatus   string         `json:"upload_status" enum:"pending,completed" example:"pending"`
	LastAccessedAt *time.Time     `json:"last_accessed_at" example:"2024-02-13T08:16:49.952238+05:30" extensions:"x-nullable"`
	CreatedAt      time.Time      `json:"created_at" example:"2024-02-13T08:14:49.952238+05:30"`
	UpdatedAt      *time.Time     `json:"updated_at" example:"2024-02-13T08:18:21.47635+05:30" extensions:"x-nullable"`
}

type PreSignedUploadSession struct {
	Id        string `json:"id" example:"object_01HPG4GN5JY2Z6S0638ERSG375"`
	Url       string `json:"url" example:"http://localhost:9000/test-bucket/avatars/user/david/avatar.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=arkam%2F20240217%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20240217T060609Z&X-Amz-Expires=120&X-Amz-SignedHeaders=content-length%3Bcontent-type%3Bhost&x-id=PutObject&X-Amz-Signature=62421fb20c67e44fee20035e6f40c0f65d105ae75496ede060c1e97e74fe5faa"`
	Method    string `json:"method" default:"PUT" example:"PUT"`
	ExpiresAt int64  `json:"expires_at" example:"1708150396"`
}

type PreSignedDownloadSession struct {
	Url       string `json:"url" example:"http://localhost:9000/test-bucket/avatars/user/david/avatar.jpg?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=arkam%2F20240217%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20240217T060816Z&X-Amz-Expires=300&X-Amz-SignedHeaders=host&x-id=GetObject&X-Amz-Signature=3daaede103c1b99ff6e4ad16f2b64d2becd29ffedd1a3814f418ec5940302cb7"`
	Method    string `json:"method" default:"GET" example:"GET"`
	ExpiresAt int64  `json:"expires_at" example:"1708150396"`
}

type PreSignedUploadSessionCreate struct {
	BucketId  string         `json:"-" params:"bucket_id" example:"bucket_01HPG4GN5JY2Z6S0638ERSG375"`
	Name      string         `json:"name" example:"user/david/avatar.jpg"`
	MimeType  *string        `json:"mime_type" example:"image/jpeg" extensions:"x-nullable"`
	Size      int64          `json:"size" example:"1218077"`
	Metadata  map[string]any `json:"metadata" extensions:"x-nullable"`
	ExpiresIn *int64         `json:"expires_in" example:"600" extensions:"x-nullable"`
}

func (p *PreSignedUploadSessionCreate) IsValid() error {
	if !IsNotEmptyTrimmedString(p.BucketId) {
		return fmt.Errorf("bucket id cannot be empty. bucket id is required to create a pre-signed upload session for an object")
	}
	if !IsNotEmptyTrimmedString(p.Name) {
		return fmt.Errorf("object name cannot be empty. name is required to create a pre-signed upload session for an object")
	}

	if !IsValidObjectName(p.Name) {
		return fmt.Errorf("invalid object name '%s'. object name cannot start or end with '/' and must be between 1 and 961 characters", p.Name)
	}

	if p.MimeType != nil {
		if !IsValidMimeType(*p.MimeType) {
			return fmt.Errorf("invalid mime type '%s'. mime type must be in the format 'type/subtype'", *p.MimeType)
		}
	}

	if p.ExpiresIn != nil {
		if *p.ExpiresIn <= 0 {
			return fmt.Errorf("expires in must be greater than 0")
		}
	}
	return nil
}
