package try

import (
	"github.com/mobilemindtech/go-utils/v2/optional"
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

type Try[T any] struct {
	result *optional.Optional[T]
}

func (this Try[T]) Opt() *optional.Optional[T] {
	return this.result
}

func failure[T any](fail *optional.Fail) *Try[T] {
	return &Try[T]{optional.OfFail[T](fail)}
}

func New[T any](f func() (T, error)) *Try[T] {
	return &Try[T]{Of[T](f)}
}

func Then[T any, R any](t *Try[T], f func(T) (R, error)) *Try[R] {

	if t.result.IsSome() {
		return New[R](func() (R, error) {
			return f(t.result.UnWrap())
		})
	}
	return failure[R](t.result.GetFail())
}
