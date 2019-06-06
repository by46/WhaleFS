package server

const (
	VERSION           = "0.0.1"
	KeyBuckets        = "system.buckets"
	KeyBucket         = "system.bucket"
	GzipScheme        = "gzip"
	GzipLimit         = 5 << 10          // 5K
	TTLChunk          = 60 * 60 * 24 * 7 //7 days
	TTLTmp            = "7d"
	FIDSep            = "|"
	ReplicationOne    = "100"
)
