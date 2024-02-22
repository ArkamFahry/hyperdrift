package services

import (
	"encoding/json"
	"strings"
)

func metadataToBytes(metadata map[string]any) []byte {
	metadataBytes, err := json.Marshal(metadata)
	if err != nil {
		return nil
	}
	return metadataBytes
}

func bytesToMetadata(metadataBytes []byte) map[string]any {
	var metadata map[string]any
	err := json.Unmarshal(metadataBytes, &metadata)
	if err != nil {
		return nil
	}
	return metadata
}

func isNotEmptyTrimmedString(value string) bool {
	if strings.Trim(value, " ") != "" {
		return true
	} else {
		return false
	}
}
