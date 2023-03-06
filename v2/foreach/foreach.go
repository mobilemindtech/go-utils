package foreach

import (
	"fmt"

	"github.com/mobilemindtec/go-utils/v2/optional"
	"github.com/mobilemindtec/go-utils/v2/util"
)

type ForState int

const (
	ForDefault ForState = iota + 1
	ForSuccess
	ForBreak
	ForError
)

type Iterable interface {
	Execute()
	SetErrorHandler(interface{})
	GetResult() interface{}
}

type Next struct {
}

func DoNext() *Next {
	return &Next{}
}

type Break struct {
}

func DoBreak() *Break {
	return &Break{}
}

func TryNext(val interface{}) interface{} {

	if val == nil || optional.IsNilFixed(val) {
		return DoNext()
	}

	switch val.(type) {
	case error:
		return optional.NewFail(val.(error))
	case *optional.Fail:
		return val.(*optional.Fail)
	}

	return DoNext()
}

type Foreach[T any] struct {
	_for         func(int) *optional.Optional[[]T]
	each         func([]T) interface{}
	eachOne      func(T) interface{}
	errorHandler interface{}
	doneHandler  func()
	breakHandler func()
	results      []T
	by           int
	State        ForState
	fail         *optional.Fail
}

func New[T any]() *Foreach[T] {
	return &Foreach[T]{by: 1, State: ForDefault}
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

func (this *Foreach[T]) Each(f func([]T) interface{}) *Foreach[T] {
	this.each = f
	return this
}

func (this *Foreach[T]) EachOne(f func(T) interface{}) *Foreach[T] {
	this.eachOne = f
	return this
}

func (this *Foreach[T]) DoneHandler(f func()) *Foreach[T] {
	this.doneHandler = f
	return this
}

func (this *Foreach[T]) BreakHandler(f func()) *Foreach[T] {
	this.breakHandler = f
	return this
}

func (this *Foreach[T]) IsSuccess() bool {
	return this.State == ForSuccess
}

func (this *Foreach[T]) Join() interface{} {
	this.Do()
	switch this.State {
	case ForError:
		return this.fail
	}
	return DoNext()
}

func (this *Foreach[T]) Do() *Foreach[T] {

	defer func() {

		if r := recover(); r != nil {
			fmt.Println("Foreach recover: ", r)
			this.processError(r)
		}

	}()

	counter := 0

	if this.each == nil && this.eachOne == nil {
		panic("set each func")
	}

	bHandller := this.breakHandler

	this.breakHandler = func() {
		this.State = ForBreak
		if bHandller != nil {
			bHandller()
		}
	}

	onResult := func(r interface{}) bool {
		switch r.(type) {
		case bool, *Break:
			if !r.(bool) {
				if this.breakHandler != nil {
					this.breakHandler()
				}
			}
			break
		case *optional.Fail:
			this.processError(r.(*optional.Fail))
			break
		case *Next:
			return true
		default:
			this.processError(optional.NewFailStr("unknown return each option. use bool | Fail | Next | Break"))
			break
		}

		return false
	}

	for {

		r := this._for(counter)

		if !r.Any() {
			break
		}
		rs := r.Get()

		if len(rs) == 0 {
			break
		}

		for _, it := range rs {
			this.results = append(this.results, it)
		}

		if this.each != nil {

			if !onResult(this.each(rs)) {
				return this
			}

		} else if this.eachOne != nil {
			for _, it := range rs {

				if !onResult(this.eachOne(it)) {
					return this
				}
			}

		}

		counter += this.by
	}

	if this.doneHandler != nil {
		this.doneHandler()
	}

	this.State = ForSuccess

	return this
}

func (this *Foreach[T]) processError(r interface{}) {

	this.State = ForError
	this.fail = optional.NewFailStr("%v", r)

	if this.errorHandler != nil {
		switch this.errorHandler.(type) {
		case util.ErrorFn:
			this.errorHandler.(util.ErrorFn)(this.fail)
			break
		case util.FailFn:
			this.errorHandler.(util.FailFn)(this.fail)
			break
		}
	}
}

func (this *Foreach[T]) SetErrorHandler(errorHandler interface{}) {
	this.ErrorHandler(errorHandler)
}

func (this *Foreach[T]) ErrorHandler(errorHandler interface{}) *Foreach[T] {
	util.ValidateErrorHandler(errorHandler)
	this.errorHandler = errorHandler
	return this
}

func (this *Foreach[T]) GetResult() interface{} {
	return optional.NewSome(this.results)
}

func (this *Foreach[T]) Execute() {
	this.Do()
}
