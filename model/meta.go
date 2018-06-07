package model

type FileEntity struct {
	RawKey       string `json:"raw_key"`
	Url          string `json:"url"`
	LastModified int64  `json:"last_modified"`
	ETag         string `json:"etag"`
	Size         int64  `json:"size"`
	MimeType     string `json:"mime_type"`
}
