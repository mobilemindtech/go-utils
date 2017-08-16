package models

import (
	"github.com/mobilemindtec/go-utils/beego/db"
	"time"
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

  query.Filter("User", user).One(entity)

  this.Session.Db.LoadRelated(entity, "Role")

  return entity.Role
}

func (this *UserRole) FindAllRolesByUser(user *User) *[]*Role {

  results := new([]*UserRole)

  query, _ := this.Session.Query(new(UserRole))

  query.Filter("User", user).All(results)

	roles := []*Role{}

	for _, it := range *results {
		this.Session.Db.LoadRelated(it, "Role")

		roles = append(roles, it.Role)
	}

  return &roles
}

func (this *UserRole) FindAllByRole(role *Role) (*[]*UserRole, error) {

  results := []*UserRole{}

  query, _ := this.Session.Query(new(UserRole))

  _, err := query.Filter("Role", role).All(&results)

	for _, it := range results {
		this.Session.Db.LoadRelated(it, "Role")
		this.Session.Db.LoadRelated(it, "User")
	}

  return &results, err
}

func (this *UserRole) FindByUser(user *User) (*UserRole, error) {

  entity := new(UserRole)

  query, _ := this.Session.Query(entity)

  err := query.Filter("User", user).One(entity)

  this.Session.Db.LoadRelated(entity, "Role")

  return entity, err
}
