// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.25.0

package client

import (
	"time"
)

type StorageBucket struct {
	ID                   string
	Name                 string
	AllowedMimeTypes     []string
	MaxAllowedObjectSize *int64
	Public               bool
	Disabled             bool
	Locked               bool
	LockReason           *string
	LockedAt             *time.Time
	CreatedAt            time.Time
	UpdatedAt            *time.Time
}

type StorageObject struct {
	ID             string
	BucketID       string
	Name           string
	PathTokens     []string
	MimeType       string
	Size           int64
	Public         bool
	Metadata       []byte
	UploadStatus   string
	LastAccessedAt *time.Time
	CreatedAt      time.Time
	UpdatedAt      *time.Time
}