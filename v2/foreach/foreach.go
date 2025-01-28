package foreach

import (

	"github.com/mobilemindtec/go-utils/v2/optional"
	"github.com/mobilemindtec/go-utils/v2/fn"
	"github.com/beego/beego/v2/core/logs"
	"runtime/debug"
	"reflect"
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
	filter       func(T) bool
	errorHandler interface{}
	doneHandler  func()
	breakHandler func()
	results      []T
	by           int
	State        ForState
	fail         *optional.Fail
	data         []T
}

func New[T any]() *Foreach[T] {
	return &Foreach[T]{by: 1, State: ForDefault}
}

func Of[T any](data []T) *Foreach[T] {
	return &Foreach[T]{by: 1, data: data, State: ForDefault}
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

func (this *Foreach[T]) Filter(f func(T) bool) *Foreach[T] {
	this.filter = f
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
			logs.Error("Foreach recover. StackTrae: %v", r, string(debug.Stack()))
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

		useFor := this._for != nil

		if useFor {
			r := this._for(counter)

			if r.IsFail() {
				this.processError(r.GetFail())
				return this
			}

			if !r.Any() {
				break
			}


			this.data = r.Get()
		}

		if len(this.data) == 0 {
			break
		}

		for _, it := range this.data {
			if this.filter != nil {
				if !this.filter(it) {
					continue
				}
			}
			this.results = append(this.results, it)
		}

		if this.each != nil {

			if !onResult(this.each(this.data)) {
				return this
			}

		} else if this.eachOne != nil {
			for _, it := range this.data {

				if !onResult(this.eachOne(it)) {
					return this
				}
			}

		}

		if !useFor {
			break
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

		info := fn.NewFuncInfo(this.errorHandler)

		if info.ArgsCount == 0 {
			info.CallEmpty()
		} else {

			if info.ArgsCount != 1  { // onSuccess = func(interface{})
				panic("step OnError: func must have one argument of error or Fail")
			}

			args := []reflect.Value{}
			if info.ArgsTypes[0] == reflect.TypeOf(this.fail.Error) {
				args = append(args, reflect.ValueOf(this.fail.Error))
			}else {
				args = append(args, reflect.ValueOf(this.fail))
			}

			info.Call(args)

		}
	}
}

func (this *Foreach[T]) SetErrorHandler(errorHandler interface{}) {
	this.ErrorHandler(errorHandler)
}

func (this *Foreach[T]) ErrorHandler(errorHandler interface{}) *Foreach[T] {
	this.errorHandler = errorHandler
	return this
}

func (this *Foreach[T]) GetResult() interface{} {
	return optional.NewSome(this.results)
}

func (this *Foreach[T]) Execute() {
	this.Do()
}
