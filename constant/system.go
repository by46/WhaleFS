package constant

const (
	BucketPdt            = "pdt"
	VERSION              = "0.0.1"
	DefaultImageDigest   = "6f922092b63db2b3bd998666f589da6de6f54b63"
	QueryNameCollection  = "collection"
	QueryNameReplication = "replication"
	QueryNameTTL         = "ttl"
	MimeSize             = 512
	KeyBucket            = "system.bucket"
	GzipScheme           = "gzip"
	GzipLimit            = 5 << 10          // 5K
	TTLChunk             = 60 * 60 * 24 * 7 //7 days
	FIDSep               = "|"
)
