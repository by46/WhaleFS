// Exception process logical
package common

import (
	"fmt"
)

const (
	CodeFileNotExists = "FileNotExists"
)

type BusinessError struct {
	Code    string
	Message string
}

func (e *BusinessError) Error() string {
	return fmt.Sprintf("Business Error Code: %s, Message: %s", e.Code, e.Message)
}

func New(code string, message string) error {
	return &BusinessError{Code: code, Message: message}
}
