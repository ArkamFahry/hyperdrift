package entities

import "time"

type Object struct {
	Id           string         `json:"id"`
	BucketId     string         `json:"bucket_id"`
	Name         string         `json:"name"`
	MimeType     string         `json:"mime_type"`
	ObjectSize   int64          `json:"object_size"`
	Metadata     map[string]any `json:"metadata"`
	UploadStatus string         `json:"upload_status"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    *time.Time     `json:"updated_at"`
}
