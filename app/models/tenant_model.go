package models

import (
  "github.com/mobilemindtec/go-utils/beego/db"
	"time"
)

type Tenant struct{

  Id int64 `form:"-" json:",string,omitempty"`
  CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
  UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`

  Enabled bool `orm:"" valid:"Required;" form:"" json:",string,omitempty"`
  Name string `orm:"size(100)"  valid:"Required;MaxSize(100)" form:""`
  Documento string `orm:"size(20)"  valid:"Required;MaxSize(14);MinSize(11)" form:""`

  Cidade *Cidade `orm:"rel(fk);on_delete(do_nothing)" valid:"RequiredRel" form:""`

  Session *db.Session `orm:"-"`
}

func (this *Tenant) TableName() string{
  return "tenants"
}

func NewTenant(session *db.Session) *Tenant{
  return &Tenant{ Session: session }
}

func (this *Tenant) IsPersisted() bool{
  return this.Id > 0
}

func (this *Tenant) List() (*[]*Tenant , error) {
  var results []*Tenant
  err := this.Session.List(this, &results)
  return &results, err
}

func (this *Tenant) Page(page *db.Page) (*[]*Tenant , error) {
  var results []*Tenant

  page.AddFilterDefault("Name").MakeDefaultSort()

  err := this.Session.Page(this, &results, page)
  return &results, err
}
