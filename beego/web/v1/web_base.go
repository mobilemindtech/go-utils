package v1

import (
	"time"

	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/beego/v2/server/web/context"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/validator"
	"github.com/mobilemindtech/go-utils/beego/web/features"
	"github.com/mobilemindtech/go-utils/beego/web/features/v1"
	"github.com/mobilemindtech/go-utils/beego/web/features/v2"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/cache"
	"github.com/mobilemindtech/go-utils/support"
)

type WebBase struct {
	features.WebAudit
	features.WebAuth
	features.WebDbSession
	features.WebJson
	v1.WebJsonRenderV1
	v2.WebJsonRenderV2
	features.WebMessages
	features.WebMisc
	features.WebModels
	features.WebPagination
	v1.WebParser
	v1.WebRespUtil
	v1.WebTemplateRender
	features.WebUpload
	features.WebValidation
	CacheService          *cache.CacheService
	Character             *support.Character
	BeegoController       *beego.Controller
	InheritedController   interface{}
	Configs               *trait.WebConfigs
	entityValidatorConfig *validator.EntityValidatorConfig
}

func (this *WebBase) InitWebBase(ctrl *beego.Controller) {
	this.BeegoController = ctrl

	this.InitWebAudit(this)
	this.InitWebAuth(this)
	this.InitWebDbSession(this)
	this.InitWebJson(this)
	this.InitWebJsonRenderV1(this)
	this.InitWebJsonRenderV2(this)
	this.InitWebMessages(this)
	this.InitWebMisc(this)
	this.InitWebModels(this)
	this.InitWebPagination(this)
	this.InitWebParser(this)
	this.IniWebRespUtil(this)
	this.InitWebTemplateRender(this)
	this.InitWebUpload(this)
	this.InitWebValidation(this)

	this.CacheService = cache.New()
	this.Character = support.NewCharacter()
	this.Configs = new(trait.WebConfigs)

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
	this.OnJsonValidationError()
}

func (this *WebBase) RenderJsonError(format string, v ...interface{}) {
	this.OnJsonError(format, v...)
}

func (this *WebBase) RenderJsonOk(format string, v ...interface{}) {
	this.OnJsonOk(format, v...)
}

func (this *WebBase) RenderJson(result interface{}) {
	this.WebJsonRenderV2.RenderJson(result)
}

func (this *WebBase) RenderJsonWithStatusCode(result interface{}, statusCode int) {
	this.WebJsonRenderV1.RenderJsonWithStatusCode(result, statusCode)
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
	return this.Configs
}

func (this *WebBase) GetInheritedController() interface{} {
	return this.InheritedController
}

func (this *WebBase) GetDefaultLocation() *time.Location {
	return this.DefaultLocation
}

func (this *WebBase) RenderJsonWithResults(results interface{}, totalCount ...int64) {
	if len(totalCount) > 0 {
		this.OnJsonResultsWithTotalCount(results, totalCount[0])
	} else {
		this.OnJsonResults(results)
	}
}
func (this *WebBase) RenderJsonWithResult(result interface{}) {
	this.OnJsonResult(result)
}
func (this *WebBase) RenderJsonResultWithMessage(result interface{}, msg string) {
	this.OnJsonResultWithMessage(result, msg)
}

func (this *WebBase) RenderTemplateWithResults(tplName string, results interface{}, totalCount ...int64) {
	if len(totalCount) > 0 {
		this.OnResultsWithTotalCount(tplName, results, totalCount[0])
	} else {
		this.OnResults(tplName, results)
	}
}
func (this *WebBase) RenderTemplateWithResult(tplName string, result interface{}) {
	this.OnResult(tplName, result)
}

func (this *WebBase) RenderTemplateWithEntity(tplName string, result interface{}) {
	this.OnEntity(tplName, result)
}
func (this *WebBase) RenderTemplateWithEntities(tplName string, result interface{}, totalCount ...int64) {
	if len(totalCount) > 0 {
		this.OnEntitiesWithTotalCount(tplName, result, totalCount[0])
	} else {
		this.OnEntities(tplName, result)
	}
}

func (this *WebBase) RenderTemplate(tplName string, data ...interface{}) {
	this.SetTemplate(tplName)
	this.SetViewData(data...)
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

// RenderJsonValidationError
func (this *WebBase) RenderJsonWithBadRequest(msg string) {
	this.WebJsonRenderV1.RenderJsonValidationError(msg)
}

// RenderJsonForbidenError
func (this *WebBase) RenderJsonWithForbidden(msg string, abort bool) {
	this.WebJsonRenderV1.RenderJsonForbidenError(msg, abort)
}

// RenderJsonUnauthorizedError
func (this *WebBase) RenderJsonWithUnauthorized(msg string, abort bool) {
	this.WebJsonRenderV1.RenderJsonUnauthorizedError(msg, abort)
}

// OnErrorAny
func (this *WebBase) RenderJsonOrRedirect(path string, msg string, args ...interface{}) {
	this.WebRespUtil.OnErrorAny(path, msg, args...)
}

// OnRedirectError
func (this *WebBase) RedirectWithError(path string, msg string, args ...interface{}) {
	this.WebRespUtil.OnRedirectError(path, msg, args...)
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
	return this.WebParser.GetParam(key)
	//return this.BeegoController.Ctx.Input.Param(key)
}
