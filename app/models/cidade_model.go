package models

import (
	"github.com/mobilemindtec/go-utils/beego/db"
)

type Cidade struct{    
  Id int64 `form:"-" json:",string,omitempty"`
  Nome string `orm:"size(100)"  valid:"Required;MaxSize(100)" form:""`
  Estado *Estado `orm:"rel(fk);on_delete(do_nothing)" valid:"Required;" form:""`

  Session *db.Session `orm:"-"`
}

func NewCidade(session *db.Session) *Cidade{
  return &Cidade{ Session: session }
}

func (this *Cidade) TableName() string{
  return "cidades"
}

func (this *Cidade) IsPersisted() bool{
  return this.Id > 0
}

func (this *Cidade) LoadRelated(entity *Cidade) { 
  this.Session.Db.LoadRelated(entity, "Estado")
}

func (this *Cidade) ListByEstado(estado *Estado) (*[]*Cidade , error) { 
  var results []*Cidade

  query, err := this.Session.Query(this)


  if err != nil {
    return nil, err
  }

  if err := this.Session.ToList(query.Filter("Estado", estado), &results); err != nil {
    return nil, err
  }

  return &results, err
}