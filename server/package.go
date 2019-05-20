package server

import (
	"archive/zip"
	"io"
	"net/http"

	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

func Package(
	pkgFileInfo *model.PackageEntity,
	w io.Writer,
	getEntityFunc func(string) (*model.FileMeta, error),
	downloadFunc func(string) (io.Reader, http.Header, error)) error {

	//tw := tar.NewWriter(w)
	tw := zip.NewWriter(w)
	defer func() { _ = tw.Close() }()

	fileReaderChan := make(chan *utils.PackageUnitEntity, len(pkgFileInfo.Items))
	defer close(fileReaderChan)

	for _, item := range pkgFileInfo.Items {
		go func(item model.PkgFileItem) {
			pkgUnitEntity := &utils.PackageUnitEntity{
				Target: item.Target,
			}
			defer func() { fileReaderChan <- pkgUnitEntity }()

			entity, err := getEntityFunc(item.RawKey)
			if err != nil {
				pkgUnitEntity.Err = err
				fileReaderChan <- pkgUnitEntity
				return
			}

			body, _, err := downloadFunc(entity.FID)
			if err != nil {
				pkgUnitEntity.Err = err
				fileReaderChan <- pkgUnitEntity
				return
			}

			pkgUnitEntity.Size = entity.Size
			pkgUnitEntity.Reader = body
		}(item)
	}

	for i := 0; i < len(pkgFileInfo.Items); i++ {
		pkgUnitEntity := <-fileReaderChan
		if pkgUnitEntity.Err != nil {
			return pkgUnitEntity.Err
		}
		//err := utils.TarUnit(tw, pkgUnitEntity)
		err := utils.ZipUnit(tw, pkgUnitEntity)
		if err != nil {
			return err
		}
	}
	return nil
}
