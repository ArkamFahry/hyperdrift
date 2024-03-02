package services

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMetadataConversion(t *testing.T) {
	metadata := map[string]any{
		"user_id": "user_123456789",
		"user_name": map[string]any{
			"first_name": "John",
			"last_name":  "Doe",
		},
		"access_only_to": []any{"admin", "user"},
	}

	metadataBytes := metadataToBytes(metadata)
	assert.NotNil(t, metadataBytes, "Metadata to bytes conversion failed")

	deserializedMetadata := bytesToMetadata(metadataBytes)
	assert.NotNil(t, deserializedMetadata, "Bytes to metadata conversion failed")
	assert.Equal(t, metadata, deserializedMetadata, "Deserialized metadata does not match original")
}

func TestInvalidBytesToMetadata(t *testing.T) {
	invalidBytes := []byte("invalid JSON")
	deserializedMetadata := bytesToMetadata(invalidBytes)
	assert.Nil(t, deserializedMetadata, "Bytes to metadata conversion should fail with invalid JSON")
}
