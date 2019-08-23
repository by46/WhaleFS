package model

import (
	"errors"
	"path"
	"strings"

	"github.com/by46/whalefs/utils"
)

const (
	Zip = "zip"
	Tar = "tar"
)

type PackageEntity struct {
	Name  string        `json:"name" form:"name"`
	Type  string        `json:"type" form:"type" default:"zip"`
	Items []PkgFileItem `json:"items" form:"items"`
}

type PkgFileItem struct {
	RawKey string `json:"rawKey" form:"rawKey"`
	Target string `json:"target" form:"target"`
}

func (i *PkgFileItem) GetTarget() string {
	if i.Target != "" {
		ext := path.Ext(i.Target)
		if ext == "" {
			return strings.TrimRight(i.Target, "/") + "/" + utils.PathLastSegment(i.RawKey)
		}
		return i.Target
	}
	return i.RawKey
}

func (e *PackageEntity) GetPkgType() string {
	pkgType := Zip
	suffix := strings.TrimLeft(path.Ext(e.Name), ".")
	if strings.ToLower(suffix) == Tar ||
		(strings.ToLower(suffix) != Zip && strings.ToLower(e.Type) == Tar) {
		pkgType = Tar
	}
	return pkgType
}

func (e *PackageEntity) GetPkgName() string {
	pkgType := e.GetPkgType()
	suffix := strings.TrimLeft(path.Ext(e.Name), ".")
	if suffix == "" || (strings.ToLower(suffix) != Zip && strings.ToLower(suffix) != Tar) {
		return e.Name + "." + pkgType
	}
	return e.Name
}

func (e *PackageEntity) Validate() error {
	// name规则

	if e.Type != "" && strings.ToLower(e.Type) != Zip && strings.ToLower(e.Type) != Tar {
		return errors.New("type must be zip or tar")
	}

	var targets []string
	for _, item := range e.Items {
		if item.RawKey == "" {
			return errors.New("exist empty rawKey")
		}

		if utils.Exists(targets, item.Target) && item.Target != "" {
			return errors.New("exist same target")
		}
		targets = append(targets, item.Target)
	}
	return nil
}
