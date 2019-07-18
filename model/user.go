package model

import (
	"time"
)

type User struct {
	Name      string   `json:"username"`
	Password  string   `json:"password"`
	Buckets   []string `json:"buckets,omitempty"` //不加system.bucket 前缀
	Type      string   `json:"type"`
	Tokens    []string `json:"tokens,omitempty"` //不加system.token 前缀
	Role      string   `json:"role"`
	AccessKeys []string `json:"access_keys,omitempty"` // 不加system.access前缀
}

type Token struct {
	Value   string    `json:"value"`
	Expires time.Time `json:"expires"`
	UserId  string    `json:"userId"`
	Type    string    `json:"type"`
}
