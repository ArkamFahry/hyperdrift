package services

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

func validateObjectName(name string) error {
	if strings.HasSuffix(name, "/") || strings.HasPrefix(name, "/") {
		return fmt.Errorf("invalid name. name cannot start or end with '/'")
	}

	if len(name) < 1 || len(name) > 961 {
		return fmt.Errorf("invalid name length: %d. name must be between 1 and 961", len(name))
	}

	pattern := `^[\s\S]+$`
	re := regexp.MustCompile(pattern)

	if re.MatchString(name) {
		return nil
	}

	return fmt.Errorf("invalid name '%s'", name)
}

func validateExpiration(expiresIn int64) error {
	if expiresIn <= 0 {
		return fmt.Errorf("expires in must be greater than 0")
	}

	return nil
}

func validateObjectSize(size int64) error {
	if size <= 0 {
		return fmt.Errorf("content lenght must be greater than 0")
	}

	return nil
}

func metadataToBytes(metadata map[string]any) ([]byte, error) {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata to bytes: %w", err)
	}
	return metadataBytes, nil
}

func bytesToMetadata(metadataBytes []byte) (map[string]any, error) {
	var metadata map[string]any
	err := json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal metadata from bytes: %w", err)
	}
	return metadata, nil
}

func validateMimeType(mimeType string) error {
	mimeTypePattern := `^[a-zA-Z]+/[a-zA-Z0-9\-\.\+]+$`

	re := regexp.MustCompile(mimeTypePattern)

	if re.MatchString(mimeType) {
		return nil
	}

	return fmt.Errorf("invalid mime type '%s'", mimeType)
}

func validateAllowedMimeTypes(mimeTypes []string) error {
	var invalidContentTypes []string
	for _, mimeType := range mimeTypes {
		if err := validateMimeType(mimeType); err != nil {
			invalidContentTypes = append(invalidContentTypes, mimeType)
		}
	}

	if len(invalidContentTypes) > 0 {
		return fmt.Errorf("invalid content types: [%s]", strings.Join(invalidContentTypes, ", "))
	}

	return nil
}

func validateNotEmptyTrimmedString(value string) bool {
	if strings.Trim(value, " ") == "" {
		return true
	}

	return false
}
