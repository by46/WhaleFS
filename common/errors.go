// Exception process logical
package common

import (
	"errors"
	"fmt"
)

var (
	ErrVolumeNotFound = errors.New("Volume not found")
	ErrFileNotFound   = errors.New("File not found")
	ErrKeyNotFound    = errors.New("Key not found")
	ErrKeyExists      = errors.New("Key exists")
)

type BusinessError struct {
	Code       string
	StatusCode string
	Internal   error
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("Business error code: %s, Interval error: %v", e.Code, e.Internal)
}

func New(code string) *BusinessError {
	return &BusinessError{Code: code}
}
