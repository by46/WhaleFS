package model

type Chunk struct {
	Fid    string `json:"fid"`
	Etag   string `json:"etag"`
	Size   int64  `json:"size,omitempty"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}
