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
	MaxAllowedObjectSize *int64 `json:"max_allowed_object_size"`
	Public               *bool  `json:"public"`
}

type BucketAddAllowedContentTypes struct {
	AddContentTypes []string `json:"add_content_types"`
}

type BucketRemoveAllowedContentTypes struct {
	RemoveContentTypes []string `json:"remove_content_types"`
}
