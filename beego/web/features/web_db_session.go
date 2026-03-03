package features

import (
	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-utils/assert"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/web/misc"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type WebDbSession struct {
	Session *db.Session
	base    trait.WebBaseInterface
}

func (this *WebDbSession) InitWebDbSession(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebDbSession) WebControllerCreateSession() {
	var session *db.Session
	if this.base.GetInheritedController() != nil {
		if app, ok := this.base.GetInheritedController().(misc.NestWebController); ok {
			session = app.WebControllerCreateSession()
		}
	}
	if session == nil {
		session = this.CreateSession()
	}

	assert.Assert(session != nil, "session is nil")

	this.Session = session
}

func (this *WebDbSession) CreateSession() *db.Session {

	session := db.NewSession()
	err := session.OpenTx()

	if err != nil {
		logs.Error("error on create db session: %v", err)
		this.base.GetBeegoController().Abort("500")
	}

	return session
}

func (this *WebDbSession) RollbackDbSession() {
	if this.Session != nil {
		this.Session.OnError()
	}
}

func (this *WebDbSession) CloseDbSession() {
	if this.Session != nil {
		this.Session.Close()
		this.Session = nil
	}
}

func (this *WebDbSession) CloseDbSessionWithError() {
	if this.Session != nil {
		this.Session.OnError().Close()
		this.Session = nil
	}
}
