package models

import (
	"time"

	"github.com/mobilemindtech/go-utils/beego/db"
)

// UserRefreshToken refresh token opaco usado para renovar o bearer token (JWT).
// Só o hash sha256 do token é persistido. Cada login inicia uma família
// (Family); cada refresh revoga o token usado e emite um novo na mesma
// família. Se um token já revogado for reutilizado, a família inteira é
// revogada (indício de token vazado).
type UserRefreshToken struct {
	Id        int64     `form:"-" json:",string,omitempty"`
	CreatedAt time.Time `orm:"auto_now_add;type(datetime)" json:"-"`
	UpdatedAt time.Time `orm:"auto_now;type(datetime)" json:"-"`

	TokenHash string    `orm:"size(64);unique" json:"-"`
	Family    string    `orm:"size(100);index" json:"-"`
	ExpiresAt time.Time `orm:"type(datetime)" json:"-"`
	RevokedAt time.Time `orm:"type(datetime);null" json:"-"`
	UserAgent string    `orm:"size(255);null" json:"-"`

	User *User `orm:"rel(fk);on_delete(do_nothing)" valid:"Required"`

	Session *db.Session `orm:"-" json:"-" inject:""`
}

func NewUserRefreshToken(session *db.Session) *UserRefreshToken {
	return &UserRefreshToken{Session: session}
}

func (this *UserRefreshToken) TableName() string {
	return "user_refresh_tokens"
}

func (this *UserRefreshToken) IsPersisted() bool {
	return this.Id > 0
}

func (this *UserRefreshToken) IsRevoked() bool {
	return !this.RevokedAt.IsZero()
}

func (this *UserRefreshToken) IsExpired(now time.Time) bool {
	return now.Unix() > this.ExpiresAt.Unix()
}
