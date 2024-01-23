package models

import "time"

type Bucket struct {
	Id                   string     `json:"id"`
	Name                 string     `json:"name"`
	AllowedContentTypes  []string   `json:"allowed_content_types"`
	MaxAllowedObjectSize *int64     `json:"max_allowed_object_size"`
	Public               bool       `json:"public"`
	Disabled             bool       `json:"enabled"`
	Locked               bool       `json:"locked"`
	LockReason           string     `json:"lock_reason"`
	LockedAt             *time.Time `json:"locked_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            *time.Time `json:"updated_at"`
}

type CreateBucket struct {
	Id                   string    `json:"id"`
	Name                 string    `json:"name"`
	AllowedContentTypes  []string  `json:"allowed_content_types"`
	MaxAllowedObjectSize *int64    `json:"max_allowed_object_size"`
	Public               bool      `json:"public"`
	Disabled             bool      `json:"enabled"`
	CreatedAt            time.Time `json:"created_at"`
}

type UpdateBucket struct {
	Id                   string `json:"id"`
	MaxAllowedObjectSize *int64 `json:"max_allowed_object_size"`
	Public               bool   `json:"public"`
}

type AddAllowedContentTypesToBucket struct {
	Id                  string     `json:"id"`
	AllowedContentTypes []string   `json:"allowed_content_types"`
	UpdatedAt           *time.Time `json:"updated_at"`
}

type RemoveAllowedContentTypesFromBucket struct {
	Id                  string     `json:"id"`
	AllowedContentTypes []string   `json:"allowed_content_types"`
	UpdatedAt           *time.Time `json:"updated_at"`
}

type MakeBucketPublic struct {
	Id        string     `json:"id"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type MakeBucketPrivate struct {
	Id        string     `json:"id"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type LockBucket struct {
	Id         string     `json:"id"`
	LockReason string     `json:"lock_reason"`
	LockedAt   *time.Time `json:"locked_at"`
	UpdatedAt  *time.Time `json:"updated_at"`
}

type UnlockBucket struct {
	Id        string     `json:"id"`
	UpdatedAt *time.Time `json:"updated_at"`
}

type DeleteBucket struct {
	Id string `json:"id"`
}
