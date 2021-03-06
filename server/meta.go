package server

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
)

func (s *Server) GetFileEntity(hash string, isRemoveOriginal bool) (*model.FileMeta, error) {
	entity := &model.FileMeta{}
	if err := s.Meta.Get(hash, entity); err != nil {
		// 兼容JC legacy 文件系统
		if err == common.ErrKeyNotFound && s.Config.LegacyFS != "" {
			return s.GetFileEntityFromLegacy(hash, isRemoveOriginal)
		}
		return nil, err
	}
	return entity, nil
}

// 兼容JC legacy 文件系统
func (s *Server) GetFileEntityFromLegacy(hash string, isRemoveOriginal bool) (*model.FileMeta, error) {
	entity := &model.FileMeta{}
	bucket, key, size := s.parseBucketAndFileKey(hash)
	if bucket == nil {
		return nil, common.ErrKeyNotFound
	}
	fileContext := &model.FileContext{
		Bucket:     bucket,
		BucketName: bucket.Name,
		ObjectName: key[len(bucket.Name)+1:],
		Size:       size,
		Key:        key,
	}
	source := fmt.Sprintf("%s/%s%s", strings.Trim(s.Config.LegacyFS, "/"), bucket.Name, fileContext.ObjectName)
	if isRemoveOriginal {
		source = fmt.Sprintf("%s/%s/Original%s", strings.Trim(s.Config.LegacyFS, "/"), bucket.Name, fileContext.ObjectName)
	}

	if err := fileContext.ParseFileContent(source, nil); err != nil {
		s.Logger.Errorf("download file failed %v", err)
		return nil, common.ErrKeyNotFound
	}
	if fileContext.File.Digest == constant.DefaultImageDigest {
		return nil, common.ErrKeyNotFound
	}
	context := &middleware.ExtendContext{nil, fileContext}
	if fileContext.File.Size > constant.ChunkSize {
		_, err := s.uploadLargeFile(context)
		if err != nil {
			s.Logger.Errorf("upload large file", errors.WithStack(err))
			return nil, common.ErrKeyNotFound
		}
	} else {
		if _, err := s.uploadFileInternal(context); err != nil {
			return nil, common.ErrKeyNotFound
		}
	}
	err := s.Meta.Get(hash, entity)
	return entity, err
}
