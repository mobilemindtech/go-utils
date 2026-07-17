package assert

import (
	"fmt"
	"github.com/mobilemindtech/go-io/util"
	"strings"
)

func True(a bool, msg string, args ...interface{}) {
	if !a {
		panic(fmt.Sprintf(msg, args...))
	}
}

func False(a bool, msg string, args ...interface{}) {
	if a {
		panic(fmt.Sprintf(msg, args...))
	}
}

func Assert(a bool, msg string, args ...interface{}) {
	True(a, msg, args...)
}

func Assertf(f func() bool, msg string, args ...interface{}) {
	if !f() {
		panic(fmt.Sprintf(msg, args...))
	}
}

func NotEmpty(a string, msg string, args ...interface{}) {
	True(len(strings.TrimSpace(a)) > 0, msg, args...)
}

func Empty(a string, msg string, args ...interface{}) {
	True(len(strings.TrimSpace(a)) == 0, msg, args...)
}

func AssertNotEmpty(a string, msg string, args ...interface{}) {
	NotEmpty(a, msg, args...)
}

func NotNil(a any, msg string, args ...interface{}) {
	True(util.IsNotNil(a), msg, args...)
}

func Nil(a any, msg string, args ...interface{}) {
	True(util.IsNil(a), msg, args...)
}

func AssertNotNil(a any, msg string, args ...interface{}) {
	NotNil(a, msg, args...)
}
