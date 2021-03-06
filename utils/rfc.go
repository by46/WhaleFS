package utils

import (
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/by46/whalefs/constant"
)

var (
	bufferPool = &sync.Pool{
		New: func() interface{} {
			return make([]byte, 0, len(constant.RFC1123))
		},
	}
)

func TimeToRFC822(dt time.Time) string {
	buf := bufferPool.Get().([]byte)
	defer bufferPool.Put(buf)
	buf = dt.AppendFormat(buf, constant.RFC1123)
	return string(buf)
}

func RFC822ToTime(value string) (time.Time, error) {
	return time.Parse(constant.RFC1123, value)
}

func TimestampToRFC822(second int64) string {
	dt := time.Unix(second, 0).UTC()
	return TimeToRFC822(dt)
}

func Name2Disposition(userAgent, name string) string {
	return fmt.Sprintf("attachment;filename=\"%s\";filename*=utf-8''%s", url.PathEscape(name), url.PathEscape(name))
}
