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
	BucketHome                 string `default:"home"`
	CollectionTmp              string `default:"tmp"`
	CollectionTmpReplication   string `default:"000"`
	CollectionTmpTTL           TTL    `default:"7d"`
	CollectionShare            string `default:"mass"`
	CollectionShareReplication string `default:"100"`
	SizeDefault                string `default:"p200"`
}

type SyncConfig struct {
	Enable             bool `default:"false"`
	DFSHost            string
	RabbitMQConnection string
	RabbitMQExchange   string `default:"whale-fs"`
	RabbitMQQueue      string `default:"whale-fs"`
	QueueName          string `default:"whale-fs"`
	LegacyFSRoot       string
}

type PortalConfig struct {
	DfsHost string `json:"dfsHost"`
}

type ExtensionMappingConfig struct {
	Src  string `json:"src"`
	Dest string `json:"dest"`
}

type Config struct {
	Host                  string `default:":8080"`
	Storage               *StorageConfig
	LegacyFS              string
	Basis                 *BasisConfig
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
	Sync                  *SyncConfig
	Portal                *PortalConfig
	ExtensionMapping      []*ExtensionMappingConfig
}
