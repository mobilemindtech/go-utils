package models

import (
	"github.com/mobilemindtec/go-utils/beego/db"
)

type Estado struct {
	Id   int64  `form:"-" json:",string,omitempty"`
	Nome string `orm:"size(100)"  valid:"Required;MaxSize(100)" form:""`
	Uf   string `orm:"size(2)" valid:"Required;MaxSize(2)" form:""`

	Session *db.Session `orm:"-" json:"-" inject:""`
}

func NewEstado(session *db.Session) *Estado {
	return &Estado{Session: session}
}

func (this *Estado) TableName() string {
	return "estados"
}

func (this *Estado) IsPersisted() bool {
	return this.Id > 0
}

func (this *Estado) List() (*[]*Estado, error) {
	var results []*Estado
	err := this.Session.List(this, &results)
	return &results, err
}

func (this *Estado) FindByUf(uf string) (*Estado, error) {

	result := new(Estado)

	criteria := db.NewCriteria(this.Session, result, nil).Eq("Uf", uf).One()

	return result, criteria.Error
}
