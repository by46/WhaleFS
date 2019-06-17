package server

import (
	"encoding/json"
	"github.com/labstack/echo"
	"io"
	"net/http"
	"net/url"
)

type doc struct {
	Id  string      `json:"id"`
	Doc interface{} `json:"doc"`
}

func (s *Server) listBucket(ctx echo.Context) error {
	result, err := url.Parse(s.Config.BucketMeta)
	url := "http://" + result.Host + "/pools/default/buckets/basis/docs?skip=0&include_docs=true"
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		s.Logger.Errorf("请求couchbase api失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	password, _ := result.User.Password()
	req.SetBasicAuth(result.User.Username(), password)
	resp, err := client.Do(req)
	if err != nil {
		s.Logger.Errorf("请求couchbase api失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	response := ctx.Response()
	_, err = io.Copy(response, resp.Body)
	return err
}

func (s *Server) updateBucket(ctx echo.Context) error {
	bucket := doc{}
	f := ctx.Request().Body
	if err := json.NewDecoder(f).Decode(&bucket); err != nil {
		s.Logger.Errorf("Json 解析失败 %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}
	if err := s.BucketMeta.Set(bucket.Id, bucket.Doc); err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	return nil
}
