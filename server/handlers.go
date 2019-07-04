package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

func (s *Server) favicon(ctx echo.Context) error {
	// TODO(benjamin): 添加新鲜度检查
	return ctx.File("static/logo.png")
}

func (s *Server) faq(ctx echo.Context) error {
	return ctx.HTML(http.StatusOK, "<!-- Newegg -->")
}

func (s *Server) tools(ctx echo.Context) error {
	if ctx.Request().Method == "GET" {
		return ctx.File("templates/tools.html")
	}
	return s.error(http.StatusForbidden, fmt.Errorf("method not implements"))
}

func (s *Server) pkgDownloadTool(ctx echo.Context) error {
	if ctx.Request().Method == "GET" {
		return ctx.File("templates/pkg-download-tool.html")
	}
	return s.error(http.StatusForbidden, fmt.Errorf("method not implements"))
}

func (s *Server) packageDownload(ctx echo.Context) error {
	content := ctx.FormValue("content")
	packageEntity := new(model.PackageEntity)
	err := json.Unmarshal([]byte(content), &packageEntity)
	if err != nil {
		return errors.WithStack(err)
	}

	err = packageEntity.Validate()
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("bad reqeust: %v", err))
	}

	var totalSize int64
	for _, item := range packageEntity.Items {
		key := utils.PathNormalize(item.RawKey)
		entity, err := s.GetFileEntity(key, len(key) != len(item.RawKey))
		if err != nil {
			if err == common.ErrKeyNotFound {
				return echo.NewHTTPError(http.StatusNotFound, fmt.Sprintf("%s not found", item.RawKey))
			}
			return errors.WithStack(err)
		}

		totalSize = totalSize + entity.Size
	}

	if totalSize > s.TaskFileSizeThreshold {
		hashKey, err := utils.Sha1(fmt.Sprintf("/%s/%s", s.TaskBucketName, packageEntity.Name))
		err = s.CreateTask(hashKey, packageEntity)
		err = ctx.Redirect(http.StatusMovedPermanently, "/tasks?key="+hashKey)
		return err
	}

	response := ctx.Response()

	pkgType := packageEntity.GetPkgType()

	if pkgType == constant.Tar {
		response.Header().Set(echo.HeaderContentType, "application/tar")
	} else {
		response.Header().Set(echo.HeaderContentType, "application/zip")
	}

	response.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", packageEntity.GetPkgName()))

	return Package(packageEntity, response, s.GetFileEntity, s.Storage.Download)
}

func (s *Server) checkTask(ctx echo.Context) error {
	key := ctx.QueryParam("key")
	if key == "" {
		err := ctx.String(http.StatusBadRequest, "没有指定key")
		if err != nil {
			return err
		}
	}
	var task = model.PackageTask{}
	err := s.TaskMeta.Get(key, &task)
	if err != nil {
		return err
	}

	if task.Status == model.TASK_SUCCESS {
		err := ctx.Redirect(http.StatusMovedPermanently, task.PackageRawKey)
		if err != nil {
			return err
		}
	} else if task.Status == model.TASK_AUTO {
		response := ctx.Response()
		response.Header().Set(echo.HeaderContentType, "application/zip")
		response.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", task.PackageInfo.GetPkgName()))

		return Package(task.PackageInfo, response, s.GetFileEntity, s.Storage.Download)
	} else {
		err := ctx.JSON(http.StatusOK, task)
		if err != nil {
			return err
		}
	}
	return nil
}
