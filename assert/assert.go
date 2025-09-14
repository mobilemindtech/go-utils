package assert

import (
	"fmt"
	"github.com/mobilemindtech/go-io/util"
	"strings"
)

func Assert(a bool, msg string, args ...interface{}) {
	if !a {
		panic(fmt.Sprintf(msg, args...))
	}
}

func Assertf(f func() bool, msg string, args ...interface{}) {
	if !f() {
		panic(fmt.Sprintf(msg, args...))
	}
}

func AssertNotEmpty(a string, msg string, args ...interface{}) {
	Assert(len(strings.TrimSpace(a)) > 0, msg, args...)
}

func AssertNotNil(a any, msg string, args ...interface{}) {
	Assert(util.IsNotNil(a), msg, args...)
}
