package controllers

import (
  "github.com/mobilemindtec/go-utils/app/services"
  "github.com/mobilemindtec/go-utils/app/models"
  "github.com/mobilemindtec/go-utils/beego/web"  
  "time"
)

type BaseController struct {
  web.BaseController
  BaseAuthController


  ModelAuditor *models.Auditor
  ModelCidade *models.Cidade
  ModelEstado *models.Estado
  ModelRole *models.Role
  ModelTenant *models.Tenant
  ModelUser *models.User
  ModelTenantUser *models.TenantUser
  ModelUserRole *models.UserRole  

}

func (this *BaseController) NestPrepareAppBase() {

  this.Log("** app.controllers.BaseController.NestPrepare")

  this.NestPrepareBase()
  this.LoadModels()

  this.NestPrepareAuth(this)

  if this.IsLoggedIn || this.IsTokenLoggedIn {
    this.Session.Tenant = this.GetAuthTenant()
  }

  this.Data["today"] = time.Now().In(this.DefaultLocation).Format("02.01.2006")

  this.loadTenants()
}

func (this *BaseController) loadTenants(){
  tenants := []*models.Tenant{}
  
  if this.IsLoggedIn {

    if this.Auth.IsRoot() {
      its, _ := this.ModelTenant.List()
      tenants = *its
    } else {
      list, _ := this.ModelTenantUser.ListByUser(this.GetAuthUser())


      for _, it := range *list {
        this.Session.Load(it.Tenant)
        tenants = append(tenants, it.Tenant)
      }
    }
    
  }

  this.Data["AvailableTenants"] = tenants  
}

func (this *BaseController) OnAuditor(format string, v ...interface{}) {
  auditor := services.NewAuditorService(this.Session, this.Lang, this.GetAuditorInfo())
  auditor.OnAudit(format, v...)
}

func (this *BaseController) GetAuditorInfo() *services.AuditorInfo{
  return &services.AuditorInfo{ Tenant: this.GetAuthTenant(), User: this.GetAuthUser() }
}

func (this *BaseController) GetLastUpdate() time.Time{
  lastUpdateUnix, _ := this.GetInt64("lastUpdate")
  var lastUpdate time.Time

  if lastUpdateUnix > 0 {
    lastUpdate = time.Unix(lastUpdateUnix, 0).In(this.DefaultLocation)
  }

  return lastUpdate
}



func (this *BaseController) LoadModels() {
  this.ModelAuditor = models.NewAuditor(this.Session)
  this.ModelCidade = models.NewCidade(this.Session)
  this.ModelEstado = models.NewEstado(this.Session)
  this.ModelRole = models.NewRole(this.Session)
  this.ModelTenant = models.NewTenant(this.Session)
  this.ModelUser = models.NewUser(this.Session)
  this.ModelTenantUser = models.NewTenantUser(this.Session)
  this.ModelUserRole = models.NewUserRole(this.Session)
}

func init() {
  
}