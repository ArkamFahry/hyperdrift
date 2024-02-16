package models

import "time"

const (
	ObjectUploadStatusPending   = "pending"
	ObjectUploadStatusCompleted = "completed"

	ObjectDefaultMimeType = "application/octet-stream"
)

type Object struct {
	Id             string         `json:"id"`
	Version        int32          `json:"version"`
	BucketId       string         `json:"bucket_id"`
	Name           string         `json:"name"`
	MimeType       string         `json:"mime_type"`
	Size           int64          `json:"size"`
	Metadata       map[string]any `json:"metadata"`
	UploadStatus   string         `json:"upload_status"`
	LastAccessedAt *time.Time     `json:"last_accessed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      *time.Time     `json:"updated_at"`
}

type ObjectCreate struct {
	Id             string     `json:"id"`
	Name           string     `json:"name"`
	BucketId       string     `json:"bucket_id"`
	ContentType    string     `json:"content_type"`
	Size           int64      `json:"size"`
	Metadata       []byte     `json:"metadata"`
	UploadStatus   string     `json:"upload_status"`
	LastAccessedAt *time.Time `json:"last_accessed_at"`
	CreatedAt      time.Time  `json:"created_at"`
}

type ObjectRename struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

type ObjectCopy struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

type ObjectMove struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

type PreSignedUploadSession struct {
	Id        string `json:"id"`
	Url       string `json:"url"`
	Method    string `json:"method"`
	ExpiresAt int64  `json:"expires_at"`
}

type PreSignedDownloadSession struct {
	Url       string `json:"url"`
	Method    string `json:"method" default:"GET" example:"GET"`
	ExpiresAt int64  `json:"expires_at" `
}

type PreSignedUploadSessionCreate struct {
	Name      string         `json:"name" example:"user/david/avatar.jpg"`
	MimeType  *string        `json:"mime_type" example:"image/jpeg" extensions:"x-nullable"`
	Size      int64          `json:"size" example:"1218077"`
	Metadata  map[string]any `json:"metadata" extensions:"x-nullable"`
	ExpiresIn *int64         `json:"expires_in" example:"600" extensions:"x-nullable"`
}
