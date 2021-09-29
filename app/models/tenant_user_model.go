package models

import (
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/beego/beego/v2/client/orm"
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
  this.Session.GetDb().LoadRelated(entity, "Tenant")
  this.Session.GetDb().LoadRelated(entity, "User")
}

func (this *TenantUser) ListByTenant(tenant *Tenant) (*[]*TenantUser , error) { 
  var results []*TenantUser

  query, err := this.Session.Query(this)


  if err != nil {
    return nil, err
  }

  query = query.Filter("Tenant", tenant).RelatedSel("User")

  if err := this.Session.ToList(query, &results); err != nil {
    return nil, err
  }

  return &results, err
}

func (this *TenantUser) ListActivesByTenant(tenant *Tenant) (*[]*TenantUser , error) { 
  var results []*TenantUser

  query, err := this.Session.Query(this)


  if err != nil {
    return nil, err
  }

  query = query.Filter("Tenant", tenant).Filter("Enabled", true).RelatedSel("Tenant")

  if err := this.Session.ToList(query, &results); err != nil {
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

  query = query.Filter("User", user).RelatedSel("Tenant")

  if err := this.Session.ToList(query, &results); err != nil {
    return nil, err
  }

  return &results, err
}

func (this *TenantUser) ListActivesByUser(user *User) (*[]*TenantUser , error) { 
  var results []*TenantUser

  query, err := this.Session.Query(this)


  if err != nil {
    return nil, err
  }

  query = query.Filter("User", user).Filter("Enabled", true).RelatedSel("Tenant")

  if err := this.Session.ToList(query, &results); err != nil {
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

func (this *TenantUser) HasActiveTenant(user *User) (bool , error) { 

  tenant, err := this.GetFirstTenant(user)

  return tenant != nil && tenant.IsPersisted(), err
}

func (this *TenantUser) Create(user *User, tenant *Tenant) error { 

  entity, err := this.FindByUserAndTenant(user, tenant)

  if err != nil && err != orm.ErrNoRows {
    return err
  }

  if entity != nil && entity.IsPersisted() {
    return nil
  }

  entity = &TenantUser{ User: user, Tenant: tenant, Enabled: true }

  return this.Session.Save(entity)
}

func (this *TenantUser) Remove(user *User, tenant *Tenant) error { 

  entity, err := this.FindByUserAndTenant(user, tenant)

  if err != nil && err != orm.ErrNoRows {
    return err
  }

  if entity != nil && entity.IsPersisted() {
    return this.Session.Remove(entity)
  }

  return nil
}

func (this *TenantUser) RemoveAllByUser(user *User) error { 

  results, err := this.ListByUser(user)

  if err != nil && err != orm.ErrNoRows {
    return err
  }

  for _, it := range *results {
    if err := this.Session.Remove(it); err != nil {
      return err
    }
  }

  return nil
}

func (this *TenantUser) RemoveAllByTenant(tenant *Tenant) error { 

  results, err := this.ListByTenant(tenant)

  if err != nil && err != orm.ErrNoRows {
    return err
  }

  for _, it := range *results {
    if err := this.Session.Remove(it); err != nil {
      return err
    }
  }

  return nil
}


func (this *TenantUser) ToUsers(results []*TenantUser) []*User {
  users := []*User{}

  for _, it := range results{
    users = append(users, it.User)
  }

  return users
}


func (this *TenantUser) ToTenants(results []*TenantUser) []*Tenant {
  tenants := []*Tenant{}

  for _, it := range results{
    tenants = append(tenants, it.Tenant)
  }

  return tenants
}


func (this *TenantUser) ListUsersByTenant(tenant *Tenant) ([]*User , error) { 
  results, err := this.ListByTenant(tenant)

  if err != nil {
    return nil, err
  }

  return this.ToUsers(*results), nil
}

func (this *TenantUser) ListUsersActivesByTenant(tenant *Tenant) ([]*User , error) { 
  results, err := this.ListActivesByTenant(tenant)

  if err != nil {
    return nil, err
  }

  return this.ToUsers(*results), nil
}

func (this *TenantUser) ListTenantsByUser(user *User) ([]*Tenant , error) { 
  results, err := this.ListByUser(user)

  if err != nil {
    return nil, err
  }

  return this.ToTenants(*results), nil
}

func (this *TenantUser) ListTenantsActivesByUser(user *User) ([]*Tenant , error) { 
  results, err := this.ListActivesByUser(user)

  if err != nil {
    return nil, err
  }

  return this.ToTenants(*results), nil
}
