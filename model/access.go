package model

type AccessKey struct {
	AppKey       string   `json:"app_key"`
	AppSecretKey string   `json:"secret_key"`
	Expires      int64    `json:"expires,omitempty"`
	Scope        []string `json:"scope,omitempty"`
	Owner        string   `json:"owner"`
	Enable       bool     `json:"enable,omitempty"`
	CreateDate   int64    `json:"create_date"`
}
