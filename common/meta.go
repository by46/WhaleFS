package common

type Meta interface {
	Get(key string, value interface{}) error
	Set(key string, value interface{}) error
	Exists(key string) (bool, error)
	SetTTL(key string, value interface{}, ttl int) error
}
