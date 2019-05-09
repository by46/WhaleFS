package common

type Meta interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	SetTTL(key string, value interface{}, ttl int) error
}
