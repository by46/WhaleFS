package utils

import (
	"time"
)

func Timestamp() int64 {
	return time.Now().UTC().Unix()
}
