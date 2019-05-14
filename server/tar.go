package server

import (
	"archive/tar"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

type TarEntity struct {
	Reader io.Reader
	Size   int64
	Target string
	Err    error
}

func Package(
	tarFileInfo *model.TarFileEntity,
	w io.Writer,
	getEntityFunc func(hash string) (*model.FileMeta, error),
	downloadFunc func(url string) (io.Reader, http.Header, error)) error {

	tw := tar.NewWriter(w)
	defer func() { _ = tw.Close() }()

	fileReaderChan := make(chan *TarEntity, len(tarFileInfo.Items))
	defer close(fileReaderChan)

	for _, item := range tarFileInfo.Items {
		go func(item model.TarFileItem) {
			tarEntity := &TarEntity{
				Target: item.Target,
			}
			defer func() { fileReaderChan <- tarEntity }()

			hashKey, err := utils.Sha1(item.RawKey)
			if err != nil {
				tarEntity.Err = err
				fileReaderChan <- tarEntity
				return
			}
			entity, err := getEntityFunc(hashKey)
			if err != nil {
				tarEntity.Err = err
				fileReaderChan <- tarEntity
				return
			}

			body, _, err := downloadFunc(entity.FID)
			if err != nil {
				tarEntity.Err = err
				fileReaderChan <- tarEntity
				return
			}

			tarEntity.Size = entity.Size
			tarEntity.Reader = body
		}(item)
	}

	for i := 0; i < len(tarFileInfo.Items); i++ {
		tarEntity := <-fileReaderChan
		if tarEntity.Err != nil {
			return tarEntity.Err
		}
		err := BuildPackage(tw, tarEntity)
		if err != nil {
			return err
		}
	}
	return nil
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
