package flatmap

import (
	"fmt"

	"github.com/mobilemindtec/go-utils/v2/optional"
)

type FlatMap[T any, R any] struct {
	errorHandler func(error)
}

func (this *FlatMap[T, R]) OnError(f func(error)) *FlatMap[T, R] {
	this.errorHandler = f
	return this
}

func (this *FlatMap[T, R]) Do(val T, fn func(T) optional.Optional[R]) optional.Optional[R] {

	if this.errorHandler != nil {
		defer func() {

			if err := recover(); err != nil {

				switch err.(type) {
				case error:
					this.errorHandler(err.(error))
					break
				default:
					this.errorHandler(fmt.Errorf("%v", err))
					break
				}
			}
		}()
	}

	return fn(val)
}
