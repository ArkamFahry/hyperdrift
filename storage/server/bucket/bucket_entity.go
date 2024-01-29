package bucket

import "time"

type Bucket struct {
	Id                   string     `json:"id"`
	Name                 string     `json:"name"`
	AllowedContentTypes  []string   `json:"allowed_content_types"`
	MaxAllowedObjectSize *int64     `json:"max_allowed_object_size"`
	Public               bool       `json:"public"`
	Disabled             bool       `json:"enabled"`
	Locked               bool       `json:"locked"`
	LockReason           *string    `json:"lock_reason"`
	LockedAt             *time.Time `json:"locked_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at"`
}
