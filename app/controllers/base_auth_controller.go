package controllers

import (
  "github.com/mobilemindtec/go-utils/app/services"
  "github.com/mobilemindtec/go-utils/app/models"
  "github.com/mobilemindtec/go-utils/app/route"
  "strings"
)

type BaseAuthController struct{

  userinfo *models.User
  tenant *models.Tenant

  IsLoggedIn  bool
  IsTokenLoggedIn  bool

  Auth *services.AuthService

  baseController *BaseController
}

func (this *BaseAuthController) NestPrepareAuth(base *BaseController) {

  base.Log("** app.controllers.BaseAuthController.NestPrepareAuth")

  this.baseController = base

  // login
  this.AppAuth()
  this.SetParams()

  this.IsLoggedIn = this.baseController.GetSession("userinfo") != nil
  this.IsTokenLoggedIn = this.baseController.GetSession("appuserinfo") != nil

  if this.IsLoggedIn || this.IsTokenLoggedIn {

    if this.IsLoggedIn {
      this.SetAuthUser(this.GetLogin())
    } else {
      this.SetAuthUser(this.GetTokenLogin())
    }

    var tenant *models.Tenant

    if this.IsTokenLoggedIn {

      ModelTenant := this.baseController.ModelTenant
      tenantUuid := this.baseController.GetHeaderByName("tenant")

      this.baseController.Log("tenantUuid = %v", tenantUuid)

      tenant, _ = ModelTenant.GetByUuidAndEnabled(tenantUuid)

    } else {

      tenant = this.GetAuthTenantSession()

      if tenant == nil {
        ModelTenantUser := this.baseController.ModelTenantUser
        tenant, _ = ModelTenantUser.GetFirstTenant(this.GetAuthUser())
      }

    }


    if tenant == nil || !tenant.IsPersisted() {
      
      this.baseController.Log("** user does not has active tenant")

      if this.IsTokenLoggedIn && !this.baseController.IsJson() {
        this.baseController.OnJsonError("set header tenant")
      } else {
        this.baseController.OnErrorAny("/", "user does not has active tenant")
      }
      return
    }

    this.SetAuthTenant(tenant)

    this.baseController.Log("**********************************")
    this.baseController.Log("** IsLoggedIn=%v", this.IsLoggedIn)
    this.baseController.Log("** UserInfo.Id=%v", this.GetAuthUser().Id)
    this.baseController.Log("** UserInfo.Name=%v", this.GetAuthUser().Name)
    this.baseController.Log("** UserInfo.Authority=%v", this.GetAuthUser().Role.Authority)
    this.baseController.Log("** Tenant.Name=%v", this.GetAuthTenant().Name)
    this.baseController.Log("**********************************")

    this.baseController.Data["UserInfo"] = this.GetAuthUser()
    this.baseController.Data["Tenant"] = this.GetAuthTenant()

    this.Auth = services.NewAuthService(this.GetAuthUser())
  }

  this.baseController.Data["IsLoggedIn"] = this.IsLoggedIn || this.IsTokenLoggedIn

  if this.IsLoggedIn || this.IsTokenLoggedIn {
    this.baseController.Data["IsAdmin"] = this.Auth.IsAdmin()
    this.baseController.Data["IsRoot"] = this.Auth.IsRoot()
  }

  this.UpSecurityAuth()

}

func (this *BaseAuthController) AppAuth(){

  token := this.baseController.GetToken()

  if strings.TrimSpace(token) != "" {

    auth := services.NewLoginService(this.baseController.Lang, this.baseController.Session)

    user, err := auth.AuthenticateToken(token)


    if err == nil && user != nil && user.Id > 0{
      this.baseController.ModelUser.LoadRelated(user)
      this.SetTokenLogin(user)
    }
  }
}

func (this *BaseAuthController) GetLogin() *models.User {
  id := this.baseController.GetSession("userinfo").(int64)
  e, err := this.baseController.Session.FindById(new(models.User), id)
  if err != nil {
    return nil
  }
  user := e.(*models.User)
  this.baseController.ModelUser.LoadRelated(user)
  return user
}

func (this *BaseAuthController) GetTokenLogin() *models.User {
  id := this.baseController.GetSession("appuserinfo").(int64)
  e, err := this.baseController.Session.FindById(new(models.User), id)
  if err != nil {
    return nil
  }
  user := e.(*models.User)
  this.baseController.ModelUser.LoadRelated(user)
  return user
}

func (this *BaseAuthController) DelLogin() {
  this.baseController.DelSession("userinfo")
}

func (this *BaseAuthController) SetLogin(user *models.User) {
  this.baseController.SetSession("userinfo", user.Id)
}

func (this *BaseAuthController) SetTokenLogin(user *models.User) {
  this.baseController.SetSession("appuserinfo", user.Id)
}

func (this *BaseAuthController) DelTokenLogin() {
  this.baseController.DelSession("appuserinfo")
}

func (this *BaseAuthController) LoginPath() string {
  return this.baseController.URLFor("LoginController.Login")
}

func (this *BaseAuthController) SetParams() {
  this.baseController.Data["Params"] = make(map[string]string)
  for k, v := range this.baseController.Input() {
    this.baseController.Data["Params"].(map[string]string)[k] = v[0]
  }
}

func (this *BaseAuthController) OnLoginRedirect() {
  path := this.baseController.Ctx.Input.URI()
  if !strings.Contains("?", path) {
    path = "?redirect=" + path
  }
  this.baseController.Ctx.Redirect(302, this.LoginPath() + path)
}

func (this *BaseAuthController) AuthCheck() {
  if !this.IsLoggedIn && !this.IsTokenLoggedIn {
    if this.baseController.IsJson(){

      this.baseController.OnJsonError(this.baseController.GetMessage("security.notLoggedIn"))
      this.baseController.Abort("401")

    } else  {
      this.OnLoginRedirect()
    }
  }
}

func (this *BaseAuthController) AuthCheckRoot() {
  if !this.IsLoggedIn {
    if this.baseController.IsJson() {
      this.baseController.OnJsonError(this.baseController.GetMessage("security.notLoggedIn"))
      this.baseController.Abort("401")
    } else {
      this.OnLoginRedirect()
    }
  }

  if !this.Auth.IsRoot() {
    if this.baseController.IsJson() {
      this.baseController.OnJsonError(this.baseController.GetMessage("security.rootRequired"))
      this.baseController.Abort("401")
    } else {
      this.baseController.OnRedirect("/")
    }
  }
}

func (this *BaseAuthController) AuthCheckAdmin() {
  if !this.IsLoggedIn {
    if this.baseController.IsJson() {
      this.baseController.OnJsonError(this.baseController.GetMessage("security.notLoggedIn"))
      this.baseController.Abort("401")
    } else {
      this.OnLoginRedirect()
      this.baseController.OnRedirectError("/", this.baseController.GetMessage("security.rootRequired"))
    }
  }

  if !this.Auth.IsRoot() && !this.Auth.IsAdmin() {
    if this.baseController.IsJson() {
      this.baseController.OnJsonError(this.baseController.GetMessage("security.rootRequired"))
      this.baseController.Abort("401")
    } else {
      this.baseController.OnRedirectError("/", this.baseController.GetMessage("security.rootRequired"))
    }
  }
}

func (this *BaseAuthController) UpSecurityAuth() bool {

  this.baseController.Log("** UpSecurityAuth")

  roles := []string{}

  if this.Auth != nil {
    roles = this.Auth.GetUserRoles()
  }

  if !route.IsRouteAuthorized(this.baseController.Ctx, roles) {

    this.baseController.Log("** not authorized ")

    if !this.IsLoggedIn && !this.IsTokenLoggedIn {
      if this.baseController.IsJson(){
        this.baseController.OnJsonError(this.baseController.GetMessage("security.notLoggedIn"))
        //this.baseController.Abort("401")
      } else  {
        this.baseController.OnLoginRedirect()
      }
      return false
    }

    if this.baseController.IsJson() {
      this.baseController.OnJsonError(this.baseController.GetMessage("security.denied"))
      //this.baseController.Abort("401")
    } else {
      this.baseController.OnRedirect("/")
    }

    return false

  }

  return true
}

func (this *BaseAuthController) HasTenantAuth(tenant *models.Tenant) bool{
  if !this.Auth.IsRoot() {

    ModelTenantUser := this.baseController.ModelTenantUser
    item, _ := ModelTenantUser.FindByUserAndTenant(this.GetAuthUser(), tenant)

    return item != nil && item.IsPersisted()

  }
  return true
}

func (this *BaseAuthController) SetAuthTenantSession(tenant *models.Tenant) {

  if(this.HasTenantAuth(tenant)){
    this.baseController.Log("Set tenant session. user %v now is using tenant %v", this.GetAuthUser().Id, tenant.Id)
    this.baseController.SetSession("authtenantid", tenant.Id)
  } else {
    this.baseController.Log("Cannot set tenant session. user %v not enable to use tenant %v", this.GetAuthUser().Id, tenant.Id)
  }

}

func (this *BaseAuthController) GetAuthTenantSession() *models.Tenant {
  if id, ok := this.baseController.GetSession("authtenantid").(int64); ok {
    if id > 0 {
      tenant := models.Tenant{ Id: int64(id) }
      this.baseController.Session.Load(&tenant)
      return &tenant
    }
  }

  return nil
}

func (this *BaseAuthController) GetAuthTenant() *models.Tenant {

  tenant := this.GetAuthTenantSession()
  if tenant != nil && tenant.IsPersisted() {
    return tenant
  }

  return this.tenant
}

func (this *BaseAuthController) SetAuthTenant(t *models.Tenant ) {
  this.tenant = t
}

func (this *BaseAuthController) SetAuthUser(u *models.User)  {
  this.userinfo = u
}

func (this *BaseAuthController) GetAuthUser() *models.User {
  return this.userinfo
}
