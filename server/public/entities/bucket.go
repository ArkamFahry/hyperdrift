package entities

import "time"

type Bucket struct {
	Id                string     `json:"id"`
	Name              string     `json:"name"`
	AllowedMimeTypes  []string   `json:"allowed_mime_types"`
	AllowedObjectSize int64      `json:"allowed_object_size"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at"`
}
