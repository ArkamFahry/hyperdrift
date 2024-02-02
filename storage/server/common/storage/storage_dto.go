package storage

type ObjectRename struct {
	OldName string `json:"old_name"`
	NewName string `json:"new_name"`
}

type ObjectCopy struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

type ObjectMove struct {
	OldPath string `json:"old_path"`
	NewPath string `json:"new_path"`
}

type PreSignedObject struct {
	Url       string `json:"url"`
	Method    string `json:"method"`
	ExpiresAt int64  `json:"expires_at"`
}

type PreSignedUploadObjectCreate struct {
	Bucket      string `json:"bucket"`
	Name        string `json:"name"`
	ExpiresIn   *int64 `json:"expires_in"`
	ContentType string `json:"content_type"`
	Size        int64  `json:"size"`
}

type PreSignedDownloadObjectCreate struct {
	Bucket    string `json:"bucket"`
	Name      string `json:"name"`
	ExpiresIn *int64 `json:"expires_in"`
}

type ObjectExistsCheck struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

type ObjectDelete struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

type BucketEmpty struct {
	Bucket string `json:"bucket"`
}
