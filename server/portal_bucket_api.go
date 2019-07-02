package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo"
	"gopkg.in/couchbase/gocb.v1"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/utils"
)

const (
	typeBucket = "bucket"
	typeUser   = "user"
	typeToken  = "token"

	prefixBucket = "system.bucket."
	prefixUser   = "system.user."
	prefixToken  = "system.token."

	roleAdmin  = "admin"
	roleNormal = "normal"
)

type userInfo struct {
	Name    string   `json:"username"`
	Buckets []string `json:"buckets"`
	Token   string   `json:"token"`
}

type basisInfo struct {
	Id      string        `json:"id"`
	Cas     uint64        `json:"cas,omitempty"`
	Version string        `json:"version"`
	Basis   *model.Bucket `json:"basis,omitempty"`
}

func (s *Server) listBucket(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)

	var results gocb.QueryResults
	var err error
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

func (s *Server) getBucket(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)

	bucketId := ctx.Param("id")
	bucketName := strings.TrimPrefix(bucketId, prefixBucket)
	if u.Role != roleAdmin && !utils.Exists(u.Buckets, bucketName) {
		return echo.NewHTTPError(http.StatusForbidden, "用户没有权限操作此bucket")
	}

	b := new(model.Bucket)
	cas, err := s.BucketMeta.GetWithCas(bucketId, b)
	if err != nil {
		if err == common.ErrKeyNotFound {
			return echo.NewHTTPError(http.StatusBadRequest, "bucket 不存在")
		}
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	result := &basisInfo{
		Id:      prefixBucket + b.Name,
		Version: strconv.FormatUint(cas, 10),
		Basis:   b,
	}

	err = ctx.JSON(http.StatusOK, result)
	return nil
}

func (s *Server) addBucket(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)

	bucket := basisInfo{}
	f := ctx.Request().Body
	if err := json.NewDecoder(f).Decode(&bucket); err != nil {
		s.Logger.Errorf("Json 解析失败 %v", err)
		return echo.NewHTTPError(http.StatusBadRequest)
	}

	if !strings.HasPrefix(bucket.Id, prefixBucket) {
		bucket.Id = prefixBucket + bucket.Id
	}

	b := new(map[string]interface{})
	err := s.BucketMeta.Get(bucket.Id, b)
	if err == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket 已经存在")
	}
	if err != common.ErrKeyNotFound {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	if bucket.Basis.Type != typeBucket {
		return echo.NewHTTPError(http.StatusBadRequest, "type 不是 bucket")
	}

	if strings.TrimPrefix(bucket.Id, prefixBucket) != strings.TrimPrefix(bucket.Basis.Name, prefixBucket) {
		return echo.NewHTTPError(http.StatusBadRequest, "name 与 bucket id 不符")
	}

	if bucket.Basis.Basis == nil || bucket.Basis.Basis.Collection == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket collection 必须设置")
	}

	bucket.Basis.LastEditUser = u.Name
	bucket.Basis.LastEditDate = time.Now().Unix()

	if err := s.BucketMeta.Set(bucket.Id, bucket.Basis); err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	bucketName := strings.TrimPrefix(bucket.Id, prefixBucket)
	err = s.BucketMeta.SubListAppend(prefixUser+u.Name, "buckets", bucketName, 0)

	ts := &model.Timestamp{
		BucketUpdate: time.Now().Unix(),
	}
	if err = s.BucketMeta.Set(constant.KeyTimeStamp, ts); err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
}

func (s *Server) updateBucket(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)

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
	if u.Role != roleAdmin && !utils.Exists(u.Buckets, bucketName) {
		return echo.NewHTTPError(http.StatusForbidden, "用户没有权限操作此bucket")
	}

	preBucketMeta := new(model.Bucket)
	cas, err := s.BucketMeta.GetWithCas(bucket.Id, preBucketMeta)
	if err != nil {
		if err == common.ErrFileNotFound {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}
	if strconv.FormatUint(cas, 10) != bucket.Version {
		return echo.NewHTTPError(http.StatusConflict, "bucket 已被其他用户修改，请刷新重试")
	}

	if bucket.Basis.Type != typeBucket {
		return echo.NewHTTPError(http.StatusBadRequest, "type 不是 bucket")
	}

	if bucket.Basis.Name != preBucketMeta.Name {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket name 不能改变")
	}

	if bucket.Basis.Basis == nil {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket basis 必须存在")
	}

	if preBucketMeta.Basis.Collection != bucket.Basis.Basis.Collection {
		return echo.NewHTTPError(http.StatusBadRequest, "bucket collection 不能改变")
	}

	bucket.Basis.LastEditUser = u.Name
	bucket.Basis.LastEditDate = time.Now().Unix()

	if err := s.BucketMeta.Set(bucket.Id, bucket.Basis); err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	ts := &model.Timestamp{
		BucketUpdate: time.Now().Unix(),
	}
	if err = s.BucketMeta.Set(constant.KeyTimeStamp, ts); err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return err
}

func (s *Server) deleteBucket(ctx echo.Context) error {
	u := s.getCurrentUser(ctx)

	bucketId := ctx.Param("id")
	if !strings.HasPrefix(bucketId, prefixBucket) {
		bucketId = prefixBucket + bucketId
	}

	bucketName := strings.TrimPrefix(bucketId, prefixBucket)
	if u.Role != roleAdmin && !utils.Exists(u.Buckets, bucketName) {
		return echo.NewHTTPError(http.StatusForbidden, "用户没有权限操作此bucket")
	}

	err := s.BucketMeta.Delete(bucketId, 0)
	if err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	user := &model.User{}
	for {
		cas, err := s.BucketMeta.GetWithCas(prefixUser+u.Name, user)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var newBuckets []string
		for _, b := range user.Buckets {
			if bucketName != b {
				newBuckets = append(newBuckets, b)
			}
		}
		user.Buckets = newBuckets
		err = s.BucketMeta.SubSet(prefixUser+u.Name, "buckets", user.Buckets, cas)
		if err != nil {
			if err == gocb.ErrKeyExists {
				continue
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		break
	}

	ts := &model.Timestamp{
		BucketUpdate: time.Now().Unix(),
	}
	if err = s.BucketMeta.Set(constant.KeyTimeStamp, ts); err != nil {
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	return nil
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
		if err == common.ErrFileNotFound {
			return echo.NewHTTPError(http.StatusNotFound)
		}
		s.Logger.Errorf("数据库操作失败 %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError)
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
	u := s.getCurrentUser(ctx)

	authToken := ctx.Request().Header.Get("Authorization")
	authToken = strings.TrimPrefix(authToken, "Bearer ")

	err := s.BucketMeta.Delete(prefixToken+authToken, 0)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError)
	}

	user := &model.User{}
	for {
		cas, err := s.BucketMeta.GetWithCas(prefixUser+u.Name, user)
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError)
		}

		var newTokens []string
		for _, token := range user.Tokens {
			if authToken != token {
				newTokens = append(newTokens, token)
			}
		}
		user.Tokens = newTokens
		err = s.BucketMeta.SubSet(prefixUser+u.Name, "tokens", user.Tokens, cas)
		if err != nil {
			if err == gocb.ErrKeyExists {
				continue
			}
			return echo.NewHTTPError(http.StatusInternalServerError)
		}
		break
	}

	return nil
}

func (s *Server) listMimeTypes(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, utils.MimeTypes)
}

func (s *Server) configuration(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "[{\"key\":\"dfsHost\",\"value\":\"http://oss.yzw.cn.qa\"}]")
}
