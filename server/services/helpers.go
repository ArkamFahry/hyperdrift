package services

import (
	"encoding/json"
	"fmt"
	"github.com/ArkamFahry/storage/server/models"
	"github.com/samber/lo"
	"github.com/zhooravell/mime"
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

func determineMimeType(bucket *models.Bucket, preSignedUploadSessionCreate *models.PreSignedUploadSessionCreate) (*string, error) {
	if lo.Contains[string](bucket.AllowedMimeTypes, models.BucketAllowedMimeTypesWildcard) {
		defaultMimeType := models.ObjectDefaultMimeType
		if preSignedUploadSessionCreate.MimeType == nil || (preSignedUploadSessionCreate.MimeType != nil && strings.Trim(*preSignedUploadSessionCreate.MimeType, " ") == "") {
			objectNameParts := strings.Split(preSignedUploadSessionCreate.Name, ".")
			if len(objectNameParts) > 1 {
				objectExtension := objectNameParts[len(objectNameParts)-1]
				mimeType, err := mime.GetMimeTypes(objectExtension)
				if err != nil {
					return &defaultMimeType, nil
				} else {
					return &mimeType[0], nil
				}
			} else {
				return &defaultMimeType, nil
			}
		}
	} else {
		if preSignedUploadSessionCreate.MimeType == nil {
			return nil, fmt.Errorf("mime_type cannot be empty. bucket only allows [%s] mime types. please specify an allowed mime type", strings.Join(bucket.AllowedMimeTypes, ", "))
		} else {
			if !lo.Contains[string](bucket.AllowedMimeTypes, *preSignedUploadSessionCreate.MimeType) {
				return nil, fmt.Errorf("mime_type '%s' is not allowed. bucket only allows [%s] mime types. please specify an allowed mime type", *preSignedUploadSessionCreate.MimeType, strings.Join(bucket.AllowedMimeTypes, ", "))
			}
		}
	}

	return preSignedUploadSessionCreate.MimeType, nil
}
