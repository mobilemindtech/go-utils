package lazy

import (
	"github.com/mobilemindtec/go-utils/v2/optional"
)

// Lazy express a lazy function
type Lazy struct {
	Exec func() interface{}
}

func New(exec func() interface{}) *Lazy {
	return &Lazy{exec}
}

// RunAllMap Run all lazy expressions, map to fn result if success
func RunMap[T any](fn func() T, opts ...*Lazy) *optional.Optional[T] {
	r := Run(opts...)

	if r.IsSome() {
		return optional.Of[T](fn())
	}

	return optional.Of[T](r.Val())
}

// RunAll Run all lazy expressions, map to Optional[bool](true) if success
func Run(opts ...*Lazy) *optional.Optional[bool] {
	for _, lazy := range opts {
		r := lazy.Exec()
		switch r.(type) {
		case *optional.Some:
			continue
		default:
			return optional.Of[bool](r)
		}
	}
	return optional.Of[bool](true)
}
