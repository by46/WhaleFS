package utils

import (
	"fmt"
	"math/rand"
	"time"
)

const (
	CharSet = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	MaxID = 1000000000
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Timestamp() int64 {
	return time.Now().UTC().Unix()
}

func RandomAppID() string {
	id := rand.Int31n(MaxID)
	return fmt.Sprintf("app%09d", id)
}

func RandomAppSecretKey() string {
	b := make([]byte, 40)
	for i := range b {
		b[i] = CharSet[rand.Intn(len(CharSet))]
	}
	return string(b)
}
