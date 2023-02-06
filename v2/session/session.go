package session

import (
	"errors"

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
			err = errors.New("invalid action")
		}

		if err != nil {
			return optional.WithFail[T](err)
		}
	}

	return optional.WithNone[T]()
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

func (this *RxSession[T]) Remove(entity T) *optional.Optional[T] {

	if err := this.session.Remove(entity); err != nil {
		return optional.WithFail[T](err)
	}
	return optional.WithSome[T](entity)
}

func (this *RxSession[T]) Persist(entity T) *optional.Optional[T] {

	if err := this.session.SaveOrUpdate(entity); err != nil {
		return optional.WithFail[T](err)
	}
	return optional.WithSome[T](entity)
}
