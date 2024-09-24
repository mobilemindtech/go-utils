package uerr

import "fmt"

type ValidationEmpty struct {
	Message string
}

func NewEmptyValidation(msg string, args ...interface{}) *ValidationEmpty {
	return &ValidationEmpty{Message: fmt.Sprintf(msg, args...)}
}

func (this *ValidationEmpty) Error() string {
	return this.Message
}

func (this *ValidationEmpty) String() string {
	return this.Message
}

func IsValidationEmpty(err error) bool {
	_, ok := err.(*ValidationEmpty)
	return ok
}
