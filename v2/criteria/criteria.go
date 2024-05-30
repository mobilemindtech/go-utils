package criteria

import (
	_ "fmt"
	"github.com/mobilemindtec/go-io/io"
	"github.com/mobilemindtec/go-io/option"
	"github.com/mobilemindtec/go-io/result"
	"github.com/mobilemindtec/go-io/types"
	"reflect"
	"strings"

	"github.com/mobilemindtec/go-utils/v2/lists"

	"github.com/mobilemindtec/go-utils/beego/db"
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type Page struct {
	TotalCount int         `json:"total_count" jsonp:""`
	Data       interface{} `json:"data" jsonp:""`
}

func (this *Page) Count() int64 {
	return int64(this.TotalCount)
}

type PageOf[T any] struct {
	TotalCount int `json:"total_count" jsonp:""`
	Data       []T `json:"data" jsonp:""`
}

func (this *PageOf[T]) Count() int64 {
	return int64(this.TotalCount)
}

func (this *PageOf[T]) ToPage() *Page {
	data := lists.Map[T, interface{}](this.Data, func(t T) interface{} { return t })
	return &Page{Data: data, TotalCount: this.TotalCount}
}

func MapPageOf[T any, R any](
	p1 *optional.Optional[*PageOf[T]], fn func(T) R) *optional.Optional[*PageOf[R]] {

	if p1.IsSome() {
		page := p1.Get()
		results := lists.Map[T, R](page.Data, fn)
		return optional.Of[*PageOf[R]](&PageOf[R]{Data: results, TotalCount: page.TotalCount})
	}

	return optional.Of[*PageOf[R]](p1.Val())
}

func GetPageData[T any](rs *Page) []T {
	return rs.Data.([]T)
}

func TryExtractPageIfPegeOf(maybePage interface{}) (interface{}, bool) {
	typeOf := reflect.TypeOf(maybePage)
	valueOf := reflect.ValueOf(maybePage)
	if typeOf.Kind() == reflect.Ptr &&
		strings.Contains(typeOf.Elem().Name(), "PageOf") {
		method := valueOf.MethodByName("ToPage")
		val := method.Call([]reflect.Value{})
		return val[0].Interface(), true
	}
	return maybePage, false
}

type Reactive struct {
	criteria *db.Criteria
}

func NewReactive(c *db.Criteria) *Reactive {
	return &Reactive{criteria: c}
}

func (this *Reactive) Get() interface{} {
	var r interface{}
	if this.criteria.IsOne() && this.criteria.Any {
		r = this.criteria.Result
	} else if this.criteria.IsList() {
		r = reflect.ValueOf(this.criteria.Results).Elem().Interface()
	} else if this.criteria.IsCount() {
		r = this.criteria.Count64
	} else if this.criteria.IsListAndCount() {
		r = &Page{
			TotalCount: this.criteria.Count32,
			Data:       reflect.ValueOf(this.criteria.Results).Elem().Interface()}
	} else if this.criteria.IsExists() {
		r = this.criteria.Any
	}

	return optional.MakeTry(r, this.criteria.Error)
}

func (this *Reactive) GetAsPage() *optional.Optional[*Page] {

	if this.criteria.HasError {
		return optional.OfFail[*Page](this.criteria.Error)
	}

	val := this.Get()

	switch val.(type) {
	case *optional.Some:
		return optional.Of[*Page](val.(*optional.Some).Item.(*Page))
	case *optional.Fail:
		return optional.OfFail[*Page](val)
	default:
		return optional.OfFail[*Page]("wrong type. expected *Page")
	}

}

func (this *Reactive) One() *Reactive {
	return this.First()
}

func (this *Reactive) First() *Reactive {
	this.criteria.One()
	return this
}

func (this *Reactive) All() *Reactive {
	return this.List()
}

func (this *Reactive) List() *Reactive {
	this.criteria.List()
	return this
}

func (this *Reactive) Any() *Reactive {
	this.criteria.Exists()
	return this
}

func (this *Reactive) Count() *Reactive {
	this.criteria.Count()
	return this
}

func (this *Reactive) Page() *Reactive {
	this.criteria.ListAndCount()
	return this
}

type Criteria[T any] struct {
	db.Criteria
}

func NewCond() *db.Criteria {
	return db.NewCondition()
}

func Read[T any]() *Criteria[T] {
	s := db.NewSession()
	if err := s.OpenNoTx(); err != nil {
		panic(err)
	}
	return New[T](s)
}

func New[T any](session *db.Session) *Criteria[T] {
	var entity T
	entities := []*T{}
	criteria := &Criteria[T]{}
	criteria.SetDefaults()
	criteria.Session = session
	criteria.Result = &entity
	criteria.Results = &entities
	return criteria
}

func (this *Criteria[T]) Id(id int64) *Criteria[T] {
	this.Criteria.Eq("Id", id)
	return this
}

func (this *Criteria[T]) Pk(id int) *Criteria[T] {
	return this.Id(int64(id))
}

func (this *Criteria[T]) Rx() *Reactive {
	return NewReactive(&this.Criteria)
}

/*
*

	return empty lias as optional.Empty
*/
func (this *Criteria[T]) OptAll() *optional.Optional[[]*T] {
	this.List()

	if this.Criteria.HasError {
		return optional.OfFail[[]*T](this.Error)
	}

	all := reflect.ValueOf(this.Criteria.Results).Elem().Interface().([]*T)

	/*
		if !this.Any {
			//logs.Debug("no data")
			return optional.OfOk[[]*T]()
		}*/

	return optional.OfSome[[]*T](all)
}

/*
*

	return empty lias as optional.Some with a empty list
*/
func (this *Criteria[T]) OptList() *optional.Optional[[]*T] {
	this.List()

	if this.Criteria.HasError {
		return optional.OfFail[[]*T](this.Error)
	}

	all := reflect.ValueOf(this.Criteria.Results).Elem().Interface().([]*T)

	return optional.OfSome[[]*T](all)
}

func (this *Criteria[T]) OptOne() *optional.Optional[*T] {
	return this.OptFirst()
}

func (this *Criteria[T]) OptFirst() *optional.Optional[*T] {
	this.One()

	if this.Criteria.HasError {
		return optional.OfFail[*T](this.Error)
	}

	if !this.Any {
		return optional.OfNone[*T]()
	}

	return optional.OfSome[*T](this.Result.(*T))
}

func (this *Criteria[T]) OptCount() *optional.Optional[int] {
	this.Count()

	if this.Criteria.HasError {
		return optional.OfFail[int](this.Error)
	}
	return optional.OfSome[int](this.Count32)
}

func (this *Criteria[T]) OptAny() *optional.Optional[bool] {
	this.Exists()

	if this.Criteria.HasError {
		return optional.OfFail[bool](this.Error)
	}
	return optional.OfSome[bool](this.Any)
}

func (this *Criteria[T]) OptPage() *optional.Optional[*PageOf[*T]] {
	this.ListAndCount()

	if this.Criteria.HasError {
		return optional.OfFail[*PageOf[*T]](this.Error)
	}

	all := reflect.ValueOf(this.Criteria.Results).Elem().Interface().([]*T)
	return optional.OfSome[*PageOf[*T]](&PageOf[*T]{
		TotalCount: this.Criteria.Count32,
		Data:       all})
}

func (this *Criteria[T]) OptDelete() *optional.Optional[int] {
	this.Delete()

	if this.Criteria.HasError {
		return optional.OfFail[int](this.Error)
	}
	return optional.OfSome[int](this.Count32)
}

func (this *Criteria[T]) GetFirstIO() *types.IO[*T] {
	return io.IO[*T](
		io.AttemptOfResultOption(func() *result.Result[*option.Option[*T]] {
			return this.GetFirst()
		}))
}

func (this *Criteria[T]) GetFirst() *result.Result[*option.Option[*T]] {
	this.One()

	if this.Criteria.HasError {
		return result.OfError[*option.Option[*T]](this.Error)
	}
	if !this.Any {
		return result.OfValue(option.None[*T]())
	}
	return result.OfValue(option.Some(this.Result.(*T)))
}

func (this *Criteria[T]) GetAllIO() *types.IO[*[]T] {
	return io.IO[*[]T](
		io.Attempt(func() *result.Result[[]*T] {
			return this.GetAll()
		}))
}

func (this *Criteria[T]) GetAll() *result.Result[[]*T] {
	this.List()
	if this.Criteria.HasError {
		return result.OfError[[]*T](this.Error)
	}
	all := reflect.ValueOf(this.Criteria.Results).Elem().Interface().([]*T)
	return result.OfValue(all)
}

func (this *Criteria[T]) GetCountIO() *types.IO[int] {
	return io.IO[int](
		io.Attempt(func() *result.Result[int] {
			return this.GetCount()
		}))
}

func (this *Criteria[T]) GetCount() *result.Result[int] {
	this.Count()
	if this.Criteria.HasError {
		return result.OfError[int](this.Error)
	}
	return result.OfValue(this.Count32)
}

func (this *Criteria[T]) GetAnyIO() *types.IO[bool] {
	return io.IO[bool](
		io.Attempt(func() *result.Result[bool] {
			return this.GetAny()
		}))
}

func (this *Criteria[T]) GetAny() *result.Result[bool] {
	this.Exists()

	if this.Criteria.HasError {
		return result.OfError[bool](this.Error)
	}
	return result.OfValue(this.Any)
}

func (this *Criteria[T]) GetPageIO() *types.IO[*PageOf[*T]] {
	return io.IO[*PageOf[*T]](
		io.Attempt(func() *result.Result[*PageOf[*T]] {
			return this.GetPage()
		}))
}
func (this *Criteria[T]) GetPage() *result.Result[*PageOf[*T]] {
	this.ListAndCount()

	if this.Criteria.HasError {
		return result.OfError[*PageOf[*T]](this.Error)
	}

	all := reflect.ValueOf(this.Criteria.Results).Elem().Interface().([]*T)
	page := &PageOf[*T]{this.Criteria.Count32, all}
	return result.OfValue(page)

}

func (this *Criteria[T]) Eager(related ...string) *Criteria[T] {
	this.Criteria.SetRelatedsSel(related...)
	return this
}

func (this *Criteria[T]) SetRelatedSel(related ...string) *Criteria[T] {
	this.Criteria.SetRelatedsSel(related...)
	return this
}

func (this *Criteria[T]) All() ([]*T, error) {
	return this.List()
}

func (this *Criteria[T]) Each(each func(*T)) error {

	all, err := this.List()

	if err != nil {
		return err
	}

	for _, it := range all {
		each(it)
	}

	return nil
}

func (this *Criteria[T]) List() ([]*T, error) {
	this.Criteria.List()
	return this.GetResults()
}

func (this *Criteria[T]) Exists() (bool, error) {
	this.Criteria.One()
	return this.Any, this.Error
}

func (this *Criteria[T]) Page() (*PageOf[*T], error) {
	this.Criteria.ListAndCount()
	r, err := this.GetResults()
	return &PageOf[*T]{Data: r, TotalCount: this.Count32}, err
}

func (this *Criteria[T]) GetResult() (*T, error) {
	if this.Any {
		c, _ := this.Result.(*T)
		return c, this.Error
	}
	return nil, this.Error
}

func (this *Criteria[T]) GetResults() ([]*T, error) {
	all := reflect.ValueOf(this.Criteria.Results).Elem().Interface().([]*T)
	return all, this.Error
}

func (this *Criteria[T]) FindById(id int64) (*T, error) {
	var entity T
	r, err := this.Session.FindById(&entity, id)
	if err != nil {
		return nil, err
	}

	m, _ := r.(db.Model)

	if m.IsPersisted() {
		return &entity, nil
	}

	return nil, nil
}

func (this *Criteria[T]) Get(id int64) *optional.Optional[*T] {
	var entity T
	r, err := this.Session.FindById(&entity, id)
	if err != nil {
		return optional.OfFail[*T](err)
	}

	m, _ := r.(db.Model)

	if m.IsPersisted() {
		return optional.Of[*T](&entity)
	}

	return optional.OfNone[*T]()
}

func (this *Criteria[T]) OrderAsc(path string) *Criteria[T] {
	this.Criteria.OrderAsc(path)
	return this
}

func (this *Criteria[T]) OrderDesc(path string) *Criteria[T] {
	this.Criteria.OrderDesc(path)
	return this
}

func (this *Criteria[T]) Eq(path string, value interface{}) *Criteria[T] {
	this.Criteria.Eq(path, value)
	return this
}

func (this *Criteria[T]) If(test bool, c func(*Criteria[T])) *Criteria[T] {
	if test {
		c(this)
	}
	return this
}

func (this *Criteria[T]) IfTest(test func() bool, c func(*Criteria[T])) *Criteria[T] {
	if test() {
		c(this)
	}
	return this
}

func (this *Criteria[T]) Ne(path string, value interface{}) *Criteria[T] {
	this.Criteria.Ne(path, value)
	return this
}

func (this *Criteria[T]) Le(path string, value interface{}) *Criteria[T] {
	this.Criteria.Le(path, value)
	return this
}

func (this *Criteria[T]) Lt(path string, value interface{}) *Criteria[T] {
	this.Criteria.Lt(path, value)
	return this
}

func (this *Criteria[T]) Ge(path string, value interface{}) *Criteria[T] {
	this.Criteria.Ge(path, value)
	return this
}

func (this *Criteria[T]) Gt(path string, value interface{}) *Criteria[T] {
	this.Criteria.Gt(path, value)
	return this
}

func (this *Criteria[T]) Like(path string, value interface{}) *Criteria[T] {
	this.Criteria.Like(path, value)
	return this
}

func (this *Criteria[T]) NotLike(path string, value interface{}) *Criteria[T] {
	this.Criteria.NotLike(path, value)
	return this
}

func (this *Criteria[T]) LikeMatch(path string, value interface{}, likeMatch db.CriteriaLikeMatch) *Criteria[T] {
	this.Criteria.LikeMatch(path, value, likeMatch)
	return this
}

func (this *Criteria[T]) LikeAnyware(path string, value interface{}) *Criteria[T] {
	this.Criteria.LikeMatch(path, value, db.Anywhare)
	return this
}

func (this *Criteria[T]) LikeIAnyware(path string, value interface{}) *Criteria[T] {
	this.Criteria.LikeMatch(path, value, db.IAnywhare)
	return this
}

func (this *Criteria[T]) LikeStarts(path string, value interface{}) *Criteria[T] {
	this.Criteria.LikeMatch(path, value, db.StartsWith)
	return this
}

func (this *Criteria[T]) LikeIStarts(path string, value interface{}) *Criteria[T] {
	this.Criteria.LikeMatch(path, value, db.IStartsWith)
	return this
}

func (this *Criteria[T]) LikeEnds(path string, value interface{}) *Criteria[T] {
	this.Criteria.LikeMatch(path, value, db.EndsWith)
	return this
}

func (this *Criteria[T]) LikeIEnds(path string, value interface{}) *Criteria[T] {
	this.Criteria.LikeMatch(path, value, db.IEndsWith)
	return this
}

func (this *Criteria[T]) NotLikeMatch(path string, value interface{}, likeMatch db.CriteriaLikeMatch) *Criteria[T] {
	this.Criteria.NotLikeMatch(path, value, likeMatch)
	return this
}

func (this *Criteria[T]) Between(path string, value interface{}, value2 interface{}) *Criteria[T] {
	this.Criteria.Between(path, value, value2)
	return this
}

func (this *Criteria[T]) IsNull(path string) *Criteria[T] {
	this.Criteria.IsNull(path)
	return this
}

func (this *Criteria[T]) IsNotNull(path string) *Criteria[T] {
	this.Criteria.IsNotNull(path)
	return this
}

func (this *Criteria[T]) In(path string, values ...interface{}) *Criteria[T] {
	this.Criteria.In(path, values)
	return this
}

func (this *Criteria[T]) NotIn(path string, values ...interface{}) *Criteria[T] {
	this.Criteria.In(path, values)
	return this
}

func (this *Criteria[T]) Or(criteria *db.Criteria) *Criteria[T] {
	this.Criteria.Or(criteria)
	return this
}

func (this *Criteria[T]) AndOr(criteria *db.Criteria) *Criteria[T] {
	this.Criteria.AndOr(criteria)
	return this
}

func (this *Criteria[T]) OrAnd(criteria *db.Criteria) *Criteria[T] {
	this.Criteria.OrAnd(criteria)
	return this
}

func (this *Criteria[T]) AndOrAnd(criteria *db.CriteriaSet) *Criteria[T] {
	this.Criteria.AndOrAnd(criteria)
	return this
}

func (this *Criteria[T]) SetPage(page *db.Page) *Criteria[T] {
	this.Criteria.SetPage(page)
	return this
}

func (this *Criteria[T]) SetLimit(limit int64) *Criteria[T] {
	this.Criteria.SetLimit(limit)
	return this
}

func (this *Criteria[T]) SetOffset(offset int64) *Criteria[T] {
	this.Criteria.SetOffset(offset)
	return this
}
