package assert

import (
	"fmt"
	"strings"
	"github.com/mobilemindtec/go-io/util"
)

func Asssert(a bool, msg string, args ...interface{}) {
	if !a {
		panic(fmt.Sprintf(msg, args...))
	}
}

func Asssertf(f func() bool, msg string, args ...interface{}) {
	if !f() {
		panic(fmt.Sprintf(msg, args...))
	}
}

func AsssertNotEmpty(a string, msg string, args ...interface{}) {
	Asssert(len(strings.TrimSpace(a)) > 0, msg, args...)
}


func AsssertNotNil(a any, msg string, args ...interface{}) {
	Asssert(util.IsNotNil(a), msg, args...)
}
