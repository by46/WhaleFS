package server

import (
	"archive/zip"
	"io"
	"net/http"

	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

func Package(
	tarFileInfo *model.TarFileEntity,
	w io.Writer,
	getEntityFunc func(hash string) (*model.FileMeta, error),
	downloadFunc func(url string) (io.Reader, http.Header, error)) error {

	//tw := tar.NewWriter(w)
	tw := zip.NewWriter(w)
	defer func() { _ = tw.Close() }()

	fileReaderChan := make(chan *utils.PackageUnitEntity, len(tarFileInfo.Items))
	defer close(fileReaderChan)

	for _, item := range tarFileInfo.Items {
		go func(item model.TarFileItem) {
			tarEntity := &utils.PackageUnitEntity{
				Target: item.Target,
			}
			defer func() { fileReaderChan <- tarEntity }()

			entity, err := getEntityFunc(item.RawKey)
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
		//err := utils.TarUnit(tw, tarEntity)
		err := utils.ZipUnit(tw, tarEntity)
		if err != nil {
			return err
		}
	}
	return nil
}
