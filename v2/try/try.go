package try

import (
	"github.com/mobilemindtec/go-utils/v2/optional"
)

func Of[T any](fn func() (T, error)) (opt *optional.Optional[T]) {

	defer func() {
		if err := recover(); err != nil {
			opt = optional.OfFail[T](err)
		}
	}()

	t, err := fn()
	return optional.Try[T](t, err)
}
