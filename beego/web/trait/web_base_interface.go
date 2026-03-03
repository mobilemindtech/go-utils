package trait

import (
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/cache"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/support"
)

type WebConfigs struct {
	UseJsonPackage          bool
	JsonPackageAsCamelCase  bool
	CustomJsonEncoder       func(interface{}) ([]byte, error)
	NewJSON                 func() *json.JSON
	ExitWithHttpCode        bool
	NotLoadTenantsOnSession bool
	CustomAppAuthenticator  func(*models.App) (*models.User, error)
	ViewPath                string
}

func (this *WebConfigs) SetViewPath(path string) *WebConfigs {
	this.ViewPath = path
	return this
}

func (this *WebConfigs) SetUseJsonPackage(v bool) *WebConfigs {
	this.UseJsonPackage = v
	return this
}
func (this *WebConfigs) SetJsonPackageAsCamelCase(v bool) *WebConfigs {
	this.JsonPackageAsCamelCase = v
	return this
}
func (this *WebConfigs) SetCustomJsonEncoder(v func(interface{}) ([]byte, error)) *WebConfigs {
	this.CustomJsonEncoder = v
	return this
}
func (this *WebConfigs) SetCustomJsonCreator(v func() *json.JSON) *WebConfigs {
	this.NewJSON = v
	return this
}
func (this *WebConfigs) SetExitWithHttpCode(v bool) *WebConfigs {
	this.ExitWithHttpCode = v
	return this
}
func (this *WebConfigs) SetNotLoadTenantsOnSession(v bool) *WebConfigs {
	this.NotLoadTenantsOnSession = v
	return this
}
func (this *WebConfigs) SetCustomAppAuthenticator(v func(*models.App) (*models.User, error)) *WebConfigs {
	this.CustomAppAuthenticator = v
	return this
}

func (this *WebConfigs) IsLoadTenantsOnSession() bool {
	return !this.NotLoadTenantsOnSession
}

type WebBaseInterface interface {
	GetBeegoController() *beego.Controller
	GetAuthUser() *models.User
	GetAuthTenant() *models.Tenant
	GetSession() *db.Session
	IsJson() bool
	GetCacheService() *cache.CacheService

	GetModelTenant() *models.Tenant
	GetModelTenantUser() *models.TenantUser
	GetModelUser() *models.User

	GetLang() string
	GetMessage(key string, args ...interface{}) string

	GetCharacter() *support.Character

	GetData() map[interface{}]interface{}
	SetViewData(values ...interface{})

	GetFlash() *beego.FlashData
	FlashError(msg string, args ...interface{})
	FlashSuccess(msg string, args ...interface{})
	FlashWarn(msg string, args ...interface{})
	FlashNotice(msg string, args ...interface{})
	SetErrorState()
	FlashEnd(store bool)

	SetTemplateName(name string)

	RenderJsonWithValidationError()
	RenderJsonError(format string, v ...interface{})
	RenderJsonOk(format string, v ...interface{})
	RenderJson(result interface{})
	RenderJsonWithStatusCode(result interface{}, statusCode int)

	GetViewPath() string
	SetViewPath(path string)

	SetWebConfigs(configs *WebConfigs)
	GetWebConfigs() *WebConfigs

	GetInheritedController() interface{}
	GetDefaultLocation() *time.Location

	RenderResults(tplName string, results interface{}, totalCount ...int64)
	RenderResult(tplName string, result interface{})

	RenderJsonResults(results interface{}, totalCount ...int64)
	RenderJsonResult(result interface{})
	RenderJsonResultWithMessage(result interface{}, msg string)

	RenderEntity(tplName string, result interface{})
	RenderEntities(tplName string, result interface{}, totalCount ...int64)

	ReleaseCacheService()
	ReleaseDbSession()
	ReleaseDbSessionWithError()

	GetHeaderByName(name string) string
	GetHeaderByNames(names ...string) string

	// RenderJsonValidationError
	RenderJsonWithBadRequest(msg string)
	// RenderJsonForbidenError
	RenderJsonWithForbidden(msg string, abort bool)
	// RenderJsonUnauthorizedError
	RenderJsonWithUnauthorized(msg string, abort bool)
	//OnErrorAny
	RenderJsonOrRedirect(path string, msg string, args ...interface{})

	// OnRedirectError
	RedirectWithError(path string, msg string, args ...interface{})

	GetRawBody() []byte
	GetCtx() *context.Context

	GetQuery(key string) string
	GetParam(key string) string
}
