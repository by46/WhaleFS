package model

type LogConfig struct {
	Level string `default:"error'`
	Home  string `default:"log"`
}

type StorageConfig struct {
	Cluster []string
}

type Config struct {
	Host               string `default:":8080"`
	Storage            *StorageConfig
	Master             []string
	Debug              bool `default:"false"`
	Log                *LogConfig
	Meta               string `default:"couchbase://scpodb01:8091/default"`
	MetaPassword       string
	BucketMeta         string `default:"couchbase://scpodb01:8091/buckets`
	BucketMetaPassword string
	TaskBucket         string
	TaskFileBucketName string
}
