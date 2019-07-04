package model

type AccessKey struct {
	AppKey       string   `json:"app_key"`
	AppSecretKey string   `json:"secret_key"`
	Expires      int64    `json:"expires"`
	Scope        []string `json:"scope,omitempty"`
	Owner        string   `json:"owner"`
}
