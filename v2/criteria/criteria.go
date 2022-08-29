package criteria

import (
	"github.com/mobilemindtec/go-utils/beego/db"
	_ "github.com/mobilemindtec/go-utils/v2/optional"
	"reflect"
	_ "fmt"
)

type DataCount[T any] struct {
	TotalCount int64
	Count32 int
	Count64 int64
	Results []*T
}

func NewDataCount[T any](totalCount int64, results []*T) *DataCount[T] {
	return &DataCount[T] { TotalCount: totalCount, Count64: totalCount, Count32: int(totalCount), Results: results }
}

type Criteria[T any] struct {
	db.Criteria

	sameFn func([]*T)
	sameOrNoneFn func([]*T)
	firstFn func(*T)
	firstOrNoneFn func(*T)
	failFn func(error)
	doneFn func()
	noneFn func()
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

func (this *Criteria[T]) Same(fn func([]*T)) *Criteria[T] {
	this.sameFn = fn
	return this
}

func (this *Criteria[T]) SameOrNone(fn func([]*T)) *Criteria[T] {
	this.sameOrNoneFn = fn
	return this
}

func (this *Criteria[T]) First(fn func(*T)) *Criteria[T] {
	this.firstFn = fn
	return this
}

func (this *Criteria[T]) FirstOrNone(fn func(*T)) *Criteria[T] {
	this.firstOrNoneFn = fn
	return this
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

func (this *Criteria[T]) Do() {
	
	if this.firstFn != nil {
		
		this.One()

		if this.Any && !this.HasError {
			r, _ := this.GetResult()
			this.firstFn(r)
		}

	} else if this.firstOrNoneFn != nil {
		
		this.One()

		if !this.HasError {
			r, _ := this.GetResult()
			this.firstOrNoneFn(r)
		}

	
	} else if this.sameFn != nil {

		this.List()

		if !this.HasError {
			if this.Any {
				rs, _ := this.GetResults()
				this.sameFn(rs)
			}
		}

	} else if this.sameOrNoneFn != nil {

		this.List()

		if !this.HasError {
			rs, _ := this.GetResults()
			this.sameOrNoneFn(rs)
		}

	}

	if this.HasError {
		if this.failFn != nil {
			this.failFn(this.Error)
		}
	}

	if this.noneFn  != nil {
		if this.Empty {
			this.noneFn()
		}
	}

	if this.doneFn != nil {
		this.doneFn()
	}
}

func (this *Criteria[T]) List() ([]*T, error) {
	this.Criteria.List()
	return this.GetResults()
}

func (this *Criteria[T]) ListAndCount() (int64, []*T, error) {
	this.Criteria.ListAndCount()
	r, err := this.GetResults()
	return this.Count64, r, err
}

func (this *Criteria[T]) One() (*T, error) {
	this.Criteria.One()
	return this.GetResult()
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

func (this *Criteria[T]) OrderAsc(path string) *Criteria[T]{
	this.Criteria.OrderAsc(path)
	return this
}

func (this *Criteria[T]) OrderDesc(path string) *Criteria[T]{
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
