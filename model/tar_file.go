package model

type TarFileEntity struct {
	Name  string        `json:"name" form:"name"`
	Items []TarFileItem `json:"items" form:"items"`
}

type TarFileItem struct {
	RawKey string `json:"rawKey" form:"rawKey"`
	Target string `json:"target" form:"target"`
}
