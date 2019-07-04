package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPolicy_Encode(t *testing.T) {
	p := &Policy{
		Bucket:   "test",
		Deadline: time.Now().UTC().Add(10 * time.Minute).Unix(),
	}
	appId := "app_id"
	appSecretKey := "app_secret_key"
	sign := p.Encode(appId, appSecretKey)
	assert.Equal(t, "app_id:TfCgmTIDp4fL69TeQO0WXMjnfPU=:eyJidWNrZXQiOiJpdGVtIiwiZGVhZGxpbmUiOjE1NjIxNzA5ODh9", sign)
}

func TestPolicy_Decode(t *testing.T) {
	p := new(Policy)
	err := p.Decode("TfCgmTIDp4fL69TeQO0WXMjnfPU=", "app_secret_key", "eyJidWNrZXQiOiJpdGVtIiwiZGVhZGxpbmUiOjE1NjIxNzA5ODh9")
	assert.Nil(t, err)
	assert.Equal(t, "item", p.Bucket)
}
