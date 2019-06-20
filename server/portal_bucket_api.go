package server

import (
	"encoding/json"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo"
	"gopkg.in/couchbase/gocb.v1"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	typeBucket = "bucket"
	typeUser   = "user"
	typeToken  = "token"

	prefixBucket = "system.bucket."
	prefixUser   = "system.user."
	prefixToken  = "system.token."

	roleAdmin = "admin"
)

type userInfo struct {
	Name    string   `json:"username"`
	Buckets []string `json:"buckets"`
	Token   string   `json:"token"`
}

type basisInfo struct {
	Id      string      `json:"id"`
	Cas     uint64      `json:"cas,omitempty"`
	Version string      `json:"version"`
	Basis   interface{} `json:"basis,omitempty"`
}

func (s *Server) listBucket(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}

	var results gocb.QueryResults
	if u.Role == roleAdmin {
		results, err = s.BucketMeta.GetAllBuckets()
	} else {
		results, err = s.BucketMeta.GetBucketsByNames(u.Buckets)
	}

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

func (s *Server) addBucket(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}

	bucket := basisInfo{}
	f := ctx.Request().Body
	if err := json.NewDecoder(f).Decode(&bucket); err != nil {
		s.Logger.Errorf("Json 解析失败 %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if !strings.HasPrefix(bucket.Id, prefixBucket) {
		bucket.Id = prefixBucket + bucket.Id
	}

	var b interface{}
	if err = s.BucketMeta.Get(bucket.Id, b); err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket 已经存在")
	}
	if err != common.ErrKeyNotFound {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	basis := bucket.Basis.(map[string]interface{})
	value, ok := basis["type"]
	if ok {
		if value.(string) != typeBucket {
			return echo.NewHTTPError(http.StatusBadRequest, "type 不是 bucket")
		}
	} else {
		basis["type"] = typeBucket
	}

	name, ok := basis["name"]
	if ok {
		if strings.TrimPrefix(bucket.Id, prefixBucket) != strings.TrimPrefix(name.(string), prefixBucket) {
			return echo.NewHTTPError(http.StatusBadRequest, "name 与 bucket id 不符")
		}
	} else {
		basis["name"] = strings.TrimPrefix(bucket.Id, prefixBucket)
	}

	if err := s.BucketMeta.Set(bucket.Id, bucket.Basis); err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	bucketName := strings.TrimPrefix(bucket.Id, prefixBucket)
	if !utils.Exists(u.Buckets, bucketName) {
		u.Buckets = append(u.Buckets, bucketName)
	}
	err = s.BucketMeta.SubSet(prefixUser+u.Name, "buckets", u.Buckets, 0)

	return nil
}

func (s *Server) updateBucket(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}

	bucket := basisInfo{}
	f := ctx.Request().Body
	if err := json.NewDecoder(f).Decode(&bucket); err != nil {
		s.Logger.Errorf("Json 解析失败 %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if !strings.HasPrefix(bucket.Id, prefixBucket) {
		bucket.Id = prefixBucket + bucket.Id
	}

	bucketName := strings.TrimPrefix(bucket.Id, prefixBucket)
	if !utils.Exists(u.Buckets, bucketName) {
		return echo.NewHTTPError(http.StatusForbidden, "用户没有权限操作此bucket")
	}

	var b interface{}
	cas, err := s.BucketMeta.GetWithCas(bucket.Id, b)
	if err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if strconv.FormatUint(cas, 10) != bucket.Version {
		return echo.NewHTTPError(http.StatusConflict, "bucket 已被其他用户修改，请刷新重试")
	}

	basis := bucket.Basis.(map[string]interface{})
	value, ok := basis["type"]
	if ok {
		if value != typeBucket {
			return echo.NewHTTPError(http.StatusBadRequest, "type 不是 bucket")
		}
	} else {
		basis["type"] = typeBucket
	}

	name, ok := basis["name"]
	if ok {
		if strings.TrimPrefix(bucket.Id, prefixBucket) != strings.TrimPrefix(name.(string), prefixBucket) {
			return echo.NewHTTPError(http.StatusBadRequest, "name 与 bucket id 不符")
		}
	} else {
		basis["name"] = strings.TrimPrefix(bucket.Id, prefixBucket)
	}

	if err := s.BucketMeta.Set(bucket.Id, bucket.Basis); err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return err
}

func (s *Server) deleteBucket(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}

	bucketId := strings.TrimPrefix(ctx.Request().URL.Path, "/api/buckets/")

	if !strings.HasPrefix(bucketId, prefixBucket) {
		bucketId = prefixBucket + bucketId
	}

	bucketName := strings.TrimPrefix(bucketId, prefixBucket)
	if !utils.Exists(u.Buckets, bucketName) {
		return echo.NewHTTPError(http.StatusForbidden, "用户没有权限操作此bucket")
	}

	err = s.BucketMeta.Delete(bucketId, 0)
	if err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	var newBuckets []string
	for _, b := range u.Buckets {
		if bucketName != b {
			newBuckets = append(newBuckets, b)
		}
	}
	u.Buckets = newBuckets
	err = s.BucketMeta.SubSet(prefixUser+u.Name, "buckets", u.Buckets, 0)

	return err
}

func (s *Server) login(ctx echo.Context) error {
	reqUser := new(model.User)
	if err := ctx.Bind(reqUser); err != nil {
		s.Logger.Error(err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	var user = &model.User{}
	err := s.BucketMeta.Get(prefixUser+reqUser.Name, user)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound)
	}

	sha1, err := utils.Sha1(reqUser.Password)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if user.Password == sha1 {
		userToken := uuid.New().String()
		expires := time.Now().Add(30 * 24 * time.Hour)

		newToken := &model.Token{
			Value:   userToken,
			Expires: expires,
			UserId:  prefixUser + reqUser.Name,
			Type:    typeToken,
		}
		err := s.BucketMeta.Set(prefixToken+userToken, newToken)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		user.Tokens = append(user.Tokens, userToken)
		err = s.BucketMeta.Set(prefixUser+reqUser.Name, user)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		uInfo := &userInfo{
			Name:    user.Name,
			Buckets: user.Buckets,
			Token:   userToken,
		}
		return ctx.JSON(http.StatusOK, uInfo)
	}
	return ctx.NoContent(http.StatusUnauthorized)
}

func (s *Server) logout(ctx echo.Context) error {
	u, err := s.GetUser(ctx)
	if err != nil {
		return err
	}

	authToken := ctx.Request().Header.Get("Authorization")
	authToken = strings.TrimPrefix(authToken, "Bearer ")

	err = s.BucketMeta.Delete(prefixToken+authToken, 0)
	if err != nil {
		return err
	}

	var newTokens []string
	for _, token := range u.Tokens {
		if authToken != token {
			newTokens = append(newTokens, token)
		}
	}
	u.Tokens = newTokens
	err = s.BucketMeta.SubSet(prefixUser+u.Name, "tokens", u.Tokens, 0)

	if err != nil {
		return err
	}

	return err
}

func (s *Server) GetUser(ctx echo.Context) (*model.User, error) {
	authToken := ctx.Request().Header.Get("Authorization")
	authToken = strings.TrimPrefix(authToken, "Bearer ")
	t := &model.Token{}
	err := s.BucketMeta.Get(prefixToken+authToken, t)
	if err != nil {
		if err == common.ErrKeyNotFound {
			return nil, echo.NewHTTPError(http.StatusUnauthorized)
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	if t.Expires.Before(time.Now()) {
		return nil, echo.NewHTTPError(http.StatusUnauthorized)
	}

	u := &model.User{}
	err = s.BucketMeta.Get(t.UserId, u)
	if err != nil {
		if err == common.ErrKeyNotFound {
			return nil, echo.NewHTTPError(http.StatusNotFound)
		}
		return nil, echo.NewHTTPError(http.StatusInternalServerError)
	}
	return u, nil
}
