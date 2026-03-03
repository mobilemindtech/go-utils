package features

import (
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/beego/web/misc"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type WebModels struct {
	ModelAuditor    *models.Auditor
	ModelCidade     *models.Cidade
	ModelEstado     *models.Estado
	ModelRole       *models.Role
	ModelTenant     *models.Tenant
	ModelUser       *models.User
	ModelTenantUser *models.TenantUser
	ModelUserRole   *models.UserRole
	base            trait.WebBaseInterface
}

func (this *WebModels) InitWebModels(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebModels) WebControllerLoadModels() {
	if this.base.GetInheritedController() != nil {
		if app, ok := this.base.GetInheritedController().(misc.NestWebController); ok {
			app.WebControllerLoadModels()
		}
	}

	this.LoadModels()
}

func (this *WebModels) LoadModels() {
	this.ModelAuditor = models.NewAuditor(this.base.GetSession())
	this.ModelCidade = models.NewCidade(this.base.GetSession())
	this.ModelEstado = models.NewEstado(this.base.GetSession())
	this.ModelRole = models.NewRole(this.base.GetSession())
	this.ModelTenant = models.NewTenant(this.base.GetSession())
	this.ModelUser = models.NewUser(this.base.GetSession())
	this.ModelTenantUser = models.NewTenantUser(this.base.GetSession())
	this.ModelUserRole = models.NewUserRole(this.base.GetSession())
}
