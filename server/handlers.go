package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo"

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

func (s *Server) tarDownload(ctx echo.Context) error {
	content := ctx.FormValue("content")
	tarFileEntity := new(model.TarFileEntity)
	err := json.Unmarshal([]byte(content), &tarFileEntity)
	if err != nil {
		return err
	}

	var totalSize int64
	for _, item := range tarFileEntity.Items {
		hashKey, err := utils.Sha1(item.RawKey)
		if err != nil {
			return err
		}

		entity, err := s.GetFileEntity(hashKey)
		if err != nil {
			return err
		}

		totalSize = totalSize + entity.Size
	}

	if totalSize > 1024 {
		hashKey, err := utils.Sha1(fmt.Sprintf("/%s/%s", s.TaskBucketName, tarFileEntity.Name))
		if err != nil {
			return err
		}

		err = s.CreateTask(hashKey, tarFileEntity)
		if err != nil {
			return err
		}
		ctx.Redirect(http.StatusMovedPermanently, "/tasks?key="+hashKey)
	}

	response := ctx.Response()
	response.Header().Set(echo.HeaderContentType, "application/tar")
	response.Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("attachment; filename=%s", tarFileEntity.Name))

	return Package(tarFileEntity, response, s.GetFileEntity, s.Storage.Download)
}

func (s *Server) checkTask(ctx echo.Context) error {
	key := ctx.QueryParam("key")
	if key == "" {
		err := ctx.String(http.StatusBadRequest, "没有指定key")
		if err != nil {
			return err
		}
	}
	var task = model.TarTask{}
	err := s.TaskMeta.Get(key, &task)
	if err != nil {
		return err
	}
	if task.Status == model.TASK_SUCCESS {
		err := ctx.Redirect(http.StatusMovedPermanently, task.TarFileRawKey)
		if err != nil {
			return err
		}
	} else if task.Status == model.TASK_PENDING || task.Status == model.TASK_RUNNING {
		err := ctx.String(http.StatusOK, "文件打包中......")
		if err != nil {
			return err
		}
	} else {
		err := ctx.String(http.StatusInternalServerError, "文件打包失败"+task.ErrorMsg)
		if err != nil {
			return err
		}
	}
	return nil
}
