package criteria

import (
	_ "fmt"
	"reflect"

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

type Page0[T any] struct {
	TotalCount int `json:"total_count" jsonp:""`
	Data       []T `json:"data" jsonp:""`
}

func (this *Page0[T]) Count() int64 {
	return int64(this.TotalCount)
}

func GetPageData[T any](rs *Page) []T {
	return rs.Data.([]T)
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

	return optional.Make(r, this.criteria.Error)
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

func New[T any](session *db.Session) *Criteria[T] {
	var entity T
	entities := []*T{}
	criteria := &Criteria[T]{}
	criteria.Defaults()
	criteria.Session = session
	criteria.Result = &entity
	criteria.Results = &entities
	return criteria
}

func (this *Criteria[T]) Rx() *Reactive {
	return NewReactive(&this.Criteria)
}

func (this *Criteria[T]) OptAll() *optional.Optional[[]*T] {
	this.List()

	if this.Criteria.HasError {
		return optional.WithFail[[]*T](this.Error)
	}

	all := reflect.ValueOf(this.Criteria.Results).Elem().Interface().([]*T)

	if !this.Any {
		return optional.WithEmpty[[]*T]()
	}

	return optional.WithSome[[]*T](all)
}

func (this *Criteria[T]) OptOne() *optional.Optional[*T] {
	return this.OptFirst()
}

func (this *Criteria[T]) OptFirst() *optional.Optional[*T] {
	this.One()

	if this.Criteria.HasError {
		return optional.WithFail[*T](this.Error)
	}

	if !this.Any {
		return optional.WithNone[*T]()
	}

	return optional.WithSome[*T](this.Result.(*T))
}

func (this *Criteria[T]) OptCount() *optional.Optional[int] {
	this.Count()

	if this.Criteria.HasError {
		return optional.WithFail[int](this.Error)
	}
	return optional.WithSome[int](this.Count32)
}

func (this *Criteria[T]) OptAny() *optional.Optional[bool] {
	this.Exists()

	if this.Criteria.HasError {
		return optional.WithFail[bool](this.Error)
	}
	return optional.WithSome[bool](true)
}

func (this *Criteria[T]) OptPage() *optional.Optional[*Page0[*T]] {
	this.ListAndCount()

	if this.Criteria.HasError {
		return optional.WithFail[*Page0[*T]](this.Error)
	}

	all := reflect.ValueOf(this.Criteria.Results).Elem().Interface().([]*T)
	return optional.WithSome[*Page0[*T]](&Page0[*T]{
		TotalCount: this.Criteria.Count32,
		Data:       all})
}

func (this *Criteria[T]) SetRelatedSel(related ...string) *Criteria[T] {
	this.Criteria.SetRelatedsSel(related...)
	return this
}

func (this *Criteria[T]) All() ([]*T, error) {
	return this.List()

}
func (this *Criteria[T]) List() ([]*T, error) {
	this.Criteria.List()
	return this.GetResults()
}

func (this *Criteria[T]) Exists() (bool, error) {
	this.Criteria.One()
	return this.Any, this.Error
}

func (this *Criteria[T]) Page() (*Page0[*T], error) {
	this.Criteria.ListAndCount()
	r, err := this.GetResults()
	return &Page0[*T]{Data: r, TotalCount: this.Count32}, err
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

func (this *Criteria[T]) Get(id int64) (*T, bool, error) {
	var entity T
	r, err := this.Session.FindById(&entity, id)
	if err != nil {
		return nil, false, err
	}

	m, _ := r.(db.Model)

	if m.IsPersisted() {
		return &entity, true, nil
	}

	return nil, false, nil
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
