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
		r = this.criteria.Results
	} else if this.criteria.IsCount() {
		r = this.criteria.Count64
	} else if this.criteria.IsListAndCount() {
		r = &Page{
			TotalCount: this.criteria.Count32,
			Data:       this.criteria.Results}
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

	someFn           func([]*T)
	someOrNoneFn     func([]*T)
	someOrNoneNextFn func([]*T) interface{}

	someNextFn func([]*T) interface{}

	firstFn           func(*T)
	firstNextFn       func(*T) interface{}
	firstOrNoneFn     func(*T)
	firstOrNoneNextFn func(*T) interface{}
	failFn            func(error)
	doneFn            func()
	noneFn            func()
	successFn         func(interface{})
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

func (this *Criteria[T]) Some(fn func([]*T)) *Criteria[T] {
	this.someFn = fn
	return this
}

func (this *Criteria[T]) SomeNext(fn func([]*T) interface{}) interface{} {
	this.someNextFn = fn
	return this.DoNext()
}

func (this *Criteria[T]) SomeOrNone(fn func([]*T)) *Criteria[T] {
	this.someOrNoneFn = fn
	return this
}

func (this *Criteria[T]) SomeOrNoneNext(fn func([]*T) interface{}) interface{} {
	this.someOrNoneNextFn = fn
	return this.DoNext()
}

func (this *Criteria[T]) First(fn func(*T)) *Criteria[T] {
	this.firstFn = fn
	return this
}

func (this *Criteria[T]) FirstNext(fn func(*T) interface{}) interface{} {
	this.firstNextFn = fn
	return this.DoNext()
}

func (this *Criteria[T]) FirstOrNone(fn func(*T)) *Criteria[T] {
	this.firstOrNoneFn = fn
	return this
}

func (this *Criteria[T]) FirstOrNoneNext(fn func(*T) interface{}) interface{} {
	this.firstOrNoneNextFn = fn
	return this.DoNext()
}

func (this *Criteria[T]) Fail(fn func(error)) *Criteria[T] {
	this.failFn = fn
	return this
}

func (this *Criteria[T]) None(fn func()) *Criteria[T] {
	this.noneFn = fn
	return this
}

func (this *Criteria[T]) Done(fn func()) *Criteria[T] {
	this.doneFn = fn
	return this
}

func (this *Criteria[T]) Success(fn func(interface{})) *Criteria[T] {
	this.successFn = fn
	return this
}

func (this *Criteria[T]) DoFailNext() interface{} {

	if this.HasError {
		return optional.NewFail(this.Error)
	}

	return nil
}

func (this *Criteria[T]) Do() *Criteria[T] {
	this.DoNext()
	return this
}

func (this *Criteria[T]) Optional() *optional.Optional[T] {
	var r interface{}

	if this.Criteria.HasError {
		r = this.Criteria.Error
	} else if this.Criteria.IsOne() && this.Criteria.Any {
		r = this.Criteria.Result
	} else if this.Criteria.IsList() {
		r = this.Criteria.Results
	} else if this.Criteria.IsCount() {
		r = this.Criteria.Count64
	} else if this.Criteria.IsExists() {
		r = this.Criteria.Any
	} else if this.Criteria.IsListAndCount() {
		rs, _ := this.GetResults()
		r = &Page0[*T]{
			TotalCount: this.Criteria.Count32,
			Data:       rs}
	}

	if this.Criteria.IsList() {
		return optional.New[T](optional.MakeSlice(r, this.Criteria.Error))
	}

	return optional.New[T](optional.Make(r, this.Criteria.Error))
}

func (this *Criteria[T]) DoNext() interface{} {

	var ret interface{}

	if this.firstFn != nil || this.firstNextFn != nil {

		this.One()

		if this.Any && !this.HasError {
			r, _ := this.GetResult()

			if this.firstFn != nil {
				this.firstFn(r)
			} else {
				ret := this.firstNextFn(r)
				if ret != nil {
					switch ret.(type) {
					case *optional.None:
						if this.noneFn != nil {
							this.noneFn()
						}
						break
					case *optional.Fail:
						this.SetError(ret.(*optional.Fail).Error)
						break
					case error:
						this.SetError(ret.(error))
						break
					}
				}
			}
		}

	} else if this.firstOrNoneFn != nil || this.firstOrNoneNextFn != nil {

		this.One()

		if !this.HasError {
			r, _ := this.GetResult()

			if this.firstOrNoneFn != nil {
				this.firstOrNoneFn(r)
			} else {
				ret := this.firstOrNoneNextFn(r)
				if ret != nil {
					switch ret.(type) {
					case *optional.None:
						if this.noneFn != nil {
							this.noneFn()
						}
						break
					case *optional.Fail:
						this.SetError(ret.(*optional.Fail).Error)
						break
					case error:
						this.SetError(ret.(error))
						break
					}
				}
			}
		}

	} else if this.someFn != nil || this.someNextFn != nil {

		this.List()

		if !this.HasError {
			if this.Any {
				rs, _ := this.GetResults()
				if this.someFn != nil {
					this.someFn(rs)
				} else {
					ret = this.someNextFn(rs)
					if ret != nil {
						switch ret.(type) {
						case *optional.None:
							if this.noneFn != nil {
								this.noneFn()
							}
							break
						case *optional.Fail:
							this.SetError(ret.(*optional.Fail).Error)
							break
						case error:
							this.SetError(ret.(error))
							break
						}
					}
				}
			}
		}

	} else if this.someOrNoneFn != nil || this.someOrNoneNextFn != nil {

		this.List()

		if !this.HasError {
			rs, _ := this.GetResults()

			if this.someOrNoneFn != nil {
				this.someOrNoneFn(rs)
			} else {
				ret = this.someOrNoneNextFn(rs)
				if ret != nil {
					switch ret.(type) {
					case *optional.None:
						if this.noneFn != nil {
							this.noneFn()
						}
						break
					case *optional.Fail:
						this.SetError(ret.(*optional.Fail).Error)
						break
					case error:
						this.SetError(ret.(error))
						break
					}
				}
			}
		}

	}

	if this.HasError {
		if this.failFn != nil {
			this.failFn(this.Error)
		} else if this.firstNextFn != nil ||
			this.someNextFn != nil ||
			this.firstOrNoneNextFn != nil ||
			this.someOrNoneNextFn != nil {
			return optional.NewFail(this.Error)
		}
	} else {
		if this.successFn != nil {
			this.successFn(ret)
		}
	}

	if this.noneFn != nil {
		if this.Empty {
			this.noneFn()
		} else if this.firstNextFn != nil || this.someNextFn != nil {
			return optional.NewNone()
		}
	}

	if this.doneFn != nil {
		this.doneFn()
	}

	return ret
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

	results := []*T{}
	ss := reflect.ValueOf(this.Results)
	s := reflect.Indirect(ss)

	for i := 0; i < s.Len(); i++ {
		it := s.Index(i)
		results = append(results, it.Interface().(*T))
	}

	return results, this.Error
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
