package validators

import "regexp"

var (
	MimeTypesValidatorExpr = regexp.MustCompile(`^[a-zA-Z]+\/[a-zA-Z+\-.]+$`)
)

func IsInvalidMimeTypeValid(mimeType string) bool {
	return !MimeTypesValidatorExpr.MatchString(mimeType)
}
