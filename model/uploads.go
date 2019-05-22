package model

type Uploads struct {
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`
	UploadId string `json:"upload_id"`
}

type Part struct {
	PartNumber int32  `json:"part_number"`
	ETag       string `json:"e_tag"`
}

type PartMeta struct {
	Key   string  `json:"key"`
	Parts []*Part `json:"parts"`
}
