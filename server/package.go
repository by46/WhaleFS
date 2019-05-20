package server

import (
	"archive/tar"
	"archive/zip"
	"io"
	"net/http"
	"strings"

	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

const (
	Zip = "zip"
	Tar = "tar"
)

func Package(
	pkgFileInfo *model.PackageEntity,
	w io.Writer,
	getEntityFunc func(string) (*model.FileMeta, error),
	downloadFunc func(string) (io.Reader, http.Header, error)) error {

	if strings.TrimSpace(pkgFileInfo.Type) == "" {
		pkgFileInfo.Type = Zip
	}

	var tw interface{}
	if strings.ToLower(pkgFileInfo.Type) == Zip {
		tw = zip.NewWriter(w)
	} else {
		tw = tar.NewWriter(w)
	}

	defer func() { _ = tw.(io.Closer).Close() }()

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
		var err error
		if strings.ToLower(pkgFileInfo.Type) == Zip {
			writer := tw.(*zip.Writer)
			err = utils.ZipUnit(writer, pkgUnitEntity)
		} else {
			writer := tw.(*tar.Writer)
			err = utils.TarUnit(writer, pkgUnitEntity)
		}

		if err != nil {
			return err
		}
	}
	return nil
}
