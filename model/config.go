package model

type LogConfig struct {
	Level string `default:"error'`
	Home  string `default:"log"`
}

type StorageConfig struct {
	Cluster []string
}

type CollectionConfig struct {
	Tmp   string `default:"tmp"`
	Share string `default:"mass"`
}

type BasisConfig struct {
	CollectionTmp   string `default:"tmp"`
	CollectionShare string `default:"mass"`
}

type Config struct {
	Host                  string `default:":8080"`
	Basis                 *BasisConfig
	Storage               *StorageConfig
	Master                []string
	Debug                 bool `default:"false"`
	Log                   *LogConfig
	Meta                  string `default:"http://localhost:8091/default"`
	BucketMeta            string `default:"http://localhost:8091/buckets"`
	ChunkMeta             string `default:"http://localhost:8091/chunks"`
	TaskBucket            string
	TaskFileBucketName    string
	TaskFileSizeThreshold int64
	HttpClientBase        string
	TempFileDir           string
}
