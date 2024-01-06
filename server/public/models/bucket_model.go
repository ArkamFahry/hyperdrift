package models

import (
	"regexp"
)

var BucketNameValidateExpr = regexp.MustCompile("^[A-Za-z0-9_-]+$")

type CreateBucket struct {
	Name              string   `json:"name"`
	AllowedMimeTypes  []string `json:"allowed_mime_types"`
	AllowedObjectSize int64    `json:"allowed_object_size"`
}
