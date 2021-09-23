package models

import (
	"github.com/mobilemindtec/go-utils/beego/db"
  "github.com/beego/beego/v2/client/orm"
  "errors"
  "time"
  "fmt"
)

type UserRole struct{

  Id int64 `form:"-" json:",string,omitempty"`
  CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
  UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`

  User *User `orm:"rel(fk)"`
  Role *Role `orm:"rel(fk)"`

  Session *db.Session `orm:"-"`
}

func (this *UserRole) TableName() string{
  return "user_roles"
}

func NewUserRole(session *db.Session) *UserRole{
  return &UserRole{ Session: session }
}

func (this *UserRole) IsPersisted() bool{
  return this.Id > 0
}

func NewUserRoleWithRole(user *User, role *Role) *UserRole {
  entity := UserRole{User: user, Role: role}
  return &entity
}

func (this *UserRole) FindRoleByUser(user *User) *Role {

  entity := new(UserRole)

  query, _ := this.Session.Query(entity)

  err := query.Filter("User", user).One(entity)

  if err == orm.ErrNoRows {
    return nil
  }  

  this.Session.GetDb().LoadRelated(entity, "Role")

  return entity.Role
}

func (this *UserRole) FindAllRolesByUser(user *User) *[]*Role {

  results := new([]*UserRole)

  query, _ := this.Session.Query(new(UserRole))

  query.Filter("User", user).All(results)

	roles := []*Role{}

	for _, it := range *results {
		this.Session.GetDb().LoadRelated(it, "Role")

		roles = append(roles, it.Role)
	}

  return &roles
}

func (this *UserRole) HasRoles(user *User, roles ...string) bool {

  userRoles := this.FindAllRolesByUser(user)

  if userRoles != nil {
    for _, it := range *userRoles {

      for _, role := range roles {
        if it.Authority == role {
          return true
        }
      }

    }
  }

  return false

}

func (this *UserRole) FindAllByRole(role *Role) (*[]*UserRole, error) {

  results := []*UserRole{}

  query, _ := this.Session.Query(new(UserRole))

  _, err := query.Filter("Role", role).All(&results)

	for _, it := range results {
		this.Session.GetDb().LoadRelated(it, "Role")
		this.Session.GetDb().LoadRelated(it, "User")
	}

  return &results, err
}

func (this *UserRole) FindAllByUser(user *User) (*[]*UserRole, error) {

  results := []*UserRole{}

  query, _ := this.Session.Query(new(UserRole))

  _, err := query.Filter("User", user).All(&results)

  for _, it := range results {
    this.Session.GetDb().LoadRelated(it, "Role")
  }

  return &results, err
}

func (this *UserRole) FindByUser(user *User) (*UserRole, error) {

  entity := new(UserRole)

  query, _ := this.Session.Query(entity)

  err := query.Filter("User", user).One(entity)

  if err == orm.ErrNoRows {
    return entity, nil  
  }  

  this.Session.GetDb().LoadRelated(entity, "Role")

  return entity, err
}

func (this *UserRole) FindByUserAndRole(user *User, role *Role) (*UserRole, error) {

  entity := new(UserRole)

  query, _ := this.Session.Query(entity)

  err := query.Filter("User", user).Filter("Role", role).One(entity)

  if err == orm.ErrNoRows {
    return entity, nil  
  }

  return entity, err
}

func (this *UserRole) Create(user *User, autority string) error { 

  search := NewRole(this.Session)
  role, err := search.FindByAuthority(autority)

  if err != nil {
    return err
  }

  if role == nil || !role.IsPersisted() {
    return errors.New(fmt.Sprintf("role %v not found", autority))
  }

  entity, err := this.FindByUserAndRole(user, role)

  if err != nil && err != orm.ErrNoRows {
    return err
  }

  if entity != nil && entity.IsPersisted() {
    return nil
  }

  entity = &UserRole{ User: user, Role: role }

  return this.Session.Save(entity)

}
