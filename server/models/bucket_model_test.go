package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBucketCreate_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		bucket   *BucketCreate
		expected error
	}{
		{
			name: "Valid BucketCreate",
			bucket: &BucketCreate{
				Name:             "avatar",
				AllowedMimeTypes: []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: func() *int64 {
					v := int64(10485760)
					return &v
				}(),
				Public: false,
			},
			expected: nil,
		},
		{
			name: "Invalid BucketCreate (Empty Name)",
			bucket: &BucketCreate{
				Name: "",
			},
			expected: fmt.Errorf("bucket name cannot be empty. bucket name is required to create bucket"),
		},
		{
			name: "Invalid BucketCreate (Invalid Name)",
			bucket: &BucketCreate{
				Name: "invalid_name!",
			},
			expected: fmt.Errorf("bucket name is not valid. it must start and end with an alphanumeric character, and can include alphanumeric characters, hyphens, and dots. The total length must be between 3 and 63 characters"),
		},
		{
			name: "Invalid BucketCreate (Invalid MIME Type)",
			bucket: &BucketCreate{
				Name:             "avatar",
				AllowedMimeTypes: []string{"invalid-mime-type"},
			},
			expected: fmt.Errorf("bucket allowed_mime_types is not valid. invalid mime types: [invalid-mime-type]. allowed mime types must be in the format 'type/subtype'"),
		},
		{
			name: "Invalid BucketCreate (Invalid MIME Type with Wildcard)",
			bucket: &BucketCreate{
				Name:             "avatar",
				AllowedMimeTypes: []string{"*/*", "image/jpeg"},
			},
			expected: fmt.Errorf("bucket allowed_mime_types cannot contain wild card ('*/*') and other mime types at the same time"),
		},
		{
			name: "Invalid BucketCreate (Negative Max Allowed Object Size)",
			bucket: &BucketCreate{
				Name:                 "avatar",
				AllowedMimeTypes:     []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: func() *int64 { v := int64(-1); return &v }(),
			},
			expected: fmt.Errorf("bucket max_allowed_object_size must be greater than 0"),
		},
		{
			name: "Invalid BucketCreate (Zero Max Allowed Object Size)",
			bucket: &BucketCreate{
				Name:                 "avatar",
				AllowedMimeTypes:     []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: func() *int64 { v := int64(0); return &v }(),
			},
			expected: fmt.Errorf("bucket max_allowed_object_size must be greater than 0"),
		},
		{
			name: "Valid BucketCreate (Null Max Allowed Object Size)",
			bucket: &BucketCreate{
				Name:                 "avatar",
				AllowedMimeTypes:     []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: nil,
				Public:               false,
			},
			expected: nil,
		},
		{
			name: "Valid BucketCreate (Null Public)",
			bucket: &BucketCreate{
				Name:                 "avatar",
				AllowedMimeTypes:     []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: func() *int64 { v := int64(10485760); return &v }(),
				Public:               false,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bucket.IsValid()
			assert.Equal(t, tt.expected, err)
		})
	}
}

func TestBucketUpdate_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		bucket   *BucketUpdate
		expected error
	}{
		{
			name: "Valid BucketUpdate",
			bucket: &BucketUpdate{
				Id:               "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				AllowedMimeTypes: []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: func() *int64 {
					v := int64(10485760)
					return &v
				}(),
				Public: func() *bool {
					v := false
					return &v
				}(),
			},
			expected: nil,
		},
		{
			name: "Invalid BucketUpdate (Empty ID)",
			bucket: &BucketUpdate{
				Id:               "",
				AllowedMimeTypes: []string{"image/jpeg", "invalid/mime"},
			},
			expected: fmt.Errorf("bucket id cannot be empty. bucket id is required to update bucket"),
		},
		{
			name: "Invalid BucketUpdate (Invalid MIME Type)",
			bucket: &BucketUpdate{
				Id:               "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				AllowedMimeTypes: []string{"invalid-mime-type"},
			},
			expected: fmt.Errorf("bucket allowed_mime_types is not valid. invalid mime types: [invalid-mime-type], allowed mime types must be in the format 'type/subtype'"),
		},
		{
			name: "Invalid BucketUpdate (Invalid MIME Type with Wildcard)",
			bucket: &BucketUpdate{
				Id:               "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				AllowedMimeTypes: []string{"*/*", "image/jpeg"},
			},
			expected: fmt.Errorf("bucket allowed_mime_types cannot contain wild card ('*/*') and other mime types at the same time"),
		},
		{
			name: "Invalid BucketUpdate (Negative Max Allowed Object Size)",
			bucket: &BucketUpdate{
				Id:                   "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				AllowedMimeTypes:     []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: func() *int64 { v := int64(-1); return &v }(),
			},
			expected: fmt.Errorf("bucket max_allowed_object_size must be greater than 0"),
		},
		{
			name: "Invalid BucketUpdate (Zero Max Allowed Object Size)",
			bucket: &BucketUpdate{
				Id:                   "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				AllowedMimeTypes:     []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: func() *int64 { v := int64(0); return &v }(),
			},
			expected: fmt.Errorf("bucket max_allowed_object_size must be greater than 0"),
		},
		{
			name: "Valid BucketUpdate (Null Max Allowed Object Size)",
			bucket: &BucketUpdate{
				Id:                   "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				AllowedMimeTypes:     []string{"image/jpeg", "image/png"},
				MaxAllowedObjectSize: nil,
			},
			expected: nil,
		},
		{
			name: "Valid BucketUpdate (Null Public)",
			bucket: &BucketUpdate{
				Id:               "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				AllowedMimeTypes: []string{"image/jpeg", "image/png"},
				Public:           nil,
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bucket.IsValid()
			assert.Equal(t, tt.expected, err)
		})
	}
}
