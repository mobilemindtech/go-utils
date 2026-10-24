package v2

import (
	"fmt"
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/validator"
	"github.com/mobilemindtech/go-utils/beego/web/features"
	"github.com/mobilemindtech/go-utils/beego/web/features/v2"
	"github.com/mobilemindtech/go-utils/beego/web/resps"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/cache"
	"github.com/mobilemindtech/go-utils/support"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	"github.com/mobilemindtech/go-utils/v2/maps"
)

type WebBase struct {
	features.WebAudit
	features.WebAuth
	features.WebDbSession
	v2.WebJsonRenderV2
	features.WebMessages
	features.WebMisc
	features.WebModels
	features.WebPagination
	v2.WebParser
	v2.WebRespUtil
	v2.WebTemplateRender
	features.WebValidation
	CacheService          *cache.CacheService
	Character             *support.Character
	BeegoController       *beego.Controller
	InheritedController   interface{}
	Configs               *trait.WebConfigs
	entityValidatorConfig *validator.EntityValidatorConfig
	DefaultLocation       *time.Location
}

func (this *WebBase) InitWebBase(ctrl *beego.Controller) {
	this.BeegoController = ctrl

	this.DefaultLocation, _ = time.LoadLocation("America/Sao_Paulo")

	this.InitWebAudit(this)
	this.InitWebAuth(this)
	this.InitWebDbSession(this)
	this.InitWebJsonRenderV2(this)
	this.InitWebMessages(this)
	this.InitWebMisc(this)
	this.InitWebModels(this)
	this.InitWebPagination(this)
	this.InitWebParser(this)
	this.IniWebRespUtil(this)
	this.InitWebTemplateRender(this)
	this.InitWebValidation(this)

	this.CacheService = cache.New()
	this.Character = support.NewCharacter()
	// pode já ter sido configurado no Prepare do controller antes do InitWebBase
	// (ex.: SetCustomAppAuthenticator), não sobrescrever
	this.GetWebConfigs()

	// lang exists only after InitWebMessages
	this.entityValidatorConfig = &validator.EntityValidatorConfig{Lang: this.Lang, ViewPath: ctrl.ViewPath}
	this.EntityValidator = validator.NewEntityValidatorWithConfig(this.entityValidatorConfig)

	this.SetDefaultValuesInDataModel()

	// create database session
	this.WebControllerCreateSession()

	// load models
	this.WebControllerLoadModels()

	// setup auth rules
	this.SetUpAuth()

	// set current tenant on session
	this.Session.Tenant = this.GetAuthTenant()

	// load session tenants for authenticated user
	this.LoadUserTenants()

}

func (this *WebBase) GetBeegoController() *beego.Controller {
	return this.BeegoController
}

func (this *WebBase) GetAuthUser() *models.User {
	return this.WebAuth.GetAuthUser()
}

func (this *WebBase) GetAuthTenant() *models.Tenant {
	return this.WebAuth.GetAuthTenant()
}

func (this *WebBase) GetSession() *db.Session {
	return this.WebDbSession.Session
}

func (this *WebBase) IsJson() bool {
	return this.WebMisc.IsJson()
}

func (this *WebBase) GetCacheService() *cache.CacheService {
	return this.CacheService
}

func (this *WebBase) GetModelTenant() *models.Tenant {
	return this.WebModels.ModelTenant
}
func (this *WebBase) GetModelTenantUser() *models.TenantUser {
	return this.WebModels.ModelTenantUser
}

func (this *WebBase) GetModelUser() *models.User {
	return this.WebModels.ModelUser
}

func (this *WebBase) GetLang() string {
	return this.Lang
}

func (this *WebBase) GetMessage(key string, args ...interface{}) string {
	return this.WebMessages.GetMessage(key, args...)
}

func (this *WebBase) GetCharacter() *support.Character {
	return this.Character
}

func (this *WebBase) GetData() map[interface{}]interface{} {
	return this.BeegoController.Data
}

func (this *WebBase) SetViewData(values ...interface{}) {
	this.WebMisc.SetViewData(values...)
}

func (this *WebBase) GetFlash() *beego.FlashData {
	return this.Flash
}

func (this *WebBase) FlashError(msg string, args ...interface{}) {
	this.WebMessages.FlashError(msg, args...)
}

func (this *WebBase) FlashSuccess(msg string, args ...interface{}) {
	this.WebMessages.FlashSuccess(msg, args...)
}

func (this *WebBase) FlashWarn(msg string, args ...interface{}) {
	this.WebMessages.FlashWarn(msg, args...)
}
func (this *WebBase) FlashNotice(msg string, args ...interface{}) {
	this.WebMessages.FlashNotice(msg, args...)
}

func (this *WebBase) SetErrorState() {
	this.RollbackDbSession()
}

func (this *WebBase) FlashEnd(store bool) {
	this.WebMessages.FlashEnd(store)
}

func (this *WebBase) SetTemplateName(name string) {
	this.BeegoController.TplName = name
}

func (this *WebBase) RenderJsonWithValidationError() {
	this.WebJsonRenderV2.RenderJsonWithValidationError()
}

func (this *WebBase) RenderJsonError(format string, v ...interface{}) {
	this.RenderJson(resps.NewBadRequest(fmt.Sprintf(format, v...)))
}

func (this *WebBase) RenderJsonOk(format string, v ...interface{}) {
	this.RenderJson(resps.NewOk(maps.JSON("message", fmt.Sprintf(format, v...))))
}

func (this *WebBase) RenderJsonServerError(format string, v ...interface{}) {
	this.RenderJson(resps.NewServerError(fmt.Sprintf(format, v...)))
}

func (this *WebBase) RenderJson(result interface{}) {
	this.WebJsonRenderV2.RenderJson(result)
}

func (this *WebBase) RenderJsonWithStatusCode(body interface{}, statusCode int) {
	this.WebJsonRenderV2.RenderJson(resps.NewResponse(statusCode, body))
}

func (this *WebBase) GetViewPath() string {
	return this.Configs.ViewPath
}

func (this *WebBase) SetViewPath(path string) {
	this.Configs.SetViewPath(path)
	this.entityValidatorConfig.ViewPath = path
}

func (this *WebBase) SetWebConfigs(configs *trait.WebConfigs) {
	this.Configs = configs
}

func (this *WebBase) GetWebConfigs() *trait.WebConfigs {
	if this.Configs == nil {
		this.Configs = new(trait.WebConfigs)
	}
	return this.Configs
}

func (this *WebBase) GetInheritedController() interface{} {
	return this.InheritedController
}

func (this *WebBase) GetDefaultLocation() *time.Location {
	return this.DefaultLocation
}

// RenderJsonWithResults RenderJsonResults RenderResults Deprecated
func (this *WebBase) RenderJsonWithResults(results interface{}, totalCount ...int64) {
	if len(totalCount) > 0 {
		this.RenderJson(&criteria.Page{Data: results, TotalCount: int(totalCount[0])})
	} else {
		this.RenderJson(results)
	}
}

// RenderJsonWithResult RenderJsonResult RenderResults Deprecated
func (this *WebBase) RenderJsonWithResult(result interface{}) {
	this.RenderJson(result)
}

// RenderJsonResultWithMessage RenderResults Deprecated
func (this *WebBase) RenderJsonResultWithMessage(result interface{}, msg string) {
	this.RenderJson(
		resps.NewResponse(200, maps.JSON("message", msg, "data", result)))

}

// RenderResults Deprecated
func (this *WebBase) RenderTemplateWithResults(tplName string, results interface{}, totalCount ...int64) {
	if len(totalCount) > 0 {
		this.SetViewData("totalCount", totalCount[0])
	}
	this.RenderTemplate(tplName, "results", results)
}

// RenderTemplateWithResult RenderResults Deprecated
func (this *WebBase) RenderTemplateWithResult(tplName string, result interface{}) {
	this.RenderTemplate(tplName, "result", result)

}

// RenderTemplateWithEntity RenderEntity RenderResults Deprecated
func (this *WebBase) RenderTemplateWithEntity(tplName string, result interface{}) {
	this.RenderTemplate(tplName, "entity", result)
}

// RenderTemplateWithEntities RenderEntities RenderResults Deprecated
func (this *WebBase) RenderTemplateWithEntities(tplName string, result interface{}, totalCount ...int64) {
	if len(totalCount) > 0 {
		this.SetViewData("totalCount", totalCount[0])
	}
	this.RenderTemplate(tplName, "entities", result)
}

func (this *WebBase) ReleaseCacheService() {
	this.CacheService.Close()
	this.CacheService = nil
}

func (this *WebBase) ReleaseDbSession() {
	this.CloseDbSession()
}

func (this *WebBase) ReleaseDbSessionWithError() {
	this.CloseDbSessionWithError()
}

func (this *WebBase) GetHeaderByName(name string) string {
	return this.WebMisc.GetHeaderByName(name)
}

func (this *WebBase) GetHeaderByNames(names ...string) string {
	return this.WebMisc.GetHeaderByNames(names...)
}

// RenderJsonWithBadRequest RenderResults Deprecated
func (this *WebBase) RenderJsonWithBadRequest(msg string) {
	this.RenderJson(resps.NewBadRequest(msg))
}

// RenderJsonWithForbidden RenderResults Deprecated
func (this *WebBase) RenderJsonWithForbidden(msg string, abort bool) {
	this.RenderJson(resps.NewForbidden(msg))
	if abort {
		this.BeegoController.Abort("403")
	}
}

// RenderJsonUnauthorizedError
func (this *WebBase) RenderJsonWithUnauthorized(msg string, abort bool) {
	this.RenderJson(resps.NewForbidden(msg))
	if abort {
		this.BeegoController.Abort("401")
	}
}

// RenderJsonOrRedirect RenderResults Deprecated
func (this *WebBase) RenderJsonOrRedirect(path string, msg string, args ...interface{}) {
	if this.IsJson() {
		this.RenderJsonWithBadRequest(msg)
	} else {
		this.RedirectWithError(path, msg, args...)
	}
}

func (this *WebBase) RenderTemplate(tplName string, data ...interface{}) {
	this.WebTemplateRender.RenderTemplate(tplName, data...)
}

// OnRedirectError
func (this *WebBase) RedirectWithError(path string, msg string, args ...interface{}) {
	this.WebRespUtil.RedirectWithError(path, msg, args...)
}

func (this *WebBase) GetRawBody() []byte {
	return this.WebMisc.GetRawBody()
}

func (this *WebBase) GetCtx() *context.Context {
	return this.BeegoController.Ctx
}

func (this *WebBase) GetQuery(key string) string {
	return this.BeegoController.Ctx.Input.Query(key)
}

// GetParam Get url param, use :key
func (this *WebBase) GetParam(key string) string {
	return this.BeegoController.Ctx.Input.Param(key)
}
