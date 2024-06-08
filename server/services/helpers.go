package services

import (
	"encoding/json"
	"fmt"
	"github.com/driftdev/storage/server/models"
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
	if preSignedUploadSessionCreate.MimeType != nil {
		if !models.IsNotEmptyTrimmedString(*preSignedUploadSessionCreate.MimeType) {
			return nil, fmt.Errorf("mime_type cannot be empty. please specify a valid mime type")
		}

		if !models.IsValidMimeType(*preSignedUploadSessionCreate.MimeType) {
			return nil, fmt.Errorf("mime_type '%s' is not valid. please specify a valid mime type", *preSignedUploadSessionCreate.MimeType)
		}

		if lo.Contains[string](bucket.AllowedMimeTypes, models.BucketAllowedMimeTypesWildcard) {
			return preSignedUploadSessionCreate.MimeType, nil
		} else {
			if !lo.Contains[string](bucket.AllowedMimeTypes, *preSignedUploadSessionCreate.MimeType) {
				return nil, fmt.Errorf("mime_type '%s' is not allowed. bucket only allows [%s] mime types. please specify an allowed mime type", *preSignedUploadSessionCreate.MimeType, strings.Join(bucket.AllowedMimeTypes, ", "))
			}
		}

		return preSignedUploadSessionCreate.MimeType, nil
	} else {
		defaultMimeType := models.ObjectDefaultMimeType
		if lo.Contains[string](bucket.AllowedMimeTypes, models.BucketAllowedMimeTypesWildcard) {
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
		} else {
			return nil, fmt.Errorf("mime_type cannot be empty. bucket only allows [%s] mime types. please specify an allowed mime type", strings.Join(bucket.AllowedMimeTypes, ", "))
		}
	}
}
