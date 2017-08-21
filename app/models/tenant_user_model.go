package models

import (
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/astaxie/beego/orm"
  "time"
)

type TenantUser struct{
    
  Id int64 `form:"-" json:",string,omitempty"`
  CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
  UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`      

  Enabled bool `orm:"" valid:"Required;" form:"" json:",string,omitempty"`
  Tenant *Tenant `orm:"rel(fk);on_delete(do_nothing)" valid:"Required" form:"" goutils:"no_set_tenant;no_filter_tenant"`
  User *User `orm:"rel(fk);on_delete(do_nothing)" valid:"Required" form:",select"`

  Session *db.Session `orm:"-"`
}


func (this *TenantUser) TableName() string{
  return "tenant_users"
}

func NewTenantUser(session *db.Session) *TenantUser{
  return &TenantUser{ Session: session }
}

func (this *TenantUser) IsPersisted() bool{
  return this.Id > 0
}

func (this *TenantUser) LoadRelated(entity *TenantUser) {
  this.Session.Db.LoadRelated(entity, "Tenant")
  this.Session.Db.LoadRelated(entity, "User")
}

func (this *TenantUser) ListByTenant(tenant *Tenant) (*[]*TenantUser , error) { 
  var results []*TenantUser

  query, err := this.Session.Query(this)


  if err != nil {
    return nil, err
  }

  if err := this.Session.ToList(query.Filter("Tenant", tenant), &results); err != nil {
    return nil, err
  }

  return &results, err
}

func (this *TenantUser) ListByUser(user *User) (*[]*TenantUser , error) { 
  var results []*TenantUser

  query, err := this.Session.Query(this)


  if err != nil {
    return nil, err
  }

  if err := this.Session.ToList(query.Filter("User", user), &results); err != nil {
    return nil, err
  }

  return &results, err
}

func (this *TenantUser) List() (*[]*TenantUser , error) { 
  var results []*TenantUser
  err := this.Session.List(this, &results)
  return &results, err
}

func (this *TenantUser) FindByUserAndTenant(user *User, tenant *Tenant) (*TenantUser, error) {
  result := new(TenantUser)

  query, err := this.Session.Query(this)

  if err != nil {
    return nil, err
  }

  err = query.Filter("User", user).Filter("Tenant", tenant).One(result)

  if err == orm.ErrNoRows {
    return nil, nil
  } else if err == orm.ErrMultiRows {
    return nil, err
  }

  return result, err
}

func (this *TenantUser) GetFirstTenant(user *User) (*Tenant , error) { 
  var results []*TenantUser

  query, err := this.Session.Query(this)


  if err != nil {
    return nil, err
  }

  query = query.Filter("User", user).Filter("Enabled", true).RelatedSel("Tenant")
  if err := this.Session.ToList(query, &results); err != nil {
    return nil, err
  }

  if len(results) > 0 {
    return results[0].Tenant, nil
  }

  return nil, nil
}