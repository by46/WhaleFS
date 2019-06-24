package server

import (
	"fmt"
	"github.com/by46/whalefs/common"
	"github.com/by46/whalefs/model"
	"github.com/by46/whalefs/server/middleware"
	"github.com/labstack/echo"
)

type chunkDigest struct {
	ChunkNo uint16 `json:"chunkNo"`
	Digest  string `json:"digest"`
}

type checkResult struct {
	Parts        []*model.Part `json:"parts,omitempty"`
	MissedDigest *chunkDigest  `json:"missedDigest,omitempty"`
}

func (s *Server) digestCheck(ctx echo.Context) (result *checkResult, err error) {
	context := ctx.(*middleware.ExtendContext)
	bucket := context.FileContext.Bucket

	digests := make([]*chunkDigest, 0)
	if err = ctx.Bind(&digests); err != nil {
		return nil, err
	}

	result = &checkResult{
		Parts: make([]*model.Part, 0),
	}

	chunkKey := fmt.Sprintf("chunks:%s", context.FileContext.UploadId)

	for _, digest := range digests {
		key := fmt.Sprintf("%s:%s", bucket.Basis.Collection, digest.Digest)
		chunk := new(model.Chunk)
		if err = s.ChunkDao.Get(key, chunk); err != nil {
			if err == common.ErrKeyNotFound {
				result.MissedDigest = digest
				return result, nil
			}
			return nil, err
		}
		part := &model.Part{
			PartNumber: int32(digest.ChunkNo),
			FID:        chunk.Fid,
		}
		result.Parts = append(result.Parts, part)
		if err = s.Meta.SubListAppend(chunkKey, "parts", part, 0); err != nil {
			return nil, err
		}
	}

	return
}
