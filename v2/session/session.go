package session

import (
	_ "errors"

	"fmt"
	"reflect"

	"github.com/mobilemindtec/go-utils/app/models"
	"github.com/mobilemindtec/go-utils/beego/db"
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type Action struct {
	entity interface{}
}

func (this *Action) Get() interface{} {
	return this.entity
}

type ActionSave struct {
	Action
}

func NewActionSave(e interface{}) *ActionSave {
	return &ActionSave{Action{entity: e}}
}

type ActionUpdate struct {
	Action
}

func NewActionUpdate(e interface{}) *ActionUpdate {
	return &ActionUpdate{Action{entity: e}}
}

type ActionPersist struct {
	Action
}

func NewActionPersist(e interface{}) *ActionPersist {
	return &ActionPersist{Action{entity: e}}
}

type ActionRemove struct {
	Action
}

func NewActionRemove(e interface{}) *ActionRemove {
	return &ActionRemove{Action{entity: e}}
}

type RxSession[T any] struct {
	session *db.Session
	actions []interface{}
}

func New[T any](session *db.Session) *RxSession[T] {
	return &RxSession[T]{session: session, actions: []interface{}{}}
}

func (this *RxSession[T]) AddAction(ac ...interface{}) *RxSession[T] {
	for _, it := range ac {
		this.actions = append(this.actions, it)
	}
	return this
}

func (this *RxSession[T]) RunWithTenantId(id int64, cb func(*RxSession[T]) T) T {
	var result T
	this.session.RunWithTenant(models.NewTenantWithId(id), func() {
		result = cb(this)
	})
	return result
}

func (this *RxSession[T]) AddPersist(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionPersist(o))
	}
	return this
}

func (this *RxSession[T]) AddSave(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionSave(o))
	}
	return this
}

func (this *RxSession[T]) AddUpdate(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionUpdate(o))
	}
	return this
}

func (this *RxSession[T]) AddRemove(items ...interface{}) *RxSession[T] {
	for _, o := range items {
		this.AddAction(NewActionRemove(o))
	}
	return this
}

func (this *RxSession[T]) Run() *optional.Optional[T] {
	for _, ac := range this.actions {

		var err error

		switch ac.(type) {
		case *ActionSave:
			err = this.session.Save(ac.(*ActionSave).Get())
			break
		case *ActionUpdate:
			err = this.session.Update(ac.(*ActionUpdate).Get())
			break
		case *ActionPersist:
			err = this.session.SaveOrUpdate(ac.(*ActionPersist).Get())
			break
		case *ActionRemove:
			err = this.session.Remove(ac.(*ActionRemove).Get())
			break
		default:
			err = fmt.Errorf("invalid action: %v", reflect.TypeOf(ac))
		}

		if err != nil {
			return optional.WithFail[T](err)
		}
	}

	return optional.WithEmpty[T]()
}

func (this *RxSession[T]) Save(entity T) *optional.Optional[T] {

	if err := this.session.Save(entity); err != nil {
		return optional.WithFail[T](err)
	}
	return optional.WithSome[T](entity)
}

func (this *RxSession[T]) Update(entity T) *optional.Optional[T] {

	if err := this.session.Update(entity); err != nil {
		return optional.WithFail[T](err)
	}
	return optional.WithSome[T](entity)
}

func (this *RxSession[T]) Remove(entity T) *optional.Optional[bool] {

	if err := this.session.Remove(entity); err != nil {
		return optional.WithFail[bool](err)
	}
	return optional.WithSome[bool](true)
}

func (this *RxSession[T]) Persist(entity T) *optional.Optional[T] {

	if err := this.session.SaveOrUpdate(entity); err != nil {
		return optional.WithFail[T](err)
	}
	return optional.WithSome[T](entity)
}
