package model

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	"github.com/by46/whalefs/utils"
)

type Policy struct {
	Bucket   string `json:"bucket"`
	Deadline int64  `json:"deadline"`
}

func (p *Policy) Decode(sign, appSecretKey, payload string) (err error) {
	content, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return fmt.Errorf("策略格式错误")
	}
	sign2 := utils.Encode(content, appSecretKey)
	if sign != sign2 {
		return fmt.Errorf("权限凭证错误")
	}
	if err = json.Unmarshal(content, p); err != nil {
		return fmt.Errorf("策略格式JSON错误")
	}
	if p.Deadline <= time.Now().UTC().Unix() {
		return fmt.Errorf("Token过期")
	}
	return nil
}

func (p *Policy) Encode(appId, appSecretKey string) string {
	content, _ := json.Marshal(p)
	encodedSign := utils.Encode(content, appSecretKey)
	return fmt.Sprintf("%s:%s:%s", appId, encodedSign, base64.StdEncoding.EncodeToString(content))
}
