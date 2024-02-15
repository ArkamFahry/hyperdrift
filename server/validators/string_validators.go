package validators

import "strings"

func ValidateNotEmptyTrimmedString(value string) bool {
	if strings.Trim(value, " ") == "" {
		return true
	}

	return false
}
