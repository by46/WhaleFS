package server

import (
	"archive/tar"
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
	downloadFunc func(string) (io.ReadCloser, http.Header, error)) error {

	pkgType := pkgFileInfo.GetPkgType()

	var tw interface{}
	if pkgType == utils.Zip {
		tw = zip.NewWriter(w)
	} else {
		tw = tar.NewWriter(w)
	}

	defer func() { _ = tw.(io.Closer).Close() }()

	for _, item := range pkgFileInfo.Items {

		entity, err := getEntityFunc(item.RawKey)
		if err != nil {
			return err
		}

		body, _, err := downloadFunc(entity.FID)
		if err != nil {
			return err
		}

		pkgUnitEntity := &utils.PackageUnitEntity{
			Target: item.GetTarget(),
			Size:   entity.Size,
			Reader: body,
		}

		if pkgType == utils.Zip {
			writer := tw.(*zip.Writer)
			err = utils.ZipUnit(writer, pkgUnitEntity)
		} else {
			writer := tw.(*tar.Writer)
			err = utils.TarUnit(writer, pkgUnitEntity)
		}
	}
	return nil
}
