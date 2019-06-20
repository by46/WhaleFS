package model

type SyncFileEntity struct {
	Url      string   `json:"url"`
	RootPath string   `json:"root_path"`
	Sizes    []string `json:"sizes"`
}
