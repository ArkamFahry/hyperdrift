package validators

import "strings"

func IsEmptyString(value string) bool {
	return strings.TrimSpace(value) == ""
}

func ContainsAnyWhiteSpaces(value string) bool {
	return strings.ContainsAny(value, " \t\r\n")
}
