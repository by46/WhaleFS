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

func (s *Server) createAccessKey(ctx echo.Context) (err error) {
	u := s.getCurrentUser(ctx)
	accessKey := &model.AccessKey{
		CreateDate: time.Now().UTC().Unix(),
		Expires:    time.Now().UTC().Add(365 * 24 * time.Hour).Unix(),
		Owner:      u.Name,
		Enable:     true,
	}

	for i := 0; i < 3; i++ {
		accessKey.AppKey = utils.RandomAppID()
		accessKey.AppSecretKey = utils.RandomAppSecretKey()
		key := fmt.Sprintf("%s.%s", constant.KeyAccess, accessKey.AppKey)
		_, err := s.BucketMeta.Insert(key, accessKey)
		if err != nil {
			if (err == common.ErrKeyExists) {
				continue
			}
			s.Logger.Errorf("create access key/secret key failed: %+v", errors.WithStack(err))
			return echo.NewHTTPError(http.StatusInternalServerError, "创建Access Key/Secret Keys失败")
		}
		break
	}
	key := fmt.Sprintf("%s.%s", constant.KeyUser, u.Name)
	user := new(model.User)
	cas, err := s.BucketMeta.GetWithCas(key, user)
	if err != nil {
		s.Logger.Errorf("get user(%s) failed: %+v", u.Name, errors.WithStack(err))
		key := fmt.Sprintf("%s.%s", constant.KeyAccess, accessKey.AppKey)
		err = s.BucketMeta.Delete(key, 0)
		s.Logger.Errorf("delete access key (%s) failed: %+v", accessKey.AppKey, errors.WithStack(err))
		return echo.NewHTTPError(http.StatusInternalServerError, "创建Access Key/Secret Keys失败")
	}
	user.AccessKeys = append(user.AccessKeys, accessKey.AppKey)
	// TODO(benjamin): process cas
	replaceCas2, err := s.BucketMeta.Replace(key, user, cas)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "创建Access Key/Secret Keys失败")
	}
	fmt.Sprintln("debug %v", replaceCas2)

	return ctx.JSON(http.StatusOK, accessKey)
}

func (s *Server) listAccessKey(ctx echo.Context) (err error) {
	u := s.getCurrentUser(ctx)
	key := fmt.Sprintf("%s.%s", constant.KeyUser, u.Name)
	user := new(model.User)
	_, err = s.BucketMeta.GetWithCas(key, user)
	if err != nil {
		s.Logger.Errorf("get user(%s) failed: %+v", errors.WithStack(err))
		return ctx.String(http.StatusOK, "[]")
	}
	if len(user.AccessKeys) == 0 {
		return ctx.String(http.StatusOK, "[]")
	}
	accessKeys := make([]*model.AccessKey, 0)
	for _, key := range user.AccessKeys {
		accessKey := new(model.AccessKey)
		if err := s.BucketMeta.Get(fmt.Sprintf("%s.%s", constant.KeyAccess, key), accessKey); err != nil {
			s.Logger.Errorf("get access key %s failed: %+v", key, errors.WithStack(err))
			continue
		}
		accessKeys = append(accessKeys, accessKey)
	}
	return ctx.JSON(http.StatusOK, accessKeys)
}

func (s *Server) updateAccessKey(ctx echo.Context) (err error) {
	u := s.getCurrentUser(ctx)

	inputAccessKey := new(model.InputAccessKey)
	if err := ctx.Bind(inputAccessKey); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	key := fmt.Sprintf("%s.%s", constant.KeyAccess, ctx.Param("key"))

	accessKey := new(model.AccessKey)
	cas, err := s.BucketMeta.GetWithCas(key, accessKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if accessKey.Owner != u.Name {
		return echo.NewHTTPError(http.StatusForbidden, "没有权限修改该AppKey")
	}
	accessKey.Expires = inputAccessKey.Expires
	accessKey.Enable = inputAccessKey.Enable
	accessKey.Scope = inputAccessKey.Scope

	_, err = s.BucketMeta.Replace(key, accessKey, cas)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "更新失败")
	}
	return
}

func (s *Server) deleteAccessKey(ctx echo.Context) (err error) {
	u := s.getCurrentUser(ctx)

	appKey := strings.ToLower(ctx.Param("key"))

	key := fmt.Sprintf("%s.%s", constant.KeyAccess, appKey)

	accessKey := new(model.AccessKey)
	cas, err := s.BucketMeta.GetWithCas(key, accessKey)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if accessKey.Owner != u.Name {
		return echo.NewHTTPError(http.StatusForbidden, "没有权限删除该AppKey")
	}

	err = s.BucketMeta.Delete(key, cas)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "删除AppKey失败")
	}
	user := new(model.User)
	key = fmt.Sprintf("%s.%s", constant.KeyUser, u.Name)
	cas, err = s.BucketMeta.GetWithCas(key, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "删除AppKey失败")
	}
	user.AccessKeys = utils.Filter(user.AccessKeys, func(item string) bool { return item != appKey })
	_, err = s.BucketMeta.Replace(key, user, cas)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "删除AppKey失败")
	}
	return nil
}

func (s *Server) generateToken(ctx echo.Context) (err error) {
	policy := new(model.Policy)
	if err := ctx.Bind(policy); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	accessKey := s.getAvailableAccessKey(ctx)

	if accessKey == nil {
		return echo.NewHTTPError(http.StatusForbidden, "没有AK/SK")
	}
	p := &struct{ Token string `json:"token"` }{Token: policy.Encode(accessKey.AppKey, accessKey.AppSecretKey)}
	return ctx.JSON(http.StatusOK, p)
}

func (s *Server) getAvailableAccessKey(ctx echo.Context) (*model.AccessKey) {
	u := s.getCurrentUser(ctx)
	key := fmt.Sprintf("%s.%s", constant.KeyUser, u.Name)
	user := new(model.User)
	_, err := s.BucketMeta.GetWithCas(key, user)
	if err != nil {
		s.Logger.Errorf("get user(%s) failed: %+v", errors.WithStack(err))
		return nil
	}
	for _, key := range user.AccessKeys {
		accessKey := new(model.AccessKey)
		if err := s.BucketMeta.Get(fmt.Sprintf("%s.%s", constant.KeyAccess, key), accessKey); err != nil {
			s.Logger.Errorf("get access key %s failed: %+v", key, errors.WithStack(err))
			continue
		}
		if accessKey.Available() {
			return accessKey
		}
	}
	return nil
}
