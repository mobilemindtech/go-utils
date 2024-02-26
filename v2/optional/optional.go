package optional

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	lst "github.com/mobilemindtec/go-utils/lists"
	"github.com/mobilemindtec/go-utils/v2/lists"
)

type Optional[T any] struct {
	some *Some
	none *None
	fail *Fail
}

func OfSome[T any](v interface{}) *Optional[T] {

	var s *Some

	switch v.(type) {
	case *Some:
		s = v.(*Some)
		break
	default:
		s = NewSome(v)
	}

	return &Optional[T]{some: s}
}

func OfFail[T any](v interface{}) *Optional[T] {

	var s *Fail

	switch v.(type) {
	case *Fail:
		s = v.(*Fail)
		break
	case error:
		s = NewFail(v.(error))
		break
	case string:
		s = NewFail(errors.New(v.(string)))
		break
	default:
		s = NewFail(errors.New("error"))
		break
	}

	return &Optional[T]{fail: s}
}

func OfNone[T any]() *Optional[T] {
	return &Optional[T]{none: NewNone()}
}

// OfOk represents ok empty result or so ignore result
func OfOk[T any]() *Optional[T] {
	return &Optional[T]{some: SomeOk()}
}

func Try[T any](val interface{}, err interface{}) *Optional[T] {
	if !IsNilFixed(err) {
		return New[T](err)
	}
	return New[T](val)
}

func Of[T any](val interface{}) *Optional[T] {
	return New[T](val)
}

func New[T any](val interface{}) *Optional[T] {

	opt := Optional[T]{}

	if IsNilFixed(val) {
		opt.none = NewNone()
		return &opt
	}

	switch val.(type) {
	case *Some:
		opt.some = val.(*Some)
		break
	case *None:
		opt.none = val.(*None)
		break
	case *Fail:
		opt.fail = val.(*Fail)
		break
	case error:
		opt.fail = NewFail(val.(error))
		break
	default:
		mkd := MakeTry(val, nil)
		switch mkd.(type) {
		case *Some:
			opt.some = mkd.(*Some)
			break
		case *None:
			opt.none = mkd.(*None)
			break
		case *Fail:
			opt.fail = mkd.(*Fail)
			break
		default:
			panic(fmt.Sprintf("can't get type from: %v", mkd))
		}
		break
	}
	return &opt
}

func (this *Optional[T]) PanicIfFail(msg ...string) *Optional[T] {
	if this.IsFail() {

		m := ""

		if len(msg) > 0 {
			m = msg[0]
		}

		m = fmt.Sprintf(": %v", this.GetFail().ErrorString())

		panic(m)
	}

	return this
}

func (this *Optional[T]) PanicIfNone(msg ...string) *Optional[T] {
	if this.IsNone() {

		m := ""

		if len(msg) > 0 {
			m = msg[0]
		}

		m = fmt.Sprintf(": None type")

		panic(m)
	}

	return this
}

func (this *Optional[T]) PanicIfNotSome(msg ...string) *Optional[T] {
	this.PanicIfFail(msg...)
	this.PanicIfNone(msg...)
	return this
}

// Try fail only if a Fail is returned
func (this *Optional[T]) Try(f func(T) interface{}) *Optional[T] {
	if this.IsSome() {

		r := f(this.some.Item.(T))

		if val, ok := TryExtractValIfOptional(r); ok {
			r = val
		}

		switch r.(type) {
		case *Fail:
			return Of[T](r)
		}
	}
	return this
}

func (this *Optional[T]) GetFail() *Fail {
	return this.fail
}

func (this *Optional[T]) GetSome() *Some {
	return this.some
}

func (this *Optional[T]) OrElse(v T) T {
	return GetOrElse[T](this.some, v)
}

func (this *Optional[T]) UnWrap() T {
	return this.Get()
}
func (this *Optional[T]) Get() T {
	return Get[T](this.some.Item)
}

func (this *Optional[T]) Any() bool {
	return ! IsNilFixed(this.some)
}

func (this *Optional[T]) Fail() bool {
	return !IsNilFixed(this.fail)
}

func (this *Optional[T]) Empty() bool {
	return !IsNilFixed(this.none)
}

func (this *Optional[T]) NonEmpty() bool {
	return !this.IsNone() && !this.IsFail()
}

func (this *Optional[T]) IsSome() bool {
	return !IsNilFixed(this.some)
}

func (this *Optional[T]) IsFail() bool {
	return !IsNilFixed(this.fail)
}

func (this *Optional[T]) IsNone() bool {
	return !IsNilFixed(this.none)
}

func (this *Optional[T]) Val() interface{} {
	if this.IsSome() {
		return this.some
	} else if this.IsNone() {
		return this.none
	} else if this.IsFail() {
		return this.fail
	} else {
		return NewNone()
	}
}

func (this *Optional[T]) IfFail(cb func(error)) *Optional[T] {
	if this.IsFail() {
		cb(this.fail.Error)
	}
	return this
}

func (this *Optional[T]) IfSome(cb func(T)) *Optional[T] {
	if this.IsSome() {
		cb(GetItem[T](this.some))
	}
	return this
}

func (this *Optional[T]) IfNone(cb func()) *Optional[T] {
	if this.IsNone() {
		cb()
	}
	return this
}

func (this *Optional[T]) IfNonEmpty(cb func(T)) *Optional[T] {
	if this.IsSome() {
		cb(GetItem[T](this.some))
	}
	return this
}

func (this *Optional[T]) IfNonEmptyOrElse(cb func(T), emptyCb func()) *Optional[T] {
	if this.IsSome() {
		cb(GetItem[T](this.some))
	} else {
		emptyCb()
	}
	return this
}

func (this *Optional[T]) Filter(filter func(T) bool) *Optional[T] {
	if this.some != nil {
		v := GetItem[T](this.some)
		if filter(v) {
			return OfSome[T](v)
		}
	}
	return OfNone[T]()
}

func (this *Optional[T]) Foreach(each func(T)) *Optional[T] {
	if this.some != nil {
		v := GetItem[T](this.some)
		each(v)
	}
	return this
}

// Exec Execute operation and map success to Some of Ok
func (this *Optional[T]) Exec(each func(T) *Optional[bool]) *Optional[T] {
	if this.some != nil {
		v := GetItem[T](this.some)
		r := each(v)
		if r.IsSome() {
			return OfOk[T]()
		}
		return Of[T](r.Val())
	}
	return this
}

func (this *Optional[T]) ListNonEmpty() bool {
	if this.IsSome() {
		if !IsSlice(this.some.Item) {
			panic("optional wrapped value is not a slice")
		}
		ss := reflect.ValueOf(this.some.Item)
		s := reflect.Indirect(ss)
		return  s.Len() > 0
	}
	return false
}

// ListMapToBool map to true id list len > 0 or else false
func (this *Optional[T]) ListMapToBool() *Optional[bool] {
	return Of[bool](!this.ListEmpty())
}

func (this *Optional[T]) ListEmpty() bool {
	if this.IsSome() {
		if !IsSlice(this.some.Item) {
			panic("optional wrapped value is not a slice")
		}
		ss := reflect.ValueOf(this.some.Item)
		s := reflect.Indirect(ss)
		return  s.Len() == 0
	}
	return false
}

func (this *Optional[T]) ListIfEmpty(cb func()) *Optional[T] {
	if this.IsSome() {
		if !IsSlice(this.some.Item) {
			panic("optional wrapped value is not a slice")
		}
		ss := reflect.ValueOf(this.some.Item)
		s := reflect.Indirect(ss)
		if s.Len() == 0 {
			cb()
		}
	}
	return this
}

func (this *Optional[T]) ListIfNonEmpty(cb func(T)) *Optional[T] {
	if this.IsSome() {
		if !IsSlice(this.some.Item) {
			panic("optional wrapped value is not a slice")
		}
		ss := reflect.ValueOf(this.some.Item)
		s := reflect.Indirect(ss)
		if s.Len() > 0 {
			cb(this.some.Item.(T))
		}
	}
	return this
}
// ListForeach Try to apply f to each list item if Some is a slice. If Some is not a list, throw panic
func (this *Optional[T]) ListForeach(f interface{}) *Optional[T] {

	if this.IsSome() {
		fnType := reflect.TypeOf(f)
		fnArgsCount := fnType.NumIn()

		if fnArgsCount != 1 {
			panic("map func should be one args ")
		}

		if !IsSlice(this.some.Item) {
			panic("optional wrapped value is not a slice")
		}

		fnValue := reflect.ValueOf(f)

		lst.Foreach(this.some.Item, func(i interface{}) {
			fnValue.Call([]reflect.Value{reflect.ValueOf(i)})
		})

	}
	return this
}

// ListFilter Try to apply list filter if Some is a slice. If Some is not a list, throw panic
func (this *Optional[T]) ListFilter(f interface{}) *Optional[T] {

	if this.IsSome() {
		fnType := reflect.TypeOf(f)
		fnArgsCount := fnType.NumIn()

		if fnArgsCount != 1 {
			panic("map func should be one args ")
		}

		if !IsSlice(this.some.Item) {
			panic("optional wrapped value is not a slice")
		}

		fnValue := reflect.ValueOf(f)

		items := lst.Filter(this.some.Item, func(i interface{}) bool {
			ret := fnValue.Call([]reflect.Value{reflect.ValueOf(i)})

			if len(ret) != 1 {
				panic("filter func should be one result")
			}

			if ret[0].Type().Kind() != reflect.Bool {
				panic("filter func should be bool")
			}

			return ret[0].Bool()
		})

		if len(items) == 0 {
			return OfNone[T]()
		}

		return Of[T](items)
	}

	return this
}

// ListMap Try to apply list filter if Some is a slice. If Some is not a list, throw panic
func (this *Optional[T]) ListMap(f interface{}) interface{} {

	if this.IsSome() {
		fnType := reflect.TypeOf(f)
		fnArgsCount := fnType.NumIn()

		if fnArgsCount != 1 {
			panic("map func should be one args ")
		}

		if !IsSlice(this.some.Item) {
			panic("optional wrapped value is not a slice")
		}

		fnValue := reflect.ValueOf(f)

		items := lst.Map(this.some.Item, func(i interface{}) interface{} {
			ret := fnValue.Call([]reflect.Value{reflect.ValueOf(i)})

			if len(ret) != 1 {
				panic("filter func should be one result")
			}

			return  ret
		})

		return NewSome(items)
	}

	if this.IsFail() {
		return  this.GetFail()
	}

	return NewNone()
}

// Map map Some to another thing
func (this *Optional[T]) Map(fn func(T) interface{}) interface{} {
	if this.IsSome() {
		v := GetItem[T](this.some)
		r := fn(v)

		switch r.(type) {
		case *Some, *None, *Fail:
			return r
		default:
			return Make(r)
		}
	}
	return NewNone()
}

// MapTo map Some to same thing
func (this *Optional[T]) MapTo(fn func(T) *Optional[T]) *Optional[T] {
	if this.IsSome() {
		v := GetItem[T](this.some)
		return fn(v)
	}

	return this
}

// MapToBool map Some to true
func (this *Optional[T]) MapToBool() *Optional[bool] {
	if this.IsSome() {
		return Of[bool](true)
	}
	return Of[bool](this.Val())
}

func (this *Optional[T]) MapToOk() *Optional[T] {
	if this.IsSome() {
		return OfOk[T]()
	}
	return this
}

func (this *Optional[T]) MapBool(f func(T) bool) *Optional[bool] {
	if this.IsSome() {
		return Of[bool](f(this.some.Item.(T)))
	}
	return Of[bool](this.Val())
}

func (this *Optional[T]) MapBoolOpt(f func(T) *Optional[bool]) *Optional[bool] {
	if this.IsSome() {
		return f(this.some.Item.(T))
	}
	return Of[bool](this.Val())
}

func (this *Optional[T]) MapToNone() interface{} {
	if this.IsSome() {
		return NewNone()
	}
	return this.Val()
}

func (this *Optional[T]) MapToSome(v interface{}) interface{} {
	if this.IsSome() {
		return NewSome(v)
	}
	return  this.Val()
}

func (this *Optional[T]) MapOpt(fn func(T) interface{}) *Optional[T] {

	if this.IsFail() {
		return this
	}

	if this.IsSome() {
		r := fn(this.some.Item.(T))
		return Of[T](r)
	}
	return OfNone[T]()
}

func (this *Optional[T]) OrElseOpt(v interface{}) *Optional[T] {

	if this.IsFail() {
		return this
	}

	if this.IsSome() {
		return Of[T](v)
	}

	return this
}

func (this *Optional[T]) If(cbSome func(T), cbNone func(), cbError func(err error)) *Optional[T] {
	this.IfFail(cbError)
	this.IfNonEmpty(cbSome)
	this.IfNone(cbNone)
	return this
}

func (this *Optional[T]) ValTo(f func(interface{})) *Optional[T] {
	return this.UnwrapTo(f)
}

func (this *Optional[T]) UnwrapTo(f func(interface{})) *Optional[T] {
	f(this.Val())
	return this
}

/*
func (this *Optional[T]) Else(cb func()) *Optional[T] {
	cb()
	return this
}*/

func OptionalMap[F any, T any](opt *Optional[F], fn func(F) *Optional[T], orElse ...func() *Optional[T]) *Optional[T] {
	var x T
	if opt.Any() {
		return fn(opt.Get())
	}
	if len(orElse) > 0 {
		return orElse[0]()
	}
	return New[T](x)
}

func MapEach[F any, T any](opt *Optional[[]F], fn func(F) T) *Optional[[]T] {
	items := []T{}
	if opt.Any() {

		for _, it := range opt.Get() {
			items = append(items, fn(it))
		}
	}
	return New[[]T](items)
}

func Map[F any, T any](opt *Optional[F], fn func(F) *Optional[T]) *Optional[T] {
	if opt.Any() {
		return fn(opt.Get())
	}
	return Of[T](opt.Val())
}

func MapMerge[T1 any, T2 any, R any](opt1 *Optional[T1], opt2 *Optional[T2], fn func(T1, T2) R) *Optional[R] {

	if opt1.IsSome() {
		if opt2.IsSome() {
			return Of[R](fn(opt1.Get(), opt2.Get()))
		}
		return Of[R](opt2.Val())
	}

	return Of[R](opt1.Val())
}


type None struct {
}

func NewNone() *None {
	return &None{}
}

type Some struct {
	Item interface{}
}

func NewSome(item interface{}) *Some {
	return &Some{Item: item}
}

// represents ignore result
type Ok struct{}

// represents ignore result
func SomeOk() *Some {
	return NewSome(&Ok{})
}

func IsSomeOk(v *Some) bool {
	switch v.Item.(type) {
	case *Ok:
		return true
	}
	return false
}
func IsOk(v interface{}) bool {
	switch v.(type) {
	case *Some:
		return IsSomeOk(v.(*Some))
	}
	return false
}

type Fail struct {
	Error error
	Item  interface{}
}

func (this *Fail) ErrorString() string {
	return this.Error.Error()
}

func NewFail(err error) *Fail {
	return &Fail{Error: err}
}

func NewFailWithItem(err error, item interface{}) *Fail {
	return &Fail{Error: err, Item: item}
}

func NewFailStr(format string, v ...interface{}) *Fail {
	return &Fail{Error: errors.New(fmt.Sprintf(format, v...))}
}

func FailIf(val bool, msg string, args ...interface{}) interface{} {
	if val {
		return NewFailStr(msg, args...)
	}
	return NewNone()
}

func FailIfOrElseDefault(val bool, msg string, def interface{}) interface{} {
	if val {
		return NewFailStr(msg)
	}
	return def
}

func FailIfFn(f func() bool, msg string, args ...interface{}) interface{} {
	return FailIf(f(), msg, args...)
}

func Maybe(val interface{}) interface{} {
	return Make(val)
}

func Make(val interface{}) interface{} {
	return MakeTry(val, nil)
}

func MakeTry(val interface{}, err error) interface{} {

	if err != nil {
		return NewFail(err)
	}

	if val == nil || IsNilFixed(val) {
		return NewNone()
	}

	switch val.(type) {
	case error:
		return NewFail(val.(error))
	case bool:
		//if val.(bool) {
		return NewSome(val)
		//}
		//return NewNone()
	case string:
		if val.(string) != "" {
			return NewSome(val)
		}
		return NewNone()
	case int:
		if val.(int) != 0 {
			return NewSome(val)
		}
		return NewNone()
	case int64:
		if val.(int64) != 0 {
			return NewSome(val)
		}
		return NewNone()
	case float32:
		if val.(float32) != 0 {
			return NewSome(val)
		}
		return NewNone()
	case float64:
		if val.(float64) != 0 {
			return NewSome(val)
		}
		return NewNone()
	case time.Time:
		if val.(time.Time).IsZero() {
			return NewNone()
		}
		return NewSome(val)
	default:
		return NewSome(val)
	}
}

func Get[R any](val interface{}) R {
	switch val.(type) {
	case *Some:
		return tryCastSome[R](val.(*Some))
	default:
		return val.(R)
	}
}

func GetPtr[R any](val interface{}) *R {
	return Get[*R](val)
}

func GetOrElse[R any](val interface{}, r R) R {
	if !IsNilFixed(val) {
		switch val.(type) {
		case *Some:
			return tryCastSome[R](val.(*Some))
		default:
			return val.(R)
		}
	}
	return r
}

func GetItem[R any](val interface{}) R {

	switch val.(type) {
	case *Some:
		return tryCastSome[R](GetSome(val))
	}
	var x R
	return x
}

func tryCastSome[T any](some *Some) T {
	item := some.Item
	if v, ok := item.(T); ok {
		return  v
	}
	var x T
	return x
}

func GetFail(val interface{}) *Fail {
	return val.(*Fail)
}

func GetSome(val interface{}) *Some {
	return val.(*Some)
}

func GetFailError(val interface{}) error {
	return val.(*Fail).Error
}

func OrElse[T any](e interface{}, v T) T {
	switch e.(type) {
	case *Some:
		return tryCastSome[T](e.(*Some))
	case *Optional[T]:
		t := e.(*Optional[T])
		if t.Any() {
			return t.Get()
		}
	}
	return v
}

func OrElseSome(e interface{}, v interface{}) *Some {
	switch e.(type) {
	case *Some:
		return e.(*Some)
	}
	return NewSome(v)
}

func IfNonEmpty[T any](e interface{}, cb func(T)) bool {
	switch e.(type) {
	case Some:
		cb(e.(T))
		return true
	default:
		return false
	}
}

func IfEmpty[T any](e interface{}, cb func()) bool {
	switch e.(type) {
	case None:
		cb()
		return true
	default:
		return false
	}
}

func IfEmptyOrElse[T any](e interface{}, emptyCb func(), elseCb func(T)) {
	if !IfEmpty[T](e, emptyCb) {
		elseCb(GetItem[T](e))
	}
}

func IfFail[T any](e interface{}, cb func(error)) bool {
	switch e.(type) {
	case error:
		cb(e.(error))
		return true
	default:
		return false
	}
}

/*
func MakeSlice(val interface{}, err error) interface{} {

	if err != nil {
		return NewFail(err)
	}

	if val == nil || IsNilFixed(val) {
		return NewNone()
	}

	ss := reflect.ValueOf(val)
	s := reflect.Indirect(ss)
	if s.Len() > 0 {
		return NewSome(val)
	}
	return NewNone()
}*/

func IsSlice(v interface{}) bool {
	return reflect.TypeOf(v).Kind() == reflect.Slice || reflect.TypeOf(v).Kind() == reflect.Array
}

func IsNilFixed(i interface{}) bool {
	if i == nil {
		return true
	}
	switch reflect.TypeOf(i).Kind() {
	case reflect.Ptr, reflect.Map, reflect.Array, reflect.Chan, reflect.Slice:
		return reflect.ValueOf(i).IsNil()
	}
	return false
}

func IsSimpleType(v interface{}) bool {
	switch v.(type) {
	case int, int64, float32, float64, bool, string:
		return true
	}
	return false
}

func FlatMap[T any, R any](vs *Optional[[]T], fn func(T) R) []R {
	var r []R

	if vs.NonEmpty() {
		return lists.Map[T, R](vs.Get(), fn)
	}
	return r
}

func Flatten[T any](vs *Optional[T], orElse ...T) T {
	var r T

	if len(orElse) > 0 {
		r = orElse[0]
	}

	if vs.NonEmpty() {
		return vs.Get()
	}

	return r
}

func Foreach[T any](optList *Optional[[]T], f func(T)) *Optional[[]T] {
	if optList.IsSome() {
		lst := optList.UnWrap()
		for _, it := range lst {
			f(it)
		}
	}
	return optList
}

func Filter[T any](optList *Optional[[]T], f func(T) bool) *Optional[[]T] {
	items := []T{}
	if optList.IsSome() {
		lst := optList.UnWrap()
		for _, it := range lst {
			if f(it) {
				items = append(items, it)
			}
		}
		return Of[[]T](items)
	}
	return optList
}

func UnwrapAll[T1 any, T2 any](opt1 *Optional[T1], opt2 *Optional[T2], fn func(T1, T2)) *Optional[bool] {
	if opt1.IsSome() {
		if opt2.IsSome() {
			fn(opt1.Get(), opt2.Get())
			return OfSome[bool](true)
		}
		return Of[bool](opt2.Val())
	}

	return Of[bool](opt1.Val())
}

func UnwrapAll3[T1 any, T2 any, T3 any](
	opt1 *Optional[T1],
	opt2 *Optional[T2],
	opt3 *Optional[T3],
	fn func(T1, T2, T3)) *Optional[bool] {
	if opt1.IsSome() {
		if opt2.IsSome() {
			if opt3.IsSome() {
				fn(opt1.Get(), opt2.Get(), opt3.Get())
				return Of[bool](true)
			}
			return Of[bool](opt3.Val())
		}
		return Of[bool](opt2.Val())
	}
	return Of[bool](opt1.Val())
}

// TryExtractValIfOptional Extract Optional.Val() from maybeOpt if it's a Optional.
func TryExtractValIfOptional(maybeOpt interface{}) (interface{}, bool) {
	typeOf := reflect.TypeOf(maybeOpt)
	valueOf := reflect.ValueOf(maybeOpt)
	if typeOf.Kind() == reflect.Ptr &&
		strings.Contains(typeOf.Elem().Name(), "Optional") {
		method := valueOf.MethodByName("Val")
		val := method.Call([]reflect.Value{})
		return val[0].Interface(), true
	}
	return maybeOpt, false
}
