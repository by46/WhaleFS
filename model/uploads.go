package model

type Uploads struct {
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`
	UploadId string `json:"uploadId"`
}

type Part struct {
	PartNumber int32  `json:"partNumber"`
	FID        string `json:"fid"`
	Size       int64  `json:"size"`
}

type PartMeta struct {
	Key      string  `json:"key"`
	MimeType string  `json:"mimeType"`
	Parts    []*Part `json:"parts"`
}
