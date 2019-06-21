package server

import (
	"encoding/json"
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
	Id      string     `json:"id"`
	Cas     uint64     `json:"cas,omitempty"`
	Version string     `json:"version"`
	User    model.User `json:"basis,omitempty"`
}

func (s *Server) addUser(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)
	if u.Role != roleAdmin {
		return echo.NewHTTPError(http.StatusForbidden, msgMustBeAdminRole)
	}

	basisInfo := &userCreate{}
	body := ctx.Request().Body
	if err := json.NewDecoder(body).Decode(basisInfo); err != nil {
		s.Logger.Errorf("Json 解析失败 %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	user := basisInfo.User
	user.Name = basisInfo.Id

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

	var infos []*basisInfo
	for {
		info := new(basisInfo)
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

	basisInfo := &userCreate{}
	body := ctx.Request().Body
	if err := json.NewDecoder(body).Decode(basisInfo); err != nil {
		s.Logger.Errorf("Json 解析失败 %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	user := basisInfo.User
	user.Name = basisInfo.Id

	user.Type = typeUser
	user.Role = roleNormal
	user.Name = strings.TrimPrefix(user.Name, prefixUser)
	err := s.BucketMeta.Set(prefixUser+user.Name, user)
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
