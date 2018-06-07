package common

import (
	"sync"
	"time"
)

const (
	RFC1123 = "Mon, 02 Jan 2006 15:04:05 GMT"
)

var (
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, len(RFC1123))
		},
	}
)

func TimeToRFC822(dt time.Time) string {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)
	buf = dt.AppendFormat(buf, RFC1123)
	return string(buf)
}

func RFC822ToTime(value string) (time.Time, error) {
	return time.Parse(RFC1123, value)
}

func TimestampToRFC822(second int64) string {
	dt := time.Unix(second, 0).UTC()
	return TimeToRFC822(dt)
}
