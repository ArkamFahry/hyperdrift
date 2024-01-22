package models

type PreSignedUrl struct {
	Url       string `json:"url"`
	Method    string `json:"method"`
	ExpiresAt int64  `json:"expires_at"`
}

type CreatePreSignedUploadUrl struct {
	Path      string `json:"path"`
	ExpiresIn *int64 `json:"expires_in"`
	MimeType  string `json:"mime_type"`
	Size      int64  `json:"size"`
}

type CreatePreSignedDownloadUrl struct {
	Path      string `json:"path"`
	ExpiresIn *int64 `json:"expires_in"`
}
