package features

import (
	"github.com/mobilemindtech/go-utils/app/models"
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
	auditorService := services.NewAuditorService(this.base.GetSession(), this.base.GetLang(), this.GetAuditorInfo())
	auditorService.OnAuditWithNewDbSession(format, v...)
}

func (this *WebAudit) GetAuditorInfo() *services.AuditorInfo {
	return &services.AuditorInfo{Tenant: this.base.GetAuthTenant(), User: this.base.GetAuthUser()}
}

func (this *WebAudit) AuditWithTenantId(tenantId int64, format string, v ...interface{}) {
	auditorInfo := &services.AuditorInfo{Tenant: models.NewTenantWithId(tenantId), User: this.base.GetAuthUser()}
	auditorService := services.NewAuditorService(this.base.GetSession(), this.base.GetLang(), auditorInfo)
	auditorService.OnAuditWithNewDbSession(format, v...)
}
