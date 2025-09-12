package models

import (
	"github.com/mobilemindtech/go-utils/beego/db"
	"time"
)

type App struct {
	Id        int64     `form:"-" json:",string,omitempty"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)"`

	Name      string    `orm:"size(50)"  valid:"Required;MinSize(2);MaxSize(50)" form:"" json:""`
	Token     string    `orm:"size(200)"  valid:"Required;MaxSize(200)" form:""  json:""`
	LastLogin time.Time `orm:"type(datetime);null"  json:""`

	Tenant *Tenant `orm:"rel(fk);on_delete(do_nothing)" valid:"" form:"" goutils:"tenant"`

	Session *db.Session `orm:"-" json:"-"`

	Count int64 `orm:"-"`
}

func NewApp(session *db.Session) *App {
	return &App{Session: session}
}

func (this *App) TableName() string {
	return "apps"
}

func (this *App) IsPersisted() bool {
	return this.Id > 0
}
