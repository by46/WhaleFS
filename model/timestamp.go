package model

// 用于记录配置的时间戳， 用于更新应用程序中的缓存信息
type Timestamp struct {
	BucketUpdate int64 `json:"bucket_update,omitempty"`
}
