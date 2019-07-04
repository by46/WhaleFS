package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

func (s *Server) authenticate(ctx echo.Context, bucket *model.Bucket, token string) (err error) {
	if bucket.Protected == false {
		return nil
	}

	segments := strings.Split(token, ":")
	if len(segments) != 3 {
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	appId, sign, payload := segments[0], segments[1], segments[2]
	accessKey := new(model.AccessKey)
	if err := s.BucketMeta.Get(fmt.Sprintf("%s.%s", constant.KeyAccess, appId), accessKey); err != nil {
		if err != common.ErrKeyNotFound {
			s.Logger.Errorf("get access key failed: %+v", errors.WithStack(err))
		}
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	if accessKey.Expires <= time.Now().UTC().Unix() {
		return echo.NewHTTPError(http.StatusUnauthorized, "Access Key/Secret Key 过期")
	}
	appSecretKey := accessKey.AppSecretKey
	p := new(model.Policy)
	if err := p.Decode(sign, appSecretKey, payload); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, err.Error())
	}

	if accessKey.Owner == constant.UserAdmin {
		return nil
	}

	user := new(model.User)
	if err := s.BucketMeta.Get(fmt.Sprintf("%s.%s", constant.KeyUser, accessKey.Owner), user); err != nil {
		if err != common.ErrKeyNotFound {
			s.Logger.Errorf("get owner faied: %+v", errors.WithStack(err))
		}
		return echo.NewHTTPError(http.StatusUnauthorized)
	}
	if utils.Exists(user.Buckets, strings.ToLower(p.Bucket)) == false {
		// TODO(benjamin): 更准确的错误提示信息
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("无权操作该bucket"))
	}
	if accessKey.Scope != nil && utils.Exists(accessKey.Scope, strings.ToLower(p.Bucket)) == false {
		return echo.NewHTTPError(http.StatusUnauthorized, fmt.Sprintf("该AccessKey无权操作该bucket"))
	}
	return nil
}
