package models

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/client/orm"
	"github.com/mobilemindtech/go-utils/app/util"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/support"
	uuid "github.com/satori/go.uuid"
	"strings"
)

var (
	UserHexPassword = false
)

func OnUserHexPassword() { UserHexPassword = true }

type User struct {
	Id        int64     `form:"-" json:",string,omitempty"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`

	Name      string    `orm:"size(100)"  valid:"Required;MaxSize(100)" form:"" json:""`
	UserName  string    `orm:"size(100);unique" valid:"Required;MaxSize(100);Email" form:""`
	Password  string    `orm:"size(100)" valid:"MaxSize(100)" form:"" json:"-"`
	Enabled   bool      `orm:"" valid:"Required;" form:"" json:""`
	LastLogin time.Time `orm:"null;type(datetime)"`

	ExpirationDate time.Time `orm:"type(datetime);null" form:"-" json:"-"`
	Token          string    `orm:"type(text);null"  valid:"" form:"-" json:"-"`

	Uuid string `orm:"size(100);unique"  valid:"MaxSize(100)" form:"-" json:""`

	ChangePwdExpirationDate time.Time `orm:"type(datetime);null" form:"-" json:"-"`
	ChangePwdToken          string    `orm:"type(text);null"  valid:"MaxSize(256)" form:"-" json:"-"`

	Tenant *Tenant `orm:"rel(fk);on_delete(do_nothing)" valid:"" form:"" goutils:"no_set_tenant;no_filter_tenant"`

	Role  *Role    `orm:"-"`
	Roles *[]*Role `orm:"-"`

	Session *db.Session `orm:"-" json:"-" inject:""`
}

func NewUser(session *db.Session) *User {
	return &User{Session: session}
}

func (this *User) TableName() string {
	return "users"
}

func (this *User) GetAuthorities() string {
	authorities := ""

	if this.Roles != nil {
		for _, it := range *this.Roles {
			authorities = fmt.Sprintf("%v, %v", authorities, it.Authority)
		}
	}

	return authorities
}

func (this *User) LoadIfExists() (bool, error) {

	err := this.Session.GetDb().Read(this)

	if err == orm.ErrNoRows {
		return false, nil
	}

	if err != nil {
		return false, err
	}

	return true, nil
}

func (this *User) GetByUserName(username string) (*User, error) {
	result := new(User)

	query, err := this.Session.Query(this)

	if err != nil {
		return nil, err
	}

	err = query.Filter("UserName", username).One(result)

	if err == orm.ErrNoRows {
		return nil, nil
	} else if err == orm.ErrMultiRows {
		return nil, err
	}

	return result, err
}

func (this *User) GetByToken(token string) (*User, error) {
	result := new(User)

	query, err := this.Session.Query(this)

	if err != nil {
		return nil, err
	}

	err = query.Filter("Token", token).One(result)

	if err == orm.ErrNoRows {
		return nil, nil
	} else if err == orm.ErrMultiRows {
		return nil, err
	}

	return result, err
}

func (this *User) EncodePassword() {
	this.GenerateToken(this.Password)
	if UserHexPassword {
		this.Password = support.TextToSha1Hex(this.Password)
	} else {
		this.Password = support.TextToSha1(this.Password)
	}
}

func (this *User) ChangePassword(newPassword string) {
	this.Password = newPassword
	this.GenerateToken(this.Password)
	this.EncodePassword()
}

func (this *User) GenereteUuid() string {

	for true {
		uuid := uuid.NewV4()
		if !db.NewCriteria(this.Session, new(User), nil).Eq("Uuid", uuid.String()).Exists() {
			return uuid.String()
		}
	}

	return ""
}

func (this *User) IsSamePassword(newPassword string) bool {

	if UserHexPassword {
		return support.IsSameHashHex(this.Password, newPassword)
	} else {
		return support.IsSameHash(this.Password, newPassword)
	}
}

func (this *User) IsPersisted() bool {
	return this.Id > 0
}

func (this *User) LoadRelated(entity *User) {
	userRole := NewUserRole(this.Session)
	entity.Role = userRole.FindRoleByUser(entity)
	entity.Roles = userRole.FindAllRolesByUser(entity)
}

func (this *User) List() (*[]*User, error) {
	var results []*User
	err := this.Session.List(this, &results)
	return &results, err
}

func (this *User) ListByTenant(tenant *Tenant) (*[]*User, error) {
	var results []*User

	query, err := this.Session.Query(this)

	if err != nil {
		return nil, err
	}

	if err := this.Session.ToList(query.Filter("Tenant", tenant), &results); err != nil {
		return nil, err
	}

	return &results, err
}

func (this *User) Page(page *db.Page) (*[]*User, error) {
	var results []*User

	page.AddFilterDefault("Name").MakeDefaultSort()

	err := this.Session.Page(this, &results, page)
	return &results, err
}

func (this *User) PageByTenant(tenant Tenant, page *db.Page) (*[]*User, error) {
	var results []*User

	page.AddFilterDefault("Name").MakeDefaultSort()

	query, err := this.Session.Query(this)

	if err != nil {
		return nil, err
	}

	if err := this.Session.ToPage(query.Filter("Tenant", tenant), &results, page); err != nil {
		return nil, err
	}

	return &results, err
}

func (this *User) GenerateToken(password string) {

	this.ExpirationDate = time.Now().In(util.GetDefaultLocation()).AddDate(50, 0, 0)

	var err error
	if this.Token, err = support.GenereteApiToken(this.Id, this.Uuid, password, this.ExpirationDate); err != nil {
		fmt.Println("** error on generete api token: %v", err)
	}

}

func (this *User) GetByChangePwdToken(token string) (*User, error) {
	result := new(User)

	query, err := this.Session.Query(this)

	if err != nil {
		return nil, err
	}

	err = query.Filter("ChangePwdToken", token).One(result)

	if err == orm.ErrNoRows {
		return nil, nil
	} else if err == orm.ErrMultiRows {
		return nil, err
	}

	return result, err
}

func (this *User) GetByUuid(uuid string) (*User, error) {

	entity := new(User)
	criteria := db.NewCriteria(this.Session, entity, nil).Eq("Uuid", uuid).One()

	return entity, criteria.Error
}

func (this *User) GetByUuidAndEnabled(uuid string) (*User, error) {

	entity := new(User)
	criteria := db.NewCriteria(this.Session, entity, nil).Eq("Uuid", uuid).Eq("Enabled", true).One()

	return entity, criteria.Error
}

func (this *User) UpdateLastLogin(userId int64) {
	query := "update users set last_login = now() where id = ?"
	_, err := db.NewRawQueryArgs(this.Session, query, userId).Execute()

	if err != nil {
		fmt.Println("erro ao atualizar last logn user")
	} else {
		fmt.Println("last login update success")
	}
}
func (this *User) FirstName() string {
	sp := strings.Split(this.Name, " ")
	if len(sp) > 0 {
		return sp[0]
	}
	return this.Name
}

func (this *User) String() string {
	return fmt.Sprintf("%v - %v", this.Id, this.Name)
}
