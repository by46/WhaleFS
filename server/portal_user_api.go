package server

import (
	"encoding/json"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"strings"
)

const (
	msgMustBeAdminRole = "只有admin用户才能操作"
)

type userCreate struct {
	Id      string      `json:"id"`
	Cas     uint64      `json:"cas,omitempty"`
	Version string      `json:"version"`
	User    *model.User `json:"basis,omitempty"`
}

type userBasisInfo struct {
	Id      string      `json:"id"`
	Cas     uint64      `json:"cas,omitempty"`
	Version string      `json:"version"`
	Basis   interface{} `json:"basis,omitempty"`
}

func (s *Server) addUser(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, msgMustBeAdminRole)
	}

	userCreateInfo := &userCreate{}
	body := ctx.Request().Body
	if err := json.NewDecoder(body).Decode(userCreateInfo); err != nil {
		s.Logger.Errorf("Json 解析失败 %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	u2 := new(model.User)
	err := s.BucketMeta.Get(userCreateInfo.Id, u2)
	if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "user 已经存在")
	}
	if err != common.ErrKeyNotFound {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	user := userCreateInfo.User
	user.Name = userCreateInfo.Id

	encryptPass, err := utils.Sha1(user.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	user.Password = encryptPass
	user.Type = typeUser
	user.Role = roleNormal
	user.Name = strings.TrimPrefix(user.Name, prefixUser)
	err = s.BucketMeta.Set(prefixUser+user.Name, user)

	return err
}

func (s *Server) getUser(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, msgMustBeAdminRole)
	}

	userId := ctx.Param("id")

	user := new(model.User)
	err := s.BucketMeta.Get(userId, user)
	if err != nil {
		if err == common.ErrKeyNotFound {
			return echo.NewHTTPError(http.StatusBadRequest, "user 不存在")
		}
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return ctx.JSON(http.StatusOK, user)
}

func (s *Server) listUser(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, msgMustBeAdminRole)
	}

	results, err := s.BucketMeta.GetAllUsers()

	if err != nil {
		s.Logger.Errorf("请求couchbase api失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var infos []*userBasisInfo
	for {
		info := new(userBasisInfo)
		if results.Next(info) == false {
			break
		}
		info.Version = strconv.FormatUint(info.Cas, 10)
		infos = append(infos, info)
	}

	err = ctx.JSON(http.StatusOK, infos)
	return err
}

func (s *Server) updateUser(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, msgMustBeAdminRole)
	}

	userUpdateInfo := &userCreate{}
	body := ctx.Request().Body
	if err := json.NewDecoder(body).Decode(userUpdateInfo); err != nil {
		s.Logger.Errorf("Json 解析失败 %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	user := new(model.User)
	err := s.BucketMeta.Get(userUpdateInfo.Id, user)
	if err != nil {
		if err == common.ErrKeyNotFound {
			return echo.NewHTTPError(http.StatusBadRequest, "user 不存在")
		}
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	user.Buckets = userUpdateInfo.User.Buckets

	err = s.BucketMeta.Set(prefixUser+user.Name, user)
	if err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (s *Server) deleteUser(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, msgMustBeAdminRole)
	}

	userId := strings.TrimPrefix(ctx.Request().URL.Path, "/api/users/")

	if !strings.HasPrefix(userId, prefixUser) {
		userId = prefixUser + userId
	}

	err := s.BucketMeta.Delete(userId, 0)
	if err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}
