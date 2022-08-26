package v2

import (
	"github.com/mobilemindtec/go-utils/beego/db"
)

func NewSession() *db.Session{
	return db.NewSession()
}

func NewSessionWithDbName(dbName string) *db.Session{
  return db.NewSessionWithDbName(dbName)
}
