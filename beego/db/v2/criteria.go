package v2

import (
	"github.com/mobilemindtec/go-utils/beego/db"
	"reflect"
	"fmt"
)

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

func NewCriteria[T any](session *db.Session) *Criteria[T] {
	var entity T


	entities := []*T{}
	fmt.Println("----------- entity ", entity, &entity, entities)
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

func (this *Criteria[T]) Eq(path string, value interface{}) *Criteria[T] {
	this.Criteria.Eq(path, value)
	return this
}

func (this *Criteria[T]) OrderAsc(path string) *Criteria[T]{
	this.Criteria.OrderAsc(path)
	return this
}

func (this *Criteria[T]) OrderDesc(path string) *Criteria[T]{
	this.Criteria.OrderDesc(path)
	return this
}