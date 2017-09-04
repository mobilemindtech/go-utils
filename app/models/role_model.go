package models

import (
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/astaxie/beego/orm"
	"time"
)

type Role struct{
  
  Id int64 `form:"-" json:",string,omitempty"`
  CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
  UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`      
 
  Authority string `orm:"size(50)"`
  Description string `orm:"size(100)"`

  Session *db.Session `orm:"-"`
}

func NewRole(session *db.Session) *Role{
  return &Role{ Session: session }
}

func (this *Role) TableName() string{
  return "roles"
}

func (this *Role) IsPersisted() bool{
  return this.Id > 0
}

func (this *Role) List() (*[]*Role , error) { 
  var results []*Role
  err := this.Session.List(this, &results)
  return &results, err
}

func (this *Role) FindByAuthority(authority string) (role *Role, err error) {
  result := new(Role)

  query, err := this.Session.Query(this)

  if err != nil {
    return nil, err
  }  
  
  err = query.Filter("Authority", authority).One(result)
  
  if err == orm.ErrNoRows {
    return result, nil  
  }

  return result, err
}
