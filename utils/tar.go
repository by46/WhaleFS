package utils

import (
	"archive/tar"
	"io"
	"os"
	"time"
)

type TarEntity struct {
	Reader io.Reader
	Size   int64
	Target string
	Err    error
}

func BuildPackage(tw *tar.Writer, tarEntity *TarEntity) error {
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
