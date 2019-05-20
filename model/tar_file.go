package model

type PackageEntity struct {
	Name  string        `json:"name" form:"name"`
	Type  string        `json:"type" form:"type" default:"zip"`
	Items []PkgFileItem `json:"items" form:"items"`
}

type PkgFileItem struct {
	RawKey string `json:"rawKey" form:"rawKey"`
	Target string `json:"target" form:"target"`
}
