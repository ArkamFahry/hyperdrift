package dto

type BucketCreate struct {
	/*
		name should start and end with an alphanumeric character,
		and can include alphanumeric characters, hyphens, and dots.
		The total length must be between 3 and 63 characters.
		name is required to create bucket cannot be empty
	*/
	Name string `json:"name" example:"avatar"`
	//	allowed_content_types should be a list of valid mimetypes,
	//	or it can be empty list or `null` then the system would infer it as wild card `*/*` allowing all content types.
	//	if a list mime types are being sent all of them should be valid.
	//	also if allowed content types are being set it can't include wild card with all other content types like `["*/*", "video/mp4", "audio/wav"]`
	// 	this will be invalid if wild card is going to be set it should only be used by itself like `["*/*"]`
	AllowedContentTypes []string `json:"allowed_content_types" example:"image/jpeg, image/png, video/mp4, audio/wav" extensions:"x-nullable"`
	/*
		max_allowed_object_size should be the max size of an object allowed to be uploaded into a bucket.
		the max allowed size should be defined in `bytes`. if it's set to `null` the system will infer this as there
		is no upper limit to the object size that can be uploaded
	*/
	MaxAllowedObjectSize *int64 `json:"max_allowed_object_size" example:"10485760" extensions:"x-nullable"`
	/*
		public can be true or false. if public is true the bucket will accessible publicly without authentication.
		if public is false the bucket will only accessible with authentication. if set to `null` defaults to `false`
	*/
	Public bool `json:"public" default:"false" example:"false" extensions:"x-nullable"`
}

type BucketUpdate struct {
	//	allowed_content_types should be a list of valid mimetypes,
	//	or it can be empty list or `null` then the system would infer it as wild card `*/*` allowing all content types.
	//	if a list mime types are being sent all of them should be valid.
	//	also if allowed content types are being set it can't include wild card with all other content types like `["*/*", "video/mp4", "audio/wav"]`
	// 	this will be invalid if wild card is going to be set it should only be used by itself like `["*/*"]`. if the new allowed_content_types are valid
	//  they will replace the previously defined allowed_content_types
	AllowedContentTypes []string `json:"allowed_content_types" example:"image/jpeg, image/png, video/mp4, audio/wav" extensions:"x-nullable"`
	/*
		max_allowed_object_size should be the max size of an object allowed to be uploaded into a bucket.
		the max allowed size should be defined in `bytes`. if it's set to `null` the system will infer this as there
		is no upper limit to the object size that can be uploaded
	*/
	MaxAllowedObjectSize *int64 `json:"max_allowed_object_size" example:"10485760" extensions:"x-nullable"`
	/*
		public can be true or false. if public is true the bucket will accessible publicly without authentication.
		if public is false the bucket will only accessible with authentication. if set to `null` defaults to `false`
	*/
	Public *bool `json:"public" example:"false" extensions:"x-nullable"`
}
