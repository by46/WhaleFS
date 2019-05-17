package model

type Chunk struct {
	Fid    string `json:"fid"`
	Etag   string `json:"etag"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
}
