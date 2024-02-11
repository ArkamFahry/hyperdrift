package entities

import "time"

type Object struct {
	Id             string         `json:"id"`
	Version        int32          `json:"version"`
	BucketId       string         `json:"bucket_id"`
	Name           string         `json:"name"`
	ContentType    string         `json:"content_type"`
	Size           int64          `json:"size"`
	Public         bool           `json:"public"`
	Metadata       map[string]any `json:"metadata"`
	UploadStatus   string         `json:"upload_status"`
	LastAccessedAt *time.Time     `json:"last_accessed_at"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      *time.Time     `json:"updated_at"`
}
