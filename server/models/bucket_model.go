package models

import (
	"fmt"
	"github.com/samber/lo"
	"strings"
	"time"
)

const (
	BucketLockedReasonBucketDeletion = "bucket.deletion"
	BucketLockedReasonBucketEmptying = "bucket.emptying"

	BucketAllowedMimeTypesWildcard = "*/*"
)

type Bucket struct {
	Id                   string     `json:"id" example:"bucket_01HPG4GN5JY2Z6S0638ERSG375"`
	Version              int32      `json:"version" example:"0"`
	Name                 string     `json:"name" example:"avatar"`
	AllowedMimeTypes     []string   `json:"allowed_mime_types" example:"image/jpeg, image/png, video/mp4, audio/wav"`
	MaxAllowedObjectSize *int64     `json:"max_allowed_object_size" example:"10485760" extensions:"x-nullable"`
	Public               bool       `json:"public" example:"false"`
	Disabled             bool       `json:"disabled" example:"false"`
	Locked               bool       `json:"locked" example:"false"`
	LockReason           *string    `json:"lock_reason" enum:"bucket.deletion,bucket.emptying" example:"bucket.deletion" extensions:"x-nullable"`
	LockedAt             *time.Time `json:"locked_at" default:"2024-02-13T08:16:49.952238+05:30" extensions:"x-nullable"`
	CreatedAt            time.Time  `json:"created_at" default:"2024-02-13T08:14:49.952238+05:30"`
	UpdatedAt            *time.Time `json:"updated_at" default:"2024-02-13T08:18:21.47635+05:30" extensions:"x-nullable"`
}

type BucketSize struct {
	Id   string `json:"id" example:"bucket_01HPG4GN5JY2Z6S0638ERSG375"`
	Name string `json:"name" example:"avatar"`
	Size int64  `json:"size" example:"10737418240"`
}

type BucketCreate struct {
	/*
		`name` should start and end with an alphanumeric character,
		and can include alphanumeric characters, hyphens, and dots.
		The total length must be between 3 and 63 characters.
		name is required to create bucket cannot be empty
	*/
	Name string `json:"name" example:"avatar"`
	//	`allowed_mime_types` should be a list of valid mimetypes,
	//	or it can be empty list or `null` then the system would infer it as wild card `*/*` allowing all content types.
	//	if a list mime types are being sent all of them should be valid.
	//	also if allowed content types are being set it can't include wild card with all other content types like `["*/*", "video/mp4", "audio/wav"]`
	// 	this will be invalid if wild card is going to be set it should only be used by itself like `["*/*"]`
	AllowedMimeTypes []string `json:"allowed_mime_types" example:"image/jpeg, image/png, video/mp4, audio/wav" extensions:"x-nullable"`
	/*
		`max_allowed_object_size` should be the max size of an object allowed to be uploaded into a bucket.
		the max allowed size should be defined in `bytes`. if it's set to `null` the system will infer this as there
		is no upper limit to the object size that can be uploaded
	*/
	MaxAllowedObjectSize *int64 `json:"max_allowed_object_size" example:"10485760" extensions:"x-nullable"`
	/*
		`public` can be true or false. if public is true the bucket will accessible publicly without authentication.
		if public is false the bucket will only accessible with authentication. if set to `null` defaults to `false`
	*/
	Public bool `json:"public" default:"false" example:"false" extensions:"x-nullable"`
}

func (b *BucketCreate) IsValid() error {
	if !isNotEmptyTrimmedString(b.Name) {
		return fmt.Errorf("bucket name cannot be empty. bucket name is required to create bucket")
	}

	if !isValidBucketName(b.Name) {
		return fmt.Errorf("bucket name is not valid. it must start and end with an alphanumeric character, and can include alphanumeric characters, hyphens, and dots. The total length must be between 3 and 63 characters")
	}

	if b.AllowedMimeTypes != nil {
		if len(b.AllowedMimeTypes) > 1 {
			if lo.Contains[string](b.AllowedMimeTypes, BucketAllowedMimeTypesWildcard) {
				return fmt.Errorf("bucket allowed_mime_types cannot contain wild card ('*/*') and other mime types at the same time")
			}
		}

		var invalidMimeTypes []string
		for _, allowedMimeType := range b.AllowedMimeTypes {
			if !isValidMimeType(allowedMimeType) {
				invalidMimeTypes = append(invalidMimeTypes, allowedMimeType)
			}
		}

		if len(invalidMimeTypes) > 0 {
			return fmt.Errorf("bucket allowed_mime_types is not valid. invalid mime types: [%s]. allowed mime types must be in the format 'type/subtype'", strings.Join(invalidMimeTypes, ", "))
		}
	}

	if b.MaxAllowedObjectSize != nil {
		if *b.MaxAllowedObjectSize <= 0 {
			return fmt.Errorf("bucket max_allowed_object_size must be greater than 0")
		}
	}

	return nil
}

func (b *BucketCreate) PreSave() {
	if b.AllowedMimeTypes == nil {
		b.AllowedMimeTypes = []string{BucketAllowedMimeTypesWildcard}
	}
}

type BucketUpdate struct {
	Id string `json:"-" params:"id" example:"bucket_01HPG4GN5JY2Z6S0638ERSG375"`
	//	`allowed_mime_types` should be a list of valid mimetypes,
	//	or it can be empty list or `null` then the system would infer it as wild card `*/*` allowing all content types.
	//	if a list mime types are being sent all of them should be valid.
	//	also if allowed content types are being set it can't include wild card with all other content types like `["*/*", "video/mp4", "audio/wav"]`
	// 	this will be invalid if wild card is going to be set it should only be used by itself like `["*/*"]`. if the new allowed_content_types are valid
	//  they will replace the previously defined allowed_mime_types
	AllowedMimeTypes []string `json:"allowed_mime_types" example:"image/jpeg, image/png, video/mp4, audio/wav" extensions:"x-nullable"`
	/*
		`max_allowed_object_size` should be the max size of an object allowed to be uploaded into a bucket.
		the max allowed size should be defined in `bytes`. if it's set to `null` the system will infer this as there
		is no upper limit to the object size that can be uploaded
	*/
	MaxAllowedObjectSize *int64 `json:"max_allowed_object_size" example:"10485760" extensions:"x-nullable"`
	/*
		`public` can be true or false. if public is true the bucket will accessible publicly without authentication.
		if public is false the bucket will only accessible with authentication. if set to `null` defaults to `false`
	*/
	Public *bool `json:"public" example:"false" extensions:"x-nullable"`
}

func (b *BucketUpdate) IsValid() error {
	if !isNotEmptyTrimmedString(b.Id) {
		return fmt.Errorf("bucket id cannot be empty. bucket id is required to update bucket")
	}

	if b.AllowedMimeTypes != nil {
		if len(b.AllowedMimeTypes) > 1 {
			if lo.Contains[string](b.AllowedMimeTypes, BucketAllowedMimeTypesWildcard) {
				return fmt.Errorf("bucket allowed_mime_types cannot contain wild card ('*/*') and other mime types at the same time")
			}
		}

		var invalidMimeTypes []string
		for _, allowedMimeType := range b.AllowedMimeTypes {
			if !isValidMimeType(allowedMimeType) {
				invalidMimeTypes = append(invalidMimeTypes, allowedMimeType)
			}
		}

		if len(invalidMimeTypes) > 0 {
			return fmt.Errorf("bucket allowed_mime_types is not valid. invalid mime types: [%s], allowed mime types must be in the format 'type/subtype'", strings.Join(invalidMimeTypes, ", "))
		}
	}

	if b.MaxAllowedObjectSize != nil {
		if *b.MaxAllowedObjectSize <= 0 {
			return fmt.Errorf("bucket max_allowed_object_size must be greater than 0")
		}
	}

	return nil
}
