// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package database

import (
	"time"
)

type StorageBucket struct {
	ID                   string
	Name                 string
	AllowedContentTypes  []string
	MaxAllowedObjectSize *int64
	Public               bool
	Disabled             bool
	Locked               bool
	LockReason           *string
	LockedAt             *time.Time
	CreatedAt            time.Time
	UpdatedAt            *time.Time
}

type StorageEvent struct {
	ID        string
	Name      string
	Payload   []byte
	Status    string
	Producer  string
	Timestamp time.Time
}

type StorageObject struct {
	ID             string
	BucketID       string
	Name           string
	PathTokens     []string
	ContentType    string
	Size           int64
	Public         bool
	Metadata       []byte
	UploadStatus   string
	LastAccessedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}
