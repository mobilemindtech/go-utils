package util

import (
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type FailFn = func(*optional.Fail)
type ErrorFn = func(interface{})

func ValidateErrorHandler(v interface{}) {
	if v != nil {
		switch v.(type) {
		case ErrorFn:
		case FailFn:
			//ok
			break
		default:
			panic("wrong error hanlder")
		}
	}
}
