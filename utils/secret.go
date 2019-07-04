package utils

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
)

func Encode(policy []byte, appSecretKey string) string {
	encodingPolicy := base64.StdEncoding.EncodeToString(policy)
	mac := hmac.New(sha1.New, []byte(appSecretKey))
	mac.Write([]byte(encodingPolicy))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}
