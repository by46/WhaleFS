// Exception process logical
package common

import (
	"fmt"
)

const (
	CodeFileNotExists   = "FileNotExists"
	CodeBucketNotExists = "BucketNotExists"
	CodeForbidden       = "Forbidden"
	CodeLimit           = "Limit"
)

type BusinessError struct {
	Code       string
	StatusCode string
	Internal   error
}

type businessError struct {
	Code     int
	Message  interface{}
	Internal error
}

func (e *businessError) Error() string {
	return fmt.Sprintf("Business error code: %s, Interval error: %v", e.Code, e.Internal)
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("Business error code: %s, Interval error: %v", e.Code, e.Internal)
}

func New(code string) *BusinessError {
	return &BusinessError{Code: code}
}

func New2(code int, message interface{}, err error) error {
	return &businessError{
		Code:     code,
		Message:  message,
		Internal: err,
	}
}

func ErrorJson(err error) []byte {
	return nil
}
