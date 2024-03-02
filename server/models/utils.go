package models

import (
	"regexp"
	"strings"
)

func IsValidBucketName(name string) bool {
	regexPattern := `^[a-z0-9][a-z0-9.-]{1,61}[a-z0-9]$`

	regex := regexp.MustCompile(regexPattern)

	if len(name) < 3 || len(name) > 63 {
		return false
	}

	if regex.MatchString(name) {
		return true
	} else {
		return false
	}
}

func IsValidObjectName(name string) bool {
	if strings.HasSuffix(name, "/") || strings.HasPrefix(name, "/") {
		return false
	}

	if strings.ContainsRune(name, '\n') || strings.ContainsRune(name, '\t') {
		return false
	}

	if len(name) < 1 || len(name) > 961 {
		return false
	}

	pattern := `^[\s\S]+$`
	regex := regexp.MustCompile(pattern)

	if regex.MatchString(name) {
		return true
	} else {
		return false
	}
}

func IsValidMimeType(mimeType string) bool {
	mimeTypePattern := `^[a-zA-Z]+/[a-zA-Z0-9\-\.\+]+$`

	regex := regexp.MustCompile(mimeTypePattern)

	if regex.MatchString(mimeType) {
		return true
	} else {
		return false
	}
}

func IsNotEmptyTrimmedString(value string) bool {
	if strings.TrimSpace(value) != "" {
		return true
	} else {
		return false
	}
}
