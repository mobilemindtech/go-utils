package services

import (
	"fmt"
	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/beego/db"
)

type AuditorInfo struct {
	Tenant *models.Tenant
	User   *models.User
}

type AuditorService struct {
	Lang        string
	AuditorInfo *AuditorInfo
	Session     *db.Session
}

func NewAuditorService(session *db.Session, lang string, info *AuditorInfo) *AuditorService {
	return &AuditorService{Lang: lang, AuditorInfo: info, Session: session}
}

func (this *AuditorService) OnAudit(format string, v ...interface{}) {
	content := fmt.Sprintf(format, v...)
	auditor := models.NewAuditorWithTenantAndContent(this.AuditorInfo.Tenant, content)
	auditor.User = this.AuditorInfo.User

	if err := this.Session.Save(auditor); err != nil {
		logs.Debug("## error on save auditor: ", err.Error())
	}
}

func (this *AuditorService) OnAuditWithNewDbSession(format string, v ...interface{}) {

	action := func() {
		content := fmt.Sprintf(format, v...)
		auditor := models.NewAuditorWithTenantAndContent(this.AuditorInfo.Tenant, content)
		auditor.User = this.AuditorInfo.User

		session := db.NewSession()
		err := session.OpenNoTx()

		if err != nil {
			logs.Debug("## error on open session: ", err.Error())
			return
		}

		defer session.Close()

		if err := session.Save(auditor); err != nil {
			logs.Debug("## error on save auditor: ", err.Error())
		}
	}

	go action()

}
