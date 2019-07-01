package model

import (
	"sort"
	"strings"

	"github.com/by46/whalefs/constant"
	"github.com/by46/whalefs/utils"
)

type Uploads struct {
	Bucket   string `json:"bucket"`
	Key      string `json:"key"`
	UploadId string `json:"uploadId"`
}

type Part struct {
	PartNumber int32  `json:"partNumber"`
	FID        string `json:"fid"`
	Size       int64  `json:"size"`
}

type Parts []*Part

func (p Parts) Len() int {
	return len(p)
}

func (p Parts) Less(i, j int) bool {
	return p[i].PartNumber < p[j].PartNumber
}

func (p Parts) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}

type PartMeta struct {
	Key          string `json:"key"`
	MimeType     string `json:"mimeType"`
	IsRandomName bool   `json:"is_random_name"`
	Parts        Parts  `json:"parts"`
	ThumbnailKey string `json:"thumbnailKey,omitempty"`
}

func (p *PartMeta) AsFileMeta() *FileMeta {
	meta := new(FileMeta)
	size := int64(0)
	if p.Parts != nil {
		sort.Sort(p.Parts)
		segments := make([]string, len(p.Parts))
		for i, part := range p.Parts {
			size += part.Size
			segments[i] = part.FID
		}
		meta.FID = strings.Join(segments, "|")
	}
	meta.Size = size
	meta.MimeType = p.MimeType
	meta.LastModified = utils.Timestamp()
	meta.ETag = utils.Sha1WithLength(meta.FID, constant.LengthEtag)
	return meta
}
