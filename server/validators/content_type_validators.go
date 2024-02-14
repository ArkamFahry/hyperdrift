package validators

import (
	"fmt"
	"regexp"
	"strings"
)

func ValidateContentType(mimeType string) bool {
	mimeTypePattern := `^[a-zA-Z]+/[a-zA-Z0-9\-\.\+]+$`

	re := regexp.MustCompile(mimeTypePattern)

	return re.MatchString(mimeType)
}

func ValidateAllowedContentTypes(mimeTypes []string) error {
	var invalidContentTypes []string
	for _, mimeType := range mimeTypes {
		if !ValidateContentType(mimeType) {
			invalidContentTypes = append(invalidContentTypes, mimeType)
		}
	}

	if len(invalidContentTypes) > 0 {
		return fmt.Errorf("invalid content types: [%s]", strings.Join(invalidContentTypes, ", "))
	}

	return nil
}
