package try

import (
	"github.com/mobilemindtec/go-utils/v2/optional"
)

func Try[T any](fn func() (T, error)) (opt *optional.Optional[T]) {

	defer func() {
		if err := recover(); err != nil {
			opt = optional.WithFail[T](err)
		}
	}()

	t, err := fn()
	return optional.TryMake[T](t, err)
}
