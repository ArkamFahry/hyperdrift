package dto

import "time"

type BucketCreate struct {
	Id                   string    `json:"id"`
	Name                 string    `json:"name"`
	AllowedContentTypes  []string  `json:"allowed_content_types"`
	MaxAllowedObjectSize *int64    `json:"max_allowed_object_size"`
	Public               bool      `json:"public"`
	Disabled             bool      `json:"enabled"`
	CreatedAt            time.Time `json:"created_at"`
}

type BucketUpdate struct {
	Id                   string `json:"id"`
	MaxAllowedObjectSize *int64 `json:"max_allowed_object_size"`
	Public               *bool  `json:"public"`
}

type BucketAddAllowedContentTypes struct {
	Id                  string   `json:"id"`
	AllowedContentTypes []string `json:"allowed_content_types"`
}

type BucketRemoveAllowedContentTypes struct {
	Id                  string   `json:"id"`
	AllowedContentTypes []string `json:"allowed_content_types"`
}

type BucketMakePublic struct {
	Id string `json:"id"`
}

type BucketMakePrivate struct {
	Id string `json:"id"`
}

type BucketLock struct {
	Id         string     `json:"id"`
	LockReason string     `json:"lock_reason"`
	LockedAt   *time.Time `json:"locked_at"`
}

type BucketUnlock struct {
	Id string `json:"id"`
}

type BucketDelete struct {
	Id string `json:"id"`
}
