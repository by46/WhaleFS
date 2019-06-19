package model

import (
	"time"
)

type User struct {
	Name     string   `json:"username"`
	Password string   `json:"password"`
	Buckets  []string `json:"buckets"`
	Type     string   `json:"type"`
	Tokens   []string `json:"tokens"`
	Role     string   `json:"role"`
}

type Token struct {
	Value   string    `json:"value"`
	Expires time.Time `json:"expires"`
	UserId  string    `json:"userId"`
	Type    string    `json:"type"`
}
