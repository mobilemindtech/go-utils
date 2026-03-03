package features

import (
	"fmt"
	"strings"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-io/option"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/app/route"
	"github.com/mobilemindtech/go-utils/app/services"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/cache"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	"github.com/mobilemindtech/go-utils/v2/inline"
	"github.com/mobilemindtech/go-utils/v2/lists"
	"github.com/mobilemindtech/go-utils/v2/maps"
)

type WebAuth struct {
	IsWebLoggedIn       bool
	IsTokenLoggedIn     bool
	IsCustomAppLoggedIn bool
	Auth                *services.AuthService

	cacheKeysDeleteOnLogOut []string
	userinfo                *models.User
	tenant                  *models.Tenant
	base                    trait.WebBaseInterface
}

func (this *WebAuth) InitWebAuth(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebAuth) SetUpAuth() {
	this.Auth = services.NewAuthService(this.base.GetSession())
	this.AuthPrepare()
}

func (this *WebAuth) tryAppAuthenticate(token string) (*models.User, error) {

	customAppAuthFn := this.base.GetWebConfigs().CustomAppAuthenticator

	if customAppAuthFn == nil {
		return nil, nil
	}

	first := db.RunWithIgnoreTenantFilter(
		this.base.GetSession(),
		func(s *db.Session) *result.Result[*option.Option[*models.App]] {
			return criteria.
				New[*models.App](s).
				Eq("Token", token).
				Eager("Tenant").
				GetFirst()
		})

	if first.IsError() {
		return nil, first.Failure()
	}

	if first.Get().IsSome() {
		app := first.Get().Get()
		user, err := customAppAuthFn(app)

		if err != nil || user == nil || !user.IsPersisted() {
			return nil, err
		}

		this.SetAuthTenant(app.Tenant)
		this.SetCustomAppLogin(user)
		this.SetCustomAppName(fmt.Sprintf("%v - %v", app.Id, app.Name))
		this.SetCustomAppId(app.Id)
		return user, nil
	}

	return nil, nil
}

func (this *WebAuth) AuthPrepare() {
	// login

	this.AppAuth()

	this.SetParams()

	this.IsWebLoggedIn = this.base.GetBeegoController().GetSession("userinfo") != nil
	this.IsTokenLoggedIn = this.base.GetBeegoController().GetSession("appuserinfo") != nil
	this.IsCustomAppLoggedIn = this.base.GetBeegoController().GetSession("customappuserinfo") != nil

	var tenant *models.Tenant

	tenantUuid := this.GetHeaderTenant()

	if len(tenantUuid) > 0 {
		loader := func() (*models.Tenant, error) {
			return this.base.GetModelTenant().GetByUuidAndEnabled(tenantUuid)

		}
		tenant, _ = cache.Memoize(this.base.GetCacheService(), tenantUuid, new(models.Tenant), loader)
		this.SetAuthTenant(tenant)
	}

	if this.IsLoggedIn() {

		if this.IsWebLoggedIn {
			// web login
			this.SetAuthUser(this.GetLogin())
		} else if this.IsTokenLoggedIn {
			// token login
			this.SetAuthUser(this.GetTokenLogin())
		} else if this.IsCustomAppLoggedIn {
			// custom app login
			this.SetAuthUser(this.GetCustomAppLogin())
		}

		if this.IsWebLoggedIn {
			tenant = this.GetAuthTenantSession()
			if tenant == nil {
				tenant, _ = this.base.GetModelTenantUser().GetFirstTenant(this.GetAuthUser())
			}
		}

		if tenant == nil || !tenant.IsPersisted() {
			tenant = this.GetAuthUser().Tenant
			this.base.GetSession().Load(tenant)
		}

		if tenant == nil || !tenant.IsPersisted() {

			logs.Error("ERROR: user does not have active tenant")

			if this.IsTokenLoggedIn || this.base.IsJson() {
				this.base.RenderJsonWithBadRequest("tenant not configured")
			} else {
				this.base.RenderJsonOrRedirect("/", "user does not has active tenant")
			}
			return
		}

		if !tenant.Enabled && !services.IsRootUser(this.GetAuthUser()) {
			logs.Error("ERROR: tenant ", tenant.Id, " - ", tenant.Name, " is disabled")

			if this.IsTokenLoggedIn || this.base.IsJson() {
				this.base.RenderJsonWithForbidden("operation not permitted", false)
			} else {
				this.LogOut()
				this.base.RenderJsonOrRedirect("/", "operation not permitted")
			}
			return
		}

		this.SetAuthTenant(tenant)

		//logs.Trace("::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::")
		logs.Trace("::> Tenant = %v - %v", this.GetAuthTenant().Id, this.GetAuthTenant().Name)
		logs.Trace("::> User = %v - %v", this.GetAuthUser().Id, this.GetAuthUser().Name)
		logs.Trace("::> User Role = %v", this.GetAuthUser().Role.Authority)
		logs.Trace("::> Custom App = %v", inline.If(len(this.GetCustomAppName()) == 0, "-", this.GetCustomAppName()))
		//logs.Trace("::> Login by Web? = %v, Token? = %v, CustomApp? = %v", this.IsWebLoggedIn, this.IsTokenLoggedIn, this.IsCustomAppLoggedIn)
		//logs.Trace(":: Auth Token = %v", this.GetHeaderToken())
		//logs.Trace("::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::::")

		this.base.GetBeegoController().Data["UserInfo"] = this.GetAuthUser()
		this.base.GetBeegoController().Data["Tenant"] = this.GetAuthTenant()

		//ioc.Get[services.AuthService](this.Container)
		this.Auth.SetUserInfo(this.GetAuthUser())
	}

	this.base.GetBeegoController().Data["IsLoggedIn"] = this.IsLoggedIn()

	if this.IsLoggedIn() {
		this.base.GetBeegoController().Data["IsAdmin"] = this.Auth.IsAdmin()
		this.base.GetBeegoController().Data["IsRoot"] = this.Auth.IsRoot()
	}

	this.UpSecurityAuth()
}

func (this *WebAuth) IsLoggedIn() bool {
	return this.IsWebLoggedIn || this.IsTokenLoggedIn || this.IsCustomAppLoggedIn
}

func (this *WebAuth) IsWebOrTokenLoggerIn() bool {
	return this.IsWebLoggedIn || this.IsTokenLoggedIn
}

func (this *WebAuth) GetHeaderToken() string {
	token := this.base.GetHeaderByName("X-Auth-Token")
	if len(token) == 0 {
		token = this.base.GetHeaderByName("Authorization")
	}
	return token
}

func (this *WebAuth) IsBearerToken() bool {
	return this.Auth.IsBearerToken(this.GetHeaderToken())
}

func (this *WebAuth) GetHeaderTenant() string {
	return this.base.GetHeaderByNames("tenant", "X-Auth-Tenant")
}

func (this *WebAuth) LoadUserTenants() {

	if this.IsWebLoggedIn {

		if this.base.GetWebConfigs().IsLoadTenantsOnSession() {

			cacheKey := cache.CacheKey("tenants_user_", this.GetAuthUser().Id)
			this.DeleteCacheOnLogout(cacheKey)

			loader := func() ([]*models.Tenant, error) {

				var tenants []*models.Tenant
				if this.Auth.IsRoot() {
					its, _ := this.base.GetModelTenant().List()
					tenants = *its
				} else {
					list, _ := this.base.GetModelTenantUser().ListByUserAdmin(this.GetAuthUser())

					for _, it := range *list {

						if !it.Enabled {
							continue
						}

						this.base.GetSession().Load(it.Tenant)
						tenants = append(tenants, it.Tenant)
					}
				}

				return tenants, nil
			}

			authorizeds, _ := cache.Memoize(this.base.GetCacheService(), cacheKey, new([]*models.Tenant), loader)

			this.base.GetSession().SetAuthorizedTenants(lists.MapToInterface(authorizeds))
			this.base.GetBeegoController().Data["AvailableTenants"] = authorizeds
		} else {
			logs.Debug("do not load user tenants on session")
		}
	} else {
		this.base.GetBeegoController().Data["AvailableTenants"] = []*models.Tenant{}
	}
}

func (this *WebAuth) AppAuth() {

	token := this.GetHeaderToken()

	if strings.TrimSpace(token) != "" {

		auth := services.NewLoginService(this.base.GetLang(), this.base.GetSession())

		//logs.Debug("authenticating with token: %v", token)

		if this.Auth.IsBearerToken(token) {

			user, err := this.Auth.CheckBearerToken(token)

			if err != nil {
				logs.Error("bearer token login error: %v", err)
				this.LogOut()
				return
			}

			if user == nil {
				logs.Error("bearer token login error: user not found")
				this.LogOut()
				return
			}

			if !user.Enabled {
				logs.Error("bearer token login error: user disabled")
				this.LogOut()
				return
			}

			this.SetTokenLogin(user)
			return

		} else {

			loader := func() (*models.User, error) {
				return auth.AuthenticateToken(token)
			}

			user, err := cache.Memoize(this.base.GetCacheService(), token, new(models.User), loader)

			if err != nil {

				if _, ok := err.(*services.LoginErrorUserNotFound); ok {
					logs.Debug("tryAppAuthenticate")
					user, err = this.tryAppAuthenticate(token)
					if err != nil {
						logs.Error("CUSTOM APP LOGIN ERROR: %v", err)
						this.LogOut()
						return
					}
				} else {
					logs.Error("LOGIN ERROR: %v", err)
					this.LogOut()
					return
				}
			}

			if user == nil || !user.IsPersisted() {
				logs.Error("LOGIN ERROR: user not found!")
				this.LogOut()
				return
			}

			this.SetTokenLogin(user)
			return
		}
	}
}

func (this *WebAuth) GetLogin() *models.User {
	id, _ := this.base.GetBeegoController().GetSession("userinfo").(int64)
	return this.memoizeUser(id)
}

func (this *WebAuth) GetTokenLogin() *models.User {
	id, _ := this.base.GetBeegoController().GetSession("appuserinfo").(int64)
	return this.memoizeUser(id)
}

func (this *WebAuth) GetCustomAppLogin() *models.User {
	id, _ := this.base.GetBeegoController().GetSession("customappuserinfo").(int64)
	return this.memoizeUser(id)
}

func (this *WebAuth) GetCustomAppName() string {
	name, _ := this.base.GetBeegoController().GetSession("customappname").(string)
	return name
}

func (this *WebAuth) GetCustomAppId() int64 {
	id, _ := this.base.GetBeegoController().GetSession("customappid").(int64)
	return id
}

func (this *WebAuth) GetCustomApp() (*models.App, error) {
	id := this.GetCustomAppId()
	return criteria.New[*models.App](this.base.GetSession()).FindById(id)
}

func (this *WebAuth) SessionLogOut() {
	this.LogOut()
}

func (this *WebAuth) LogOut() {
	this.base.GetCacheService().Delete(this.cacheKeysDeleteOnLogOut...)
	bee := this.base.GetBeegoController()
	bee.DelSession("userinfo")
	bee.DelSession("appuserinfo")
	bee.DelSession("customappuserinfo")
	bee.DelSession("authtenantid")
	bee.DelSession("customappname")
	bee.DelSession("customappid")
	bee.DestroySession()
}

func (this *WebAuth) SetCustomAppName(appname string) {
	this.base.GetBeegoController().SetSession("customappname", appname)
}

func (this *WebAuth) SetCustomAppId(id int64) {
	this.base.GetBeegoController().SetSession("customappid", id)
}
func (this *WebAuth) SetLogin(user *models.User) {
	this.SetAuthUserSession(user)
}

func (this *WebAuth) SetAuthUserSession(user *models.User) {
	this.base.GetBeegoController().SetSession("userinfo", user.Id)
}

func (this *WebAuth) SetTokenLogin(user *models.User) {
	this.base.GetBeegoController().SetSession("appuserinfo", user.Id)
}

// set login from models.App (custom app login)
func (this *WebAuth) SetCustomAppLogin(user *models.User) {
	this.base.GetBeegoController().SetSession("customappuserinfo", user.Id)
}

func (this *WebAuth) LoginPath() string {
	return this.base.GetBeegoController().URLFor("LoginController.Login")
}

func (this *WebAuth) OnLoginRedirect() {
	path := this.base.GetBeegoController().Ctx.Input.URI()
	if !strings.Contains("?", path) {
		path = "?redirect=" + path
	}
	this.base.GetBeegoController().Ctx.Redirect(302, this.LoginPath()+path)
}

func (this *WebAuth) AuthCheck() {
	if !this.IsLoggedIn() {
		if this.base.IsJson() {
			msgKey := "security.notLoggedIn"
			if this.base.GetWebConfigs().ExitWithHttpCode {
				msgKey = "security.unauthorized"
			}
			this.base.RenderJsonWithUnauthorized(this.base.GetMessage(msgKey), true)
		} else {
			this.OnLoginRedirect()
		}
	}
}

func (this *WebAuth) AuthCheckRoot() {
	if !this.IsWebLoggedIn {
		if this.base.IsJson() {
			msgKey := "security.notLoggedIn"
			if this.base.GetWebConfigs().ExitWithHttpCode {
				msgKey = "security.unauthorized"
			}
			this.base.RenderJsonWithUnauthorized(this.base.GetMessage(msgKey), true)
		} else {
			this.OnLoginRedirect()
		}
	}

	if !this.Auth.IsRoot() {
		if this.base.IsJson() {
			this.base.RenderJsonWithForbidden(this.base.GetMessage("security.rootRequired"), true)
		} else {
			this.base.GetBeegoController().Redirect("/", 302)
		}
	}
}

func (this *WebAuth) AuthCheckAdmin() {
	if !this.IsWebLoggedIn {
		if this.base.IsJson() {
			msgKey := "security.notLoggedIn"
			if this.base.GetWebConfigs().ExitWithHttpCode {
				msgKey = "security.unauthorized"
			}
			this.base.RenderJsonWithUnauthorized(this.base.GetMessage(msgKey), true)
		} else {
			this.OnLoginRedirect()
			this.base.RedirectWithError("/", this.base.GetMessage("security.rootRequired"))
		}
	}

	if !this.Auth.IsRoot() && !this.Auth.IsAdmin() {
		if this.base.IsJson() {
			this.base.RenderJsonWithForbidden(this.base.GetMessage("security.rootRequired"), true)
		} else {
			this.base.RedirectWithError("/", this.base.GetMessage("security.rootRequired"))
		}
	}
}

func (this *WebAuth) UpSecurityAuth() bool {

	roles := []string{}

	if this.Auth.IsAuthenticated() {
		roles = this.Auth.GetUserRoles()
	}

	if !route.IsRouteAuthorized(this.base.GetBeegoController().Ctx, roles) {

		logs.Warn("WARN: path %v not authorized ", this.base.GetBeegoController().Ctx.Input.URL())

		if !this.IsLoggedIn() {

			if this.IsBearerToken() {
				this.base.RenderJsonWithStatusCode(
					maps.JSON("message", "unauthorized"), 401)
			} else {
				if this.base.IsJson() {
					msgKey := "security.notLoggedIn"
					if this.base.GetWebConfigs().ExitWithHttpCode {
						msgKey = "security.unauthorized"
					}
					this.base.RenderJsonWithUnauthorized(
						this.base.GetMessage(msgKey), false)
				} else {
					this.OnLoginRedirect()
				}
			}
			return false
		}

		if this.IsBearerToken() {
			this.base.RenderJsonWithStatusCode(
				maps.JSON("message", "forbidden"), 403)
		} else {
			if this.base.IsJson() {
				this.base.RenderJsonWithUnauthorized(this.base.GetMessage("security.denied"), false)
			} else {
				this.base.RedirectWithError("/", this.base.GetMessage("security.denied"))
			}
		}

		return false

	}

	return true
}

func (this *WebAuth) HasTenantAuth(tenant *models.Tenant) bool {
	if !this.Auth.IsRoot() {

		ModelTenantUser := this.base.GetModelTenantUser()

		loader := func() bool {
			item, _ := ModelTenantUser.FindByUserAndTenant(this.GetAuthUser(), tenant)
			return item != nil && item.IsPersisted()
		}
		parser := func(v string) bool {
			return v == "true"
		}
		cacheKey := cache.CacheKey("has_user", this.GetAuthUser().Id, "tenant", tenant.Id)
		return cache.MemoizeVal(this.base.GetCacheService(), cacheKey, parser, loader)

	}
	return true
}

func (this *WebAuth) SetAuthTenantSession(tenant *models.Tenant) {
	if this.HasTenantAuth(tenant) {
		logs.Info("Set tenant session. user %v now is using tenant %v", this.GetAuthUser().Id, tenant.Id)
		this.base.GetBeegoController().SetSession("authtenantid", tenant.Id)
	} else {
		logs.Error("Cannot set tenant session. user %v not enable to use tenant %v", this.GetAuthUser().Id, tenant.Id)
	}

}

func (this *WebAuth) GetAuthTenantSession() *models.Tenant {
	var tenant *models.Tenant

	if id, ok := this.base.GetBeegoController().GetSession("authtenantid").(int64); ok && id > 0 {
		loader := func() (*models.Tenant, error) {
			tenant := models.Tenant{Id: int64(id)}
			this.base.GetSession().Load(&tenant)
			return &tenant, nil
		}
		tenant, _ = cache.Memoize(this.base.GetCacheService(), cache.CacheKey("tenant_", id), new(models.Tenant), loader)
	}

	return tenant
}

func (this *WebAuth) GetAuthTenant() *models.Tenant {
	tenant := this.GetAuthTenantSession()
	if tenant != nil && tenant.IsPersisted() {
		return tenant
	}
	return this.tenant
}

func (this *WebAuth) GetAuthTenantString() string {
	tenant := this.GetAuthTenantSession()
	if tenant != nil && tenant.IsPersisted() {

		return fmt.Sprintf("%v - %v", tenant.Id, tenant.Name)

	}
	return "[has not auth tenant]"
}

func (this *WebAuth) SetAuthTenant(t *models.Tenant) {
	this.tenant = t
}

func (this *WebAuth) SetAuthUser(u *models.User) {
	this.userinfo = u
}

func (this *WebAuth) GetAuthUser() *models.User {
	return this.userinfo
}

func (this *WebAuth) DeleteCacheOnLogout(keys ...string) {
	this.cacheKeysDeleteOnLogOut = append(this.cacheKeysDeleteOnLogOut, keys...)
}

func (this *WebAuth) memoizeUser(id int64) *models.User {
	var user *models.User
	loader := func() (*models.User, error) {
		e, err := this.base.GetSession().FindById(new(models.User), id)
		if err != nil {
			return nil, err
		}
		user := e.(*models.User)
		this.base.GetModelUser().LoadRelated(user)
		return user, nil
	}
	user, _ = cache.Memoize(this.base.GetCacheService(), cache.CacheKey("user_", id), new(models.User), loader)
	return user
}

func (this *WebAuth) SetParams() {
	this.base.GetBeegoController().Data["Params"] = make(map[string]string)

	values, _ := this.base.GetBeegoController().Input()

	for k, v := range values {
		this.base.GetBeegoController().Data["Params"].(map[string]string)[k] = v[0]
	}
}
