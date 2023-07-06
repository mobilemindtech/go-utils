package optional

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

type Optional[T any] struct {
	some  *Some
	none  *None
	fail  *Fail
	empty *Empty
}

func WithSome[T any](v interface{}) *Optional[T] {

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

func WithFail[T any](v interface{}) *Optional[T] {

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

func WithNone[T any]() *Optional[T] {
	return &Optional[T]{none: NewNone()}
}

func WithEmpty[T any]() *Optional[T] {
	return &Optional[T]{empty: NewEmpty()}
}

func NewE[T any](val interface{}, err interface{}) *Optional[T] {
	return TryMake[T](val, err)
}

func TryMake[T any](val interface{}, err interface{}) *Optional[T] {
	if !IsNilFixed(err) {
		return New[T](err)
	}
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
	case *Empty:
		opt.empty = val.(*Empty)
		break
	case error:
		opt.fail = NewFail(val.(error))
		break
	default:
		mkd := Make(val, nil)
		switch mkd.(type) {
		case *Some:
			opt.some = mkd.(*Some)
			break
		case *None:
			opt.none = mkd.(*None)
			break
		case *Empty:
			opt.empty = mkd.(*Empty)
			break
		case *Fail:
			opt.fail = mkd.(*Fail)
			break
		default:
			panic(fmt.Sprint("can't get type from: %v", mkd))
		}
		break
	}
	return &opt
}

func (this *Optional[T]) GetFail() *Fail {
	return this.fail
}

func (this *Optional[T]) GetSome() *Some {
	return this.some
}

func (this *Optional[T]) OrElse(v T) T {
	return GetOrElese[T](this.some, v)
}

func (this *Optional[T]) Get() T {
	return Get[T](this.some.Item)
}

func (this *Optional[T]) Any() bool {
	return this.some != nil
}

func (this *Optional[T]) Fail() bool {
	return this.fail != nil
}

func (this *Optional[T]) Empty() bool {
	return this.none != nil
}

func (this *Optional[T]) NonEmpty() bool {
	return this.none == nil && this.fail == nil
}

func (this *Optional[T]) Val() interface{} {
	if this.some != nil {
		return this.some
	} else if this.none != nil {
		return this.none
	} else if this.empty != nil {
		return this.empty
	} else if this.fail != nil {
		return this.fail
	} else {
		return NewEmpty()
	}
}

func (this *Optional[T]) IfFail(cb func(error)) *Optional[T] {
	if this.fail != nil {
		cb(this.fail.Error)
	}
	return this
}

func (this *Optional[T]) IfSome(cb func(T)) *Optional[T] {
	if this.some != nil {
		cb(GetItem[T](this.some))
	}
	return this
}

func (this *Optional[T]) IfNone(cb func()) *Optional[T] {
	if this.none != nil {
		cb()
	}
	return this
}

func (this *Optional[T]) IfEmpty(cb func()) *Optional[T] {
	if this.empty != nil || this.none != nil {
		cb()
	}
	return this
}

func (this *Optional[T]) IfNonEmpty(cb func(T)) *Optional[T] {
	if this.some != nil {
		cb(GetItem[T](this.some))
	}
	return this
}

func (this *Optional[T]) Filter(filter func(T) bool) *Optional[T] {
	if this.some != nil {
		v := GetItem[T](this.some)
		if filter(v) {
			return WithSome[T](v)
		}
	}
	return WithNone[T]()
}

func (this *Optional[T]) Foreach(each func(T)) *Optional[T] {
	if this.some != nil {
		v := GetItem[T](this.some)
		each(v)
	}
	return this
}

func (this *Optional[T]) Map(fn func(T) interface{}) interface{} {
	if this.some != nil {
		v := GetItem[T](this.some)
		r := fn(v)

		switch r.(type) {
		case *Some, *None, *Empty, *Fail:
			return r
		default:
			return Make0(r)
		}
	}
	return NewNone()
}

func (this *Optional[T]) MapToNone() interface{} {
	return NewNone()
}

func (this *Optional[T]) MapToEmpty() interface{} {
	return NewEmpty()
}

func (this *Optional[T]) MapToSome(v interface{}) interface{} {
	return NewSome(v)
}

func (this *Optional[T]) IfOrElse(cbSome func(T), cbNone func()) *Optional[T] {

	this.IfNonEmpty(cbSome)
	this.IfEmpty(cbNone)

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

func Map[F any, T any](opt *Optional[[]F], fn func(F) T) *Optional[[]T] {
	items := []T{}
	if opt.Any() {

		for _, it := range opt.Get() {
			items = append(items, fn(it))
		}
	}
	return New[[]T](items)
}

type Empty struct {
}

func NewEmpty() *Empty {
	return &Empty{}
}

func DoNext() *Empty {
	return NewEmpty()
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

func NewSomeEmpty() *Some {
	return &Some{}
}

type Try struct {
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

func NextOrFail(val interface{}) interface{} {
	return NewFailOrEmpty(val)
}

func NewFailOrEmpty(val interface{}) interface{} {

	if val == nil || IsNilFixed(val) {
		return NewEmpty()
	}

	switch val.(type) {
	case error:
		return NewFail(val.(error))
	case *Fail:
		return val.(*Fail)
	}

	return NewEmpty()
}

type Success struct {
	Item interface{}
}

func (this *Success) WithItem(item interface{}) *Success {
	this.Item = item
	return this
}

func NewSuccess() *Success {
	return &Success{}
}

type Either struct {
}

type Left struct {
	Item interface{}
}

func (this *Left) WithItem(item interface{}) *Left {
	this.Item = item
	return this
}

func NewLeft() *Left {
	return &Left{}
}

type Rigth struct {
	Item interface{}
}

func (this *Rigth) WithItem(item interface{}) *Rigth {
	this.Item = item
	return this
}

func NewRigth() *Rigth {
	return &Rigth{}
}

func Get[R any](val interface{}) R {
	switch val.(type) {
	case *Some:
		return val.(*Some).Item.(R)
	default:
		return val.(R)
	}
}

func GetPtr[R any](val interface{}) *R {
	return Get[*R](val)
}

func GetOrElese[R any](val interface{}, r R) R {
	if !IsNilFixed(val) {
		switch val.(type) {
		case *Some:
			return val.(*Some).Item.(R)
		default:
			return val.(R)
		}
	}
	return r
}

func GetItem[R any](val interface{}) R {

	switch val.(type) {
	case *Some:
		return GetSome(val).Item.(R)
	case *Success:
		return GetSuccess(val).Item.(R)
	case *Left:
		return GetLeft(val).Item.(R)
	case *Rigth:
		return GetRigth(val).Item.(R)
	default:
		var x R
		return x
	}
}

func GetFail(val interface{}) *Fail {
	return val.(*Fail)
}

func GetSuccess(val interface{}) *Success {
	return val.(*Success)
}

func GetSome(val interface{}) *Some {
	return val.(*Some)
}

func GetLeft(val interface{}) *Left {
	return val.(*Left)
}

func GetRigth(val interface{}) *Rigth {
	return val.(*Rigth)
}

func GetFailError(val interface{}) error {
	return val.(*Fail).Error
}

func OrElse[T any](e interface{}, v T) T {
	switch e.(type) {
	case *Some:
		return e.(*Some).Item.(T)
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

func Just[T any](e interface{}) *Optional[T] {
	return New[T](e)
}

func Maybe[T any](e interface{}) *Optional[T] {
	return New[T](e)
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
	case None, Empty:
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
}

func Make0(val interface{}) interface{} {
	return Make(val, nil)
}

func Make(val interface{}, err error) interface{} {

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
		if val.(bool) {
			return NewSome(val)
		}
		return NewNone()
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

func IsSimple(v interface{}) bool {
	switch v.(type) {
	case int, int64, float32, float64, bool, string:
		return true
	}
	return false
}
