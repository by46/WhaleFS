package model

// api upuload result
type UploadResut struct {
	Name      string `json:"file_name,omitempty"`
	Extension string `json:"file_ext,omitempty"`
	Url       string `json:"file_path,omitempty"`
}
type ResponseInfo struct {
	Code    string       `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
	Data    *UploadResut `json:"data,omitempty"`
}
