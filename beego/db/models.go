package db

import (
	"github.com/beego/beego/v2/client/orm"
)

type Model interface {
  IsPersisted() bool
  TableName() string
}


// after hooks persis
type ModelHookAfterSave interface {
  AfterSave() error
}

type ModelHookAfterUpdate interface {
  AfterUpdate() error
}

type ModelHookAfterRemove interface {
  AfterRemove() error
}

// before hooks persist
type ModelHookBeforeSave interface {
  BeforeSave() error
}

type ModelHookBeforeUpdate interface {
  BeforeUpdate() error
}

type ModelHookBeforeRemove interface {
  BeforeRemove() error
}

// before load
type ModelHookBeforeCriteria interface {
  BeforeCriteria(criteria *Criteria)
}

type ModelHookBeforeQuery interface {
  BeforeQuery(querySeter orm.QuerySeter) orm.QuerySeter
}


// after load
type ModelHookAfterLoad interface {
  AfterLoad(entity interface{}) (next bool, err error)
}

type ModelHookAfterList interface {
  AfterList(entities interface{})
}

type TenantModel interface {
  GetId() int64
}