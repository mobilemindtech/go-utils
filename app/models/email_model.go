package models

import (
	"time"
  "github.com/mobilemindtec/go-utils/beego/db"
)

type Email struct{
    
  Id int64 `form:"-" json:",string,omitempty"`
  CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
  UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`      

  To string `orm:"size(100)"  valid:"Required;MaxSize(100)"`
  Cco string `orm:"size(100)"  valid:"Required;MaxSize(100)"`
  Subject string `orm:"size(100)"  valid:"Required;MaxSize(50)"`
  Body string `orm:"type(text)"  valid:"Required;"`
  Enabled bool `orm:""  valid:"" `

  Tenant *Tenant `orm:"rel(fk);on_delete(do_nothing)" form:"" goutils:"tenant"`

  Session *db.Session `orm:"-"`
}

func (this *Email) TableName() string{
  return "emails"
}

func NewEmail(session *db.Session) *Email{
  return &Email{ Session: session }
}

func (this *Email) IsPersisted() bool{
  return this.Id > 0
}

func (this *Email) List() (*[]*Email , error) { 
  var results []*Email

  query, err := this.Session.Query(this)


  if err != nil {
    return nil, err
  }

  if err := this.Session.ToList(query.Filter("Enabled", true), &results); err != nil {
    return nil, err
  }

  return &results, err
}


func (this *Email) Create(to string, subject string, body string) (*Email, error){
  
  email := new(Email)
  email.To = to  
  email.Subject = subject
  email.Body = body
  email.Enabled = true

  return email, this.Session.Save(email)
}

func (this *Email) CreateWithCco(to string, cco string, subject string, body string) (*Email, error){
  
  email := new(Email)
  email.To = to
  email.Cco = cco
  email.Subject = subject
  email.Body = body
  email.Enabled = true

  return email, this.Session.Save(email)
}