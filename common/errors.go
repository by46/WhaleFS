// Exception process logical
package common

import (
	"fmt"
)

const (
	CodeFileNotExists   = "FileNotExists"
	CodeBucketNotExists = "BucketNotExists"
	CodeForbidden       = "Forbidden"
)

type BusinessError struct {
	Code       string
	StatusCode string
	Internal   error
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("Business error code: %s, Interval error: %s", e.Code, e.Internal)
}

func New(code string) *BusinessError {
	return &BusinessError{Code: code}
}
