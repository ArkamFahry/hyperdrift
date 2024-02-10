package dto

type BucketCreate struct {
	Name                 string   `json:"name"`
	AllowedContentTypes  []string `json:"allowed_content_types"`
	MaxAllowedObjectSize *int64   `json:"max_allowed_object_size"`
	Public               bool     `json:"public"`
	Disabled             bool     `json:"enabled"`
}

type BucketUpdate struct {
	AllowedContentTypes  []string `json:"allowed_content_types"`
	MaxAllowedObjectSize *int64   `json:"max_allowed_object_size"`
	Public               *bool    `json:"public"`
}
