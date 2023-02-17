package controllers

import (
	_ "fmt"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-utils/app/models"
	"github.com/mobilemindtec/go-utils/app/route"
	"github.com/mobilemindtec/go-utils/app/services"
	"github.com/mobilemindtec/go-utils/cache"
)

type BaseAuthController struct {
	userinfo *models.User
	tenant   *models.Tenant

	IsLoggedIn      bool
	IsTokenLoggedIn bool

	Auth *services.AuthService

	baseController *BaseController
}

func (this *BaseAuthController) NestPrepareAuth(base *BaseController) {

	//base.Log("** app.controllers.BaseAuthController.NestPrepareAuth")

	this.baseController = base

	// login
	this.AppAuth()
	this.SetParams()

	this.IsLoggedIn = this.baseController.GetSession("userinfo") != nil
	this.IsTokenLoggedIn = this.baseController.GetSession("appuserinfo") != nil

	var tenant *models.Tenant
	tenantUuid := this.baseController.GetHeaderByNames("tenant", "X-Auth-Tenant")

	if len(tenantUuid) > 0 {
		//this.baseController.Log("tenantUuid = %v", tenantUuid)

		loader := func() (*models.Tenant, error) {
			ModelTenant := this.baseController.ModelTenant
			return ModelTenant.GetByUuidAndEnabled(tenantUuid)

		}

		tenant, _ = cache.Memoize(this.baseController.CacheService, tenantUuid, new(models.Tenant), loader)
		this.SetAuthTenant(tenant)
	}

	//fmt.Println("this.IsLoggedIn || this.IsTokenLoggedIn", this.IsLoggedIn , this.IsTokenLoggedIn)

	if this.IsLoggedIn || this.IsTokenLoggedIn {

		if this.IsLoggedIn {
			this.SetAuthUser(this.GetLogin())
		} else {
			this.SetAuthUser(this.GetTokenLogin())
		}

		if !this.IsTokenLoggedIn {

			tenant = this.GetAuthTenantSession()

			//this.baseController.Log("SESSION TENANR = ", tenant)

			if tenant == nil {
				ModelTenantUser := this.baseController.ModelTenantUser
				tenant, _ = ModelTenantUser.GetFirstTenant(this.GetAuthUser())
				//this.baseController.Log("FIRST TENANR = ", tenant)
			}

		}

		if tenant == nil || !tenant.IsPersisted() {

			logs.Error("** user does not has active tenant")

			if this.IsTokenLoggedIn && !this.baseController.IsJson() {
				this.baseController.OnJsonError("set header tenant")
			} else {
				this.baseController.OnErrorAny("/", "user does not has active tenant")
			}
			return
		}

		this.SetAuthTenant(tenant)

		this.baseController.Log("::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::")
		this.baseController.Log(":: Tenant Id = %v", this.GetAuthTenant().Id)
		this.baseController.Log(":: Tenant Name = %v", this.GetAuthTenant().Name)
		this.baseController.Log(":: User Id = %v", this.GetAuthUser().Id)
		this.baseController.Log(":: User Name = %v", this.GetAuthUser().Name)
		this.baseController.Log(":: User Authority = %v", this.GetAuthUser().Role.Authority)
		this.baseController.Log(":: User Roles = %v", this.GetAuthUser().GetAuthorities())
		this.baseController.Log(":: User IsLoggedIn = %v", this.IsLoggedIn)
		this.baseController.Log(":: User IsTokenLoggedIn = %v", this.IsTokenLoggedIn)
		this.baseController.Log(":: User Auth Token = %v", this.baseController.GetToken())
		this.baseController.Log("::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::")

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

func (this *BaseAuthController) AppAuth() {

	token := this.baseController.GetToken()

	if strings.TrimSpace(token) != "" {

		auth := services.NewLoginService(this.baseController.Lang, this.baseController.Session)

		logs.Debug("Authenticate by token %v", token)

		loader := func() (*models.User, error) {
			return auth.AuthenticateToken(token)
		}

		user, err := cache.Memoize(this.baseController.CacheService, token, new(models.User), loader)

		if err != nil {
			logs.Error("LOGIN ERROR: %v", err)
			this.LogOut()
			return
		}

		if user == nil {
			logs.Error("LOGIN ERROR: user not found!")
			this.LogOut()
			return
		}

		this.SetTokenLogin(user)
	}
}

func (this *BaseAuthController) GetLogin() *models.User {
	id := this.baseController.GetSession("userinfo").(int64)
	return this.memoizeUser(id)
}

func (this *BaseAuthController) GetTokenLogin() *models.User {
	id := this.baseController.GetSession("appuserinfo").(int64)
	return this.memoizeUser(id)
}

func (this *BaseAuthController) SessionLogOut() {
	this.LogOut()
}

func (this *BaseAuthController) LogOut() {

	this.baseController.BaseController.LogoutHanlder()
	this.baseController.DelSession("userinfo")
	this.baseController.DelSession("appuserinfo")
	this.baseController.DelSession("authtenantid")
	this.baseController.DestroySession()
}

func (this *BaseAuthController) SetLogin(user *models.User) {
	this.baseController.SetSession("userinfo", user.Id)
}

func (this *BaseAuthController) SetTokenLogin(user *models.User) {
	this.baseController.SetSession("appuserinfo", user.Id)
}

func (this *BaseAuthController) LoginPath() string {
	return this.baseController.URLFor("LoginController.Login")
}

func (this *BaseAuthController) SetParams() {
	this.baseController.Data["Params"] = make(map[string]string)

	values, _ := this.baseController.Input()

	for k, v := range values {
		this.baseController.Data["Params"].(map[string]string)[k] = v[0]
	}
}

func (this *BaseAuthController) OnLoginRedirect() {
	path := this.baseController.Ctx.Input.URI()
	if !strings.Contains("?", path) {
		path = "?redirect=" + path
	}
	this.baseController.Ctx.Redirect(302, this.LoginPath()+path)
}

func (this *BaseAuthController) AuthCheck() {
	if !this.IsLoggedIn && !this.IsTokenLoggedIn {
		if this.baseController.IsJson() {

			this.baseController.OnJsonError(this.baseController.GetMessage("security.notLoggedIn"))
			this.baseController.Abort("401")

		} else {
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

	//this.baseController.Log("** UpSecurityAuth")

	roles := []string{}

	if this.Auth != nil {
		roles = this.Auth.GetUserRoles()
	}

	if !route.IsRouteAuthorized(this.baseController.Ctx, roles) {

		logs.Debug("** path %v not authorized ", this.baseController.Ctx.Input.URL())

		if !this.IsLoggedIn && !this.IsTokenLoggedIn {
			if this.baseController.IsJson() {
				this.baseController.OnJsonError(this.baseController.GetMessage("security.notLoggedIn"))
				//this.baseController.Abort("401")
			} else {
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

func (this *BaseAuthController) HasTenantAuth(tenant *models.Tenant) bool {
	if !this.Auth.IsRoot() {

		ModelTenantUser := this.baseController.ModelTenantUser

		loader := func() bool {
			item, _ := ModelTenantUser.FindByUserAndTenant(this.GetAuthUser(), tenant)
			return item != nil && item.IsPersisted()
		}
		parser := func(v string) bool {
			return v == "true"
		}
		cacheKey := cache.CacheKey("has_user", this.GetAuthUser().Id, "tenant", tenant.Id)
		return cache.MemoizeVal(this.baseController.CacheService, cacheKey, parser, loader)

	}
	return true
}

func (this *BaseAuthController) SetAuthTenantSession(tenant *models.Tenant) {

	if this.HasTenantAuth(tenant) {
		logs.Debug("Set tenant session. user %v now is using tenant %v", this.GetAuthUser().Id, tenant.Id)
		this.baseController.SetSession("authtenantid", tenant.Id)
	} else {
		logs.Debug("Cannot set tenant session. user %v not enable to use tenant %v", this.GetAuthUser().Id, tenant.Id)
	}

}

func (this *BaseAuthController) GetAuthTenantSession() *models.Tenant {

	var tenant *models.Tenant

	if id, ok := this.baseController.GetSession("authtenantid").(int64); ok && id > 0 {
		loader := func() (*models.Tenant, error) {
			tenant := models.Tenant{Id: int64(id)}
			this.baseController.Session.Load(&tenant)
			return &tenant, nil
		}
		tenant, _ = cache.Memoize(this.baseController.CacheService, cache.CacheKey("tenant_", id), new(models.Tenant), loader)
	}

	return tenant
}

func (this *BaseAuthController) GetAuthTenant() *models.Tenant {
	tenant := this.GetAuthTenantSession()
	if tenant != nil && tenant.IsPersisted() {
		return tenant
	}
	return this.tenant
}

func (this *BaseAuthController) SetAuthTenant(t *models.Tenant) {
	this.tenant = t
}

func (this *BaseAuthController) SetAuthUser(u *models.User) {
	this.userinfo = u
}

func (this *BaseAuthController) GetAuthUser() *models.User {
	return this.userinfo
}

func (this *BaseAuthController) memoizeUser(id int64) *models.User {
	var user *models.User
	loader := func() (*models.User, error) {
		e, err := this.baseController.Session.FindById(new(models.User), id)
		if err != nil {
			return nil, err
		}
		user := e.(*models.User)
		this.baseController.ModelUser.LoadRelated(user)
		return user, nil
	}
	user, _ = cache.Memoize(this.baseController.CacheService, cache.CacheKey("user_", id), new(models.User), loader)
	return user
}
