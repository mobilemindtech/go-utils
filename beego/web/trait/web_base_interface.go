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

	GetViewPath() string
	SetViewPath(path string)

	SetWebConfigs(configs *WebConfigs)
	GetWebConfigs() *WebConfigs

	GetInheritedController() interface{}
	GetDefaultLocation() *time.Location

	RenderJsonWithValidationError()
	RenderJsonError(format string, v ...interface{})
	RenderJsonOk(format string, v ...interface{})
	RenderJson(result interface{})
	RenderJsonWithStatusCode(result interface{}, statusCode int)

	RenderJsonWithResults(results interface{}, totalCount ...int64)
	RenderJsonWithResult(result interface{})
	RenderJsonResultWithMessage(result interface{}, msg string)

	RenderTemplateWithResults(tplName string, results interface{}, totalCount ...int64)
	RenderTemplateWithResult(tplName string, result interface{})
	RenderTemplateWithEntity(tplName string, result interface{})
	RenderTemplateWithEntities(tplName string, result interface{}, totalCount ...int64)

	RenderTemplate(tplName string, data ...interface{})

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

	GetCurrentTimeUnix() int64
	GetCurrentTime() time.Time
}
