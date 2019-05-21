package utils

import (
	"archive/tar"
	"archive/zip"
	"io"
	"os"
	"path"
	"strings"
	"time"
)

const (
	Zip = "zip"
	Tar = "tar"
)

type PackageUnitEntity struct {
	Reader io.Reader
	Size   int64
	Target string
	Err    error
}

func TarUnit(tw *tar.Writer, tarEntity *PackageUnitEntity) error {
	t := time.Now()
	header := &tar.Header{
		Name:       tarEntity.Target,
		Mode:       0644,
		ModTime:    t,
		Uid:        os.Getuid(),
		Gid:        os.Getgid(),
		Typeflag:   tar.TypeReg,
		AccessTime: t,
		ChangeTime: t,
	}
	header.Size = tarEntity.Size

	err := tw.WriteHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(tw, tarEntity.Reader)

	return err
}

func ZipUnit(zw *zip.Writer, zipEntity *PackageUnitEntity) error {
	t := time.Now()
	header := &zip.FileHeader{
		Name:               zipEntity.Target,
		UncompressedSize64: uint64(zipEntity.Size),
		Modified:           t.UTC(),
	}
	header.SetMode(0644)
	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, zipEntity.Reader)
	return err
}

func GetPkgType(pkgFileName string, pkgFileType string) string {
	pkgType := Zip
	suffix := strings.TrimLeft(path.Ext(pkgFileName), ".")
	if strings.ToLower(suffix) == Tar ||
		(strings.ToLower(suffix) != Zip && strings.ToLower(pkgFileType) == Tar) {
		pkgType = Tar
	}
	return pkgType
}

func GetPkgFileName(originName string, pkgType string) string {
	suffix := strings.TrimLeft(path.Ext(originName), ".")

	if suffix == "" || (strings.ToLower(suffix) != Zip && strings.ToLower(suffix) != Tar) {
		return originName + "." + pkgType
	}
	return originName
}
