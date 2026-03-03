package features

import (
	"github.com/mobilemindtech/go-utils/app/services"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type WebAudit struct {
	base trait.WebBaseInterface
}

func (this *WebAudit) InitWebAudit(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebAudit) Audit(format string, v ...interface{}) {
	auditor := services.NewAuditorService(this.base.GetSession(), this.base.GetLang(), this.GetAuditorInfo())
	auditor.OnAuditWithNewDbSession(format, v...)
}

func (this *WebAudit) GetAuditorInfo() *services.AuditorInfo {
	return &services.AuditorInfo{Tenant: this.base.GetAuthTenant(), User: this.base.GetAuthUser()}
}
