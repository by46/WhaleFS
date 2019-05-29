package model

// seaweed fs 返回的文件数据
type Needle struct {
	FID          string
	ETag         string
	Size         int64
	LastModified int64
	Mime         string
}

func (n *Needle) AsFileMeta() *FileMeta {
	return &FileMeta{
		FID:          n.FID,
		ETag:         n.ETag,
		Size:         n.Size,
		LastModified: n.LastModified,
		MimeType:     n.Mime,
	}
}
