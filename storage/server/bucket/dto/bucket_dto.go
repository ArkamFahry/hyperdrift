package dto

type BucketCreate struct {
	Id                   string   `json:"id"`
	Name                 string   `json:"name"`
	AllowedContentTypes  []string `json:"allowed_content_types"`
	MaxAllowedObjectSize *int64   `json:"max_allowed_object_size"`
	Public               bool     `json:"public"`
	Disabled             bool     `json:"enabled"`
}

type BucketUpdate struct {
	Id                   string `json:"id"`
	MaxAllowedObjectSize *int64 `json:"max_allowed_object_size"`
	Public               *bool  `json:"public"`
}

type BucketAddAllowedContentTypes struct {
	Id                  string   `json:"id"`
	AllowedContentTypes []string `json:"allowed_content_types"`
}

type BucketRemoveAllowedContentTypes struct {
	Id                  string   `json:"id"`
	AllowedContentTypes []string `json:"allowed_content_types"`
}
