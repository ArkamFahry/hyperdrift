package models

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPreSignedUploadSessionCreate_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		session  *PreSignedUploadSessionCreate
		expected error
	}{
		{
			name: "Valid PreSignedUploadSessionCreate",
			session: &PreSignedUploadSessionCreate{
				BucketId: "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				Name:     "user/david/avatar.jpg",
				MimeType: func() *string {
					v := "image/jpeg"
					return &v
				}(),
				Size: 1218077,
				ExpiresIn: func() *int64 {
					v := int64(600)
					return &v
				}(),
			},
			expected: nil,
		},
		{
			name: "Invalid PreSignedUploadSessionCreate (Empty BucketId)",
			session: &PreSignedUploadSessionCreate{
				BucketId: "",
				Name:     "user/david/avatar.jpg",
				Size:     1218077,
			},
			expected: fmt.Errorf("bucket id cannot be empty. bucket id is required to create a pre-signed upload session for an object"),
		},
		{
			name: "Invalid PreSignedUploadSessionCreate (Empty Name)",
			session: &PreSignedUploadSessionCreate{
				BucketId: "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				Name:     "",
				Size:     1218077,
			},
			expected: fmt.Errorf("object name cannot be empty. name is required to create a pre-signed upload session for an object"),
		},
		{
			name: "Invalid PreSignedUploadSessionCreate (Invalid Name)",
			session: &PreSignedUploadSessionCreate{
				BucketId: "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				Name:     "/invalid/name",
				Size:     1218077,
			},
			expected: fmt.Errorf("invalid object name '/invalid/name'. object name cannot start or end with '/' and must be between 1 and 961 characters"),
		},
		{
			name: "Invalid PreSignedUploadSessionCreate (Invalid MIME Type)",
			session: &PreSignedUploadSessionCreate{
				BucketId: "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				Name:     "user/david/avatar.jpg",
				MimeType: func() *string {
					v := "invalid-mime-type"
					return &v
				}(),
				Size: 1218077,
			},
			expected: fmt.Errorf("invalid mime type 'invalid-mime-type'. mime type must be in the format 'type/subtype'"),
		},
		{
			name: "Invalid PreSignedUploadSessionCreate (Negative ExpiresIn)",
			session: &PreSignedUploadSessionCreate{
				BucketId: "bucket_01HPG4GN5JY2Z6S0638ERSG375",
				Name:     "user/david/avatar.jpg",
				Size:     1218077,
				ExpiresIn: func() *int64 {
					v := int64(-1)
					return &v
				}(),
			},
			expected: fmt.Errorf("expires in must be greater than 0"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.session.IsValid()
			assert.Equal(t, tt.expected, err)
		})
	}
}
