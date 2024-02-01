package dto

import "time"

type Bucket struct {
	Id                   string     `json:"id"`
	Version              int32      `json:"version"`
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

type BucketSize struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Size int64  `json:"size"`
}

type BucketCreate struct {
	Id                   string   `json:"id"`
	Name                 string   `json:"name"`
	AllowedContentTypes  []string `json:"allowed_content_types"`
	MaxAllowedObjectSize *int64   `json:"max_allowed_object_size"`
	Public               bool     `json:"public"`
	Disabled             bool     `json:"enabled"`
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

type BucketEmpty struct {
	Id string `json:"id"`
}

type BucketDelete struct {
	Id string `json:"id"`
}
