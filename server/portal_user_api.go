package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/labstack/echo"
	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
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

type AuthenticationClaims struct {
	Name string `json:"name"`
	jwt.StandardClaims
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

func (s *Server) login(ctx echo.Context) error {
	u := &struct {
		Name     string `json:"username"`
		Password string `json:"password"`
	}{}
	if err := ctx.Bind(u); err != nil {
		s.Logger.Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	user := &model.User{}
	err := s.BucketMeta.Get(prefixUser+u.Name, user)
	if err != nil {
		if err == common.ErrFileNotFound {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	sha1, err := utils.Sha1(u.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if user.Password != sha1 {
		return ctx.NoContent(http.StatusUnauthorized)
	}

	claims := AuthenticationClaims{
		user.Name,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24 * 30).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(s.Config.JwtSecretKey))

	if err != nil {
		s.Logger.Errorf("jwt sign failed: %+v", errors.WithStack(err))
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	info := struct {
		Token   string   `json:"token"`
		Name    string   `json:"username"`
		Buckets []string `json:"buckets"`
	}{signed, user.Name, user.Buckets,}
	return ctx.JSON(http.StatusOK, info)
}

func (s *Server) AuthenticateUser(token string) (*model.User, error) {
	t, err := jwt.ParseWithClaims(token, &AuthenticationClaims{}, func(token *jwt.Token) (i interface{}, e error) {
		return []byte(s.Config.JwtSecretKey), nil
	})
	if err != nil || t.Valid == false {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}
	claims, ok := t.Claims.(*AuthenticationClaims);
	if ok == false {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}
	key := fmt.Sprintf("%s.%s", constant.KeyUser, claims.Name)
	u := new(model.User)
	err = s.BucketMeta.Get(key, u)
	return u, errors.WithStack(err)
}

func (s *Server) getCurrentUser(ctx echo.Context) *model.User {
	return ctx.Get("user").(*model.User)
}
