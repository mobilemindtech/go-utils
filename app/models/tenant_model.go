package models

import (
	"time"

	"github.com/mobilemindtech/go-utils/beego/db"
	uuid "github.com/satori/go.uuid"
)

type Tenant struct {
	Id        int64     `form:"-" json:",string,omitempty"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`

	Name      string `orm:"size(100)"  valid:"Required;MaxSize(100)" form:""`
	Documento string `orm:"size(20)"  valid:"Required;MaxSize(14);MinSize(11)" form:""`

	Enabled bool   `orm:""  form:"" json:",string"`
	Uuid    string `orm:"size(100);unique"  valid:"MaxSize(100)" form:"-" json:""`

	Cidade *Cidade `orm:"rel(fk);on_delete(do_nothing)" valid:"RequiredRel" form:""`

	Session *db.Session `orm:"-" json:"-" inject:""`
}

func (this *Tenant) TableName() string {
	return "tenants"
}

func NewTenant(session *db.Session) *Tenant {
	return &Tenant{Session: session}
}

func NewTenantWithId(id int64) *Tenant {
	return &Tenant{Id: id}
}

func (this *Tenant) IsPersisted() bool {
	return this.Id > 0
}

func (this *Tenant) GetId() int64 {
	return this.Id
}

func (this *Tenant) GenereteUuid() string {

	for true {
		uuid := uuid.NewV4()
		if !db.NewCriteria(this.Session, new(Tenant), nil).Eq("Uuid", uuid.String()).Exists() {
			return uuid.String()
		}
	}

	return ""
}

func (this *Tenant) First() *Tenant {

	first := new(Tenant)
	criteria := db.NewCriteria(this.Session, first, nil)
	criteria.OrderAsc("Id")
	criteria.One()

	if criteria.Empty {
		return nil
	}

	return first
}

func (this *Tenant) List() (*[]*Tenant, error) {
	entities := []*Tenant{}
	criteria := db.NewCriteria(this.Session, new(Tenant), &entities)
	criteria.OrderAsc("Name")
	criteria.List()

	return &entities, criteria.Error
}

func (this *Tenant) Page(page *db.Page) (*[]*Tenant, error) {
	var results []*Tenant

	page.AddFilterDefault("Name").MakeDefaultSort()

	err := this.Session.Page(this, &results, page)
	return &results, err
}

func (this *Tenant) GetByUuid(uuid string) (*Tenant, error) {

	entity := new(Tenant)
	criteria := db.NewCriteria(this.Session, entity, nil).Eq("Uuid", uuid).One()

	return entity, criteria.Error
}

func (this *Tenant) GetByUuidAndEnabled(uuid string) (*Tenant, error) {

	entity := new(Tenant)
	criteria := db.NewCriteria(this.Session, entity, nil).Eq("Uuid", uuid).Eq("Enabled", true).One()

	return entity, criteria.Error

}

func (this *Tenant) FindByDocumento(documento string) (*Tenant, error) {

	entity := new(Tenant)
	criteria := db.NewCriteria(this.Session, entity, nil).Eq("Documento", documento).One()

	return entity, criteria.Error

}

func (this *Tenant) LoadRelated(entity *Tenant) {

	this.Session.Load(entity.Cidade)
	this.Session.Load(entity.Cidade.Estado)

}
