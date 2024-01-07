package validators

import "strings"

func IsEmptyString(value string) bool {
	return strings.TrimSpace(value) == ""
}

func IsZeroOrEmptyString(value string) bool {
	return strings.TrimSpace(value) == "" || strings.TrimSpace(value) == "0"
}

func ContainsAnyWhiteSpaces(value string) bool {
	return strings.ContainsAny(value, " \t\r\n")
}
