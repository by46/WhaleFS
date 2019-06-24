package utils

import (
	"archive/tar"
	"archive/zip"
	"io"
	"os"
	"time"
)

type PackageUnitEntity struct {
	Reader io.ReadCloser
	Size   int64
	Target string
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
	err = tarEntity.Reader.Close()

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
	err = zipEntity.Reader.Close()

	return err
}
