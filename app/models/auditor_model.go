package models

import (
	"time"

	"github.com/mobilemindtec/go-utils/beego/db"
)

type Auditor struct {
	Id        int64     `form:"-" json:",string,omitempty"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`
	Content   string    `orm:"type(text)"  valid:"Required;MaxSize(300)" form:""`

	Tenant *Tenant `orm:"null;rel(fk);on_delete(do_nothing)" valid:""`
	User   *User   `orm:"null;rel(fk);on_delete(do_nothing)"`

	Session *db.Session `orm:"-" inject:""`
}

func NewAuditor(session *db.Session) *Auditor {
	return &Auditor{Session: session}
}

func (this *Auditor) IsPersisted() bool {
	return this.Id > 0
}

func NewAuditorWithTenant(tenant *Tenant) *Auditor {
	return &Auditor{Tenant: tenant}
}

func NewAuditorWithTenantAndContent(tenant *Tenant, content string) *Auditor {
	return &Auditor{Tenant: tenant, Content: content}
}

func (this *Auditor) TableName() string {
	return "auditoria"
}

func (this *Auditor) LoadRelated(entity *Auditor) {
	this.Session.GetDb().LoadRelated(entity, "Tenant")
	this.Session.GetDb().LoadRelated(entity, "User")
}

func (this *Auditor) ListByTenant(tenant *Tenant) (*[]*Auditor, error) {
	var results []*Auditor

	query, err := this.Session.Query(this)

	if err != nil {
		return nil, err
	}

	if err := this.Session.ToList(query.Filter("Tenant", tenant), &results); err != nil {
		return nil, err
	}

	return &results, err
}

func (this *Auditor) List() (*[]*Auditor, error) {
	var results []*Auditor
	err := this.Session.List(this, &results)
	return &results, err
}
