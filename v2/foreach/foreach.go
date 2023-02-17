package foreach

import (
	"fmt"

	"github.com/mobilemindtec/go-utils/v2/optional"
	"github.com/mobilemindtec/go-utils/v2/util"
)

type Iterable interface {
	Execute()
	ErrorHandler(interface{})
	GetResult() interface{}
}

type Foreach[T any] struct {
	_for         func(int) *optional.Optional[[]T]
	each         func([]T) bool
	eachOne      func(T) bool
	errorHandler interface{}
	doneHandler  func()
	results      []T
	by           int
}

func New[T any]() *Foreach[T] {
	return &Foreach[T]{by: 1}
}

func (this *Foreach[T]) GetResults() []T {
	return this.results
}

func (this *Foreach[T]) By(i int) *Foreach[T] {
	this.by = i
	return this
}

func (this *Foreach[T]) For(f func(int) *optional.Optional[[]T]) *Foreach[T] {
	this._for = f
	return this
}

func (this *Foreach[T]) Each(f func([]T) bool) *Foreach[T] {
	this.each = f
	return this
}

func (this *Foreach[T]) EachOne(f func(T) bool) *Foreach[T] {
	this.eachOne = f
	return this
}

func (this *Foreach[T]) DoneHandler(f func()) *Foreach[T] {
	this.doneHandler = f
	return this
}

func (this *Foreach[T]) Do() *Foreach[T] {

	defer func() {

		if r := recover(); r != nil {
			fmt.Println("Foreach recover: ", r)

			if this.errorHandler != nil {
				switch this.errorHandler.(type) {
				case util.ErrorFn:
					this.errorHandler.(util.ErrorFn)(optional.NewFailStr("%v", r))
					break
				case util.FailFn:
					this.errorHandler.(util.FailFn)(optional.NewFailStr("%v", r))
					break
				}
			}

		}

	}()

	counter := 0

	if this.each == nil && this.eachOne == nil {
		panic("set each func")
	}

	for {

		r := this._for(counter)

		if r.Any() {
			rs := r.Get()

			if len(rs) == 0 {
				break
			}

			for _, it := range rs {
				this.results = append(this.results, it)
			}

			if this.each != nil {
				if !this.each(rs) {
					break
				}
			} else if this.eachOne != nil {
				for _, it := range rs {
					if !this.eachOne(it) {
						break
					}
				}
			}
		} else {
			break
		}

		counter += this.by
	}

	if this.doneHandler != nil {
		this.doneHandler()
	}

	return this
}

func (this *Foreach[T]) ErrorHandler(errorHandler interface{}) {
	util.ValidateErrorHandler(errorHandler)
	this.errorHandler = errorHandler
}

func (this *Foreach[T]) GetResult() interface{} {
	return optional.NewSome(this.results)
}

func (this *Foreach[T]) Execute() {
	this.Do()
}
