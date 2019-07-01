package constant

const (
	BucketPdt            = "pdt"
	VERSION              = "1.0.2"
	DefaultImageDigest   = "6f922092b63db2b3bd998666f589da6de6f54b63"
	QueryNameCollection  = "collection"
	QueryNameReplication = "replication"
	QueryNameTTL         = "ttl"
	MimeSize             = 512
	KeyBucket            = "system.bucket"
	KeyTimeStamp         = "system.timestamp"
	GzipScheme           = "gzip"
	GzipLimit            = 5 << 10          // 5K
	TTLChunk             = 60 * 60 * 24 * 7 //7 days
	FIDSep               = "|"
	ChunkSize            = 1024 * 1024 * 16 // 16M
	LengthEtag              = 14
)
