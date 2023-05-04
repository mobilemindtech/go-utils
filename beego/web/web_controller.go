package web

import (
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/core/validation"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/i18n"
	"github.com/mobilemindtec/go-utils/app/models"
	"github.com/mobilemindtec/go-utils/app/route"
	"github.com/mobilemindtec/go-utils/app/services"
	"github.com/mobilemindtec/go-utils/beego/db"
	"github.com/mobilemindtec/go-utils/beego/validator"
	"github.com/mobilemindtec/go-utils/cache"
	"github.com/mobilemindtec/go-utils/json"
	"github.com/mobilemindtec/go-utils/support"
	"github.com/mobilemindtec/go-utils/v2/criteria"
	"github.com/mobilemindtec/go-utils/v2/optional"
	"github.com/satori/go.uuid"
)

type WebController struct {
	beego.Controller
	support.JsonParser
	i18n.Locale

	EntityValidator *validator.EntityValidator
	Flash           *beego.FlashData
	Session         *db.Session
	ViewPath        string

	// models
	ModelAuditor    *models.Auditor
	ModelCidade     *models.Cidade
	ModelEstado     *models.Estado
	ModelRole       *models.Role
	ModelTenant     *models.Tenant
	ModelUser       *models.User
	ModelTenantUser *models.TenantUser
	ModelUserRole   *models.UserRole

	defaultPageLimit int64

	// auth
	userinfo *models.User
	tenant   *models.Tenant

	IsLoggedIn bool

	IsTokenLoggedIn bool

	Auth *services.AuthService

	UseJsonPackage bool

	InheritedController interface{}

	DoNotLoadTenantsOnSession bool

	CacheService            *cache.CacheService
	Character               *support.Character
	CacheKeysDeleteOnLogOut []string
	UploadPathDestination   string
}

func init() {
	LoadIl8n()
	LoadFuncs()
}

func (this *WebController) SetUseJsonPackage() *WebController {
	this.UseJsonPackage = true
	return this
}

func (this *WebController) loadLang() {
	// Reset language option.
	this.Lang = "" // This field is from i18n.Locale.

	// 1. Get language information from 'Accept-Language'.
	al := this.Ctx.Request.Header.Get("Accept-Language")
	if len(al) > 4 {
		al = al[:5] // Only compare first 5 letters.
		if i18n.IsExist(al) {
			this.Lang = al
		}
	}

	// 2. Default language is English.
	if len(this.Lang) == 0 {
		this.Lang = "pt-BR"
	}
}

// Prepare implemented Prepare() method for WebController.
// It's used for language option check and setting.
func (this *WebController) Prepare() {

	this.CacheService = cache.New()
	this.Character = support.NewCharacter()
	this.EntityValidator = validator.NewEntityValidator(this.Lang, this.ViewPath)
	this.DefaultLocation, _ = time.LoadLocation("America/Sao_Paulo")
	this.defaultPageLimit = 25

	this.loadLang()

	this.Flash = beego.NewFlash()
	this.FlashRead()

	// Set template level language option.
	this.Data["Lang"] = this.Lang
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())
	this.Data["dateLayout"] = dateLayout
	this.Data["datetimeLayout"] = datetimeLayout
	this.Data["timeLayout"] = timeLayout
	this.Data["today"] = time.Now().In(this.DefaultLocation).Format("02.01.2006")

	this.Session = this.WebControllerCreateSession()
	this.WebControllerLoadModels()

	this.AuthPrepare()

	this.Session.Tenant = this.GetAuthTenant()

	this.LoadTenants()
}

func (this *WebController) WebControllerCreateSession() *db.Session {
	if this.InheritedController != nil {
		if app, ok := this.InheritedController.(NestWebController); ok {
			return app.WebControllerCreateSession()
		}
	}
	return this.CreateSession()
}

func (this *WebController) CreateSession() *db.Session {

	session := db.NewSession()
	err := session.OpenTx()

	if err != nil {
		logs.Error("ERROR: db.NewSession: %v", err)
		this.Abort("500")
	}

	return session
}

func (this *WebController) AuthPrepare() {
	// login
	this.AppAuth()
	this.SetParams()

	this.IsLoggedIn = this.GetSession("userinfo") != nil
	this.IsTokenLoggedIn = this.GetSession("appuserinfo") != nil

	var tenant *models.Tenant
	tenantUuid := this.GetHeaderByNames("tenant", "X-Auth-Tenant")

	if len(tenantUuid) > 0 {
		loader := func() (*models.Tenant, error) {
			return this.ModelTenant.GetByUuidAndEnabled(tenantUuid)

		}
		tenant, _ = cache.Memoize(this.CacheService, tenantUuid, new(models.Tenant), loader)
		this.SetAuthTenant(tenant)
	}

	if this.IsLoggedIn || this.IsTokenLoggedIn {

		if this.IsLoggedIn {
			this.SetAuthUser(this.GetLogin())
		} else {
			this.SetAuthUser(this.GetTokenLogin())
		}

		if !this.IsTokenLoggedIn {
			tenant = this.GetAuthTenantSession()
			if tenant == nil {
				tenant, _ = this.ModelTenantUser.GetFirstTenant(this.GetAuthUser())
			}
		}

		if tenant == nil || !tenant.IsPersisted() {
			tenant = this.GetAuthUser().Tenant
			this.Session.Load(tenant)
		}

		if tenant == nil || !tenant.IsPersisted() {

			logs.Error("ERROR: user does not have active tenant")

			if this.IsTokenLoggedIn || this.IsJson() {
				this.OnJsonError("set header tenant")
			} else {
				this.OnErrorAny("/", "user does not has active tenant")
			}
			return
		}

		if !tenant.Enabled && !services.IsRootUser(this.GetAuthUser()) {
			logs.Error("ERROR: tenant ", tenant.Id, " - ", tenant.Name, " is disabled")

			if this.IsTokenLoggedIn || this.IsJson() {
				this.OnJsonError("operation not permitted to tenant")
			} else {
				this.LogOut()
				this.OnErrorAny("/", "operation not permitted to tenant")
			}
			return
		}

		this.SetAuthTenant(tenant)

		logs.Trace(":::::::::::::::::::::::::::::::::::::::::::::::::::::::::")
		logs.Trace(":: Tenant Id = %v", this.GetAuthTenant().Id)
		logs.Trace(":: Tenant Name = %v", this.GetAuthTenant().Name)
		logs.Trace(":: User Id = %v", this.GetAuthUser().Id)
		logs.Trace(":: User Name = %v", this.GetAuthUser().Name)
		logs.Trace(":: User Authority = %v", this.GetAuthUser().Role.Authority)
		logs.Trace(":: User Roles = %v", this.GetAuthUser().GetAuthorities())
		logs.Trace(":: User IsLoggedIn = %v", this.IsLoggedIn)
		logs.Trace(":: User IsTokenLoggedIn = %v", this.IsTokenLoggedIn)
		logs.Trace(":: User Auth Token = %v", this.GetToken())
		logs.Trace(":::::::::::::::::::::::::::::::::::::::::::::::::::::::::::")

		this.Data["UserInfo"] = this.GetAuthUser()
		this.Data["Tenant"] = this.GetAuthTenant()

		this.Auth = services.NewAuthService(this.GetAuthUser())
	}

	this.Data["IsLoggedIn"] = this.IsLoggedIn || this.IsTokenLoggedIn

	if this.IsLoggedIn || this.IsTokenLoggedIn {
		this.Data["IsAdmin"] = this.Auth.IsAdmin()
		this.Data["IsRoot"] = this.Auth.IsRoot()
	}

	this.UpSecurityAuth()
}

func (this *WebController) DisableXSRF(pathList []string) {
	if os.Getenv("BEEGO_MODE") == "test" {
		this.EnableXSRF = false
		logs.Trace("DISABLE ALL XSRF IN TEST MODE")
	} else {
		for _, url := range pathList {
			if strings.HasPrefix(this.Ctx.Input.URL(), url) {
				this.EnableXSRF = false
			}
		}
	}
}

func (this *WebController) FlashRead() {
	Flash := beego.ReadFromRequest(&this.Controller)

	if n, ok := Flash.Data["notice"]; ok {
		this.Flash.Notice(n)
	}

	if n, ok := Flash.Data["error"]; ok {
		this.Flash.Error(n)
	}

	if n, ok := Flash.Data["warning"]; ok {
		this.Flash.Warning(n)
	}

	if n, ok := Flash.Data["success"]; ok {
		this.Flash.Success(n)
	}
}

func (this *WebController) Finish() {

	logs.Trace("Finish, Commit")

	this.Session.Close()
	this.Session = nil
	this.CacheService.Close()
	this.CacheService = nil

	if app, ok := this.AppController.(NestFinisher); ok {
		app.NestFinish()
	}
}

func (this *WebController) Finally() {

	logs.Trace("Finally, Rollback")

	if this.Session != nil {
		this.Session.OnError().Close()
	}

	if this.CacheService != nil {
		this.CacheService.Close()
	}
}

func (this *WebController) Recover(info interface{}) {
	/*
	  this.Log("--------------- Controller.Recover ---------------")
	  this.Log("INFO: %v", info)
	  this.Log("STACKTRACE: %v", string(debug.Stack()))
	  this.Log("--------------- Controller.Recover ---------------")
	*/
	if app, ok := this.AppController.(NestRecover); ok {
		info := &RecoverInfo{Error: fmt.Sprintf("%v", info), StackTrace: string(debug.Stack())}
		app.NextOnRecover(info)
	}

}

func (this *WebController) Rollback() {
	if this.Session != nil {
		this.Session.OnError()
	}
}

func (this *WebController) OnEntity(viewName string, entity interface{}) {
	this.Data["entity"] = entity
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *WebController) OnEntityError(viewName string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Rollback()
	this.Flash.Error(message)
	this.Data["entity"] = entity
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *WebController) OnEntities(viewName string, entities interface{}) {
	this.Data["entities"] = entities
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *WebController) OnEntitiesWithTotalCount(viewName string, entities interface{}, totalCount int64) {
	this.Data["entities"] = entities
	this.Data["totalCount"] = totalCount
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *WebController) OnResult(viewName string, result interface{}) {
	this.Data["result"] = result
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *WebController) OnResults(viewName string, results interface{}) {
	this.Data["results"] = results
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *WebController) SetViewModel(name string, data interface{}) *WebController {
	this.Data[name] = data
	return this
}

func (this *WebController) SetResults(results interface{}) *WebController {
	this.Data["results"] = results
	this.Data["entities"] = results
	return this
}

func (this *WebController) SetResultsAndTotalCount(results interface{}, totalCount int64) *WebController {
	this.Data["results"] = results
	this.Data["entities"] = results
	this.Data["totalCount"] = totalCount
	return this
}

func (this *WebController) SetResult(result interface{}) *WebController {
	this.Data["result"] = result
	this.Data["entity"] = result
	return this
}

func (this *WebController) OnResultsWithTotalCount(viewName string, results interface{}, totalCount int64) {
	this.Data["results"] = results
	this.Data["totalCount"] = totalCount
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *WebController) RenderJsonResult(opt interface{}) {

	switch opt.(type) {
	case *optional.Some:
		it := opt.(*optional.Some).Item
		if optional.IsSlice(it) {
			this.OnJsonResults(it)
		} else {
			this.OnJsonResult(it)
		}
		break
	case *optional.Empty:
		this.OnJsonResults([]interface{}{})
		break
	case *optional.None:
		this.OnJson200()
		break
	case *optional.Fail:
		this.OnJsonError(fmt.Sprintf("%v", opt.(*optional.Fail).Error))
		break
	case *support.JsonResult:
		this.OnJson(opt.(*support.JsonResult))
		break
	default:
		this.OnJsonError(fmt.Sprintf("unknow optional value: %v", opt))
		break
	}

}

func (this *WebController) RenderJson(opt interface{}) {

	var dataResult interface{}
	var statusCodeResult = 200

	switch opt.(type) {
	case *optional.Some:

		someVal := opt.(*optional.Some).Item

		switch someVal.(type) {
		case *criteria.Page:
			dataResult = someVal
			break
		default:
			dataResult = map[string]interface{}{
				"data": someVal,
			}
		}

		break
	case *optional.None, *optional.Empty:
		statusCodeResult = 404
		break
	case *optional.Fail:

		f := opt.(*optional.Fail)
		err := f.Error
		statusCode := 500

		data := map[string]interface{}{
			"error":   true,
			"message": fmt.Sprintf("%v", err),
		}

		if err.Error() == "validation error" {
			data["validation"] = f.Item
			statusCode = 400
		}

		dataResult = data
		statusCodeResult = statusCode
		break
	default:
		dataResult = map[string]interface{}{
			"error":   true,
			"message": fmt.Sprintf("unknow optional value: %v", opt),
		}

		statusCodeResult = 500
		break
	}

	j, err := json.Encode(dataResult)

	if err != nil {
		logs.Error("ERROR JSON ENCODE: %v", err)
		this.Ctx.Output.SetStatus(500)
		this.Ctx.Output.Body([]byte(fmt.Sprint(`{ "error": true, "message": "%v" }`, err.Error())))
	} else {
		logs.Trace("REPONSE STATUS CODE = %v", statusCodeResult)
		this.Ctx.Output.SetStatus(statusCodeResult)
		this.Ctx.Output.Body(j)
	}

	this.ServeJSON()
}

func (this *WebController) OnJsonResult(result interface{}) {
	this.Data["json"] = &support.JsonResult{Result: result, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *WebController) GetJsonResult() (*support.JsonResult, bool) {
	if this.Data["json"] != nil {
		if j, ok := this.Data["json"].(*support.JsonResult); ok {
			return j, ok
		}
	}
	return nil, false
}

func (this *WebController) OnJsonMessage(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = &support.JsonResult{Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultError(result interface{}, format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = &support.JsonResult{Result: result, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultWithMessage(result interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = &support.JsonResult{Result: result, Error: false, Message: message, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *WebController) OnJsonResults(results interface{}) {
	this.Data["json"] = &support.JsonResult{Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultAndResults(result interface{}, results interface{}) {
	this.Data["json"] = &support.JsonResult{Result: result, Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultsWithTotalCount(results interface{}, totalCount int64) {
	this.Data["json"] = &support.JsonResult{Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultAndResultsWithTotalCount(result interface{}, results interface{}, totalCount int64) {
	this.Data["json"] = &support.JsonResult{Result: result, Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultsError(results interface{}, format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = &support.JsonResult{Results: results, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *WebController) OnJson(json *support.JsonResult) {
	this.Data["json"] = json
	this.ServeJSON()
}

func (this *WebController) OnJsonMap(jsonMap map[string]interface{}) {
	this.Data["json"] = jsonMap
	this.ServeJSON()
}

func (this *WebController) OnJsonError(format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	result := &support.JsonResult{Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.OnJson(result)
}

func (this *WebController) ServeJSON() {
	if this.UseJsonPackage {
		result := this.Data["json"]
		bdata, err := json.Encode(result)
		if err != nil {
			this.Data["json"] = &support.JsonResult{Message: fmt.Sprintf("Error json.Encode: %v", err), Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()}
			this.Controller.ServeJSON()
		} else {
			this.Ctx.Output.Header("Content-Type", "application/json")
			this.Ctx.Output.Body(bdata)
		}
	} else {
		this.Controller.ServeJSON()
	}
}

func (this *WebController) RenderTemplate(viewName string) {
	this.OnTemplate(viewName)
}

func (this *WebController) RenderJsonMap(jsonMap map[string]interface{}) {
	this.OnJsonMap(jsonMap)
}

func (this *WebController) OnRender(data interface{}) {
	switch data.(type) {
	case string:
		this.OnTemplate(data.(string))
		break
	case map[string]interface{}:
		this.OnJsonMap(data.(map[string]interface{}))
		break
	case *support.JsonResult:
		this.OnJson(data.(*support.JsonResult))
		break
	default:
		panic("no render selected")
	}
}

func (this *WebController) RenderJsonError(format string, v ...interface{}) {
	this.OnJsonError(format, v...)
}

func (this *WebController) OnJsonErrorNotRollback(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&support.JsonResult{Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonOk(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&support.JsonResult{Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJson200() {
	this.OnJson(&support.JsonResult{CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OkAsJson(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&support.JsonResult{CurrentUnixTime: this.GetCurrentTimeUnix(), Message: message})
}

func (this *WebController) OkAsHtml(message string) {
	this.Ctx.Output.Body([]byte(message))
}

func (this *WebController) Ok() {
	this.Ctx.Output.SetStatus(200)
}

func (this *WebController) OkAsText(message string) {
	this.Ctx.Output.Body([]byte(message))
}

func (this *WebController) OnJsonValidationError() {
	this.Rollback()
	errors := this.Data["errors"].(map[string]string)
	this.OnJson(&support.JsonResult{Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonValidationWithErrors(errors map[string]string) {
	this.Rollback()
	this.OnJson(&support.JsonResult{Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonValidationMessageWithErrors(message string, errors map[string]string) {
	this.Rollback()
	this.OnJson(&support.JsonResult{Message: message, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonValidationWithResultAndMessageAndErrors(result interface{}, message string, errors map[string]string) {
	this.Rollback()
	this.OnJson(&support.JsonResult{Message: message, Result: result, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonValidationWithResultsAndMessageAndErrors(results interface{}, message string, errors map[string]string) {
	this.Rollback()
	this.OnJson(&support.JsonResult{Message: message, Results: results, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnTemplate(viewName string) {
	this.TplName = fmt.Sprintf("%s/%s.tpl", this.ViewPath, viewName)
	this.OnFlash(false)
}

func (this *WebController) OnFullTemplate(tplName string) {
	this.TplName = fmt.Sprintf("%s.tpl", tplName)
	this.OnFlash(false)
}

func (this *WebController) OnPureTemplate(templateName string) {
	this.TplName = templateName
	this.OnFlash(false)
}

func (this *WebController) OnRedirect(action string) {
	this.OnFlash(true)
	if this.Ctx.Input.URL() == "action" {
		this.Abort("500")
	} else {
		this.Redirect(action, 302)
	}
}

func (this *WebController) OnRedirectError(action string, format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.Flash.Error(message)
	this.OnFlash(true)
	if this.Ctx.Input.URL() == "action" {
		this.Abort("500")
	} else {
		this.Redirect(action, 302)
	}
}

func (this *WebController) OnRedirectSuccess(action string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Flash.Success(message)
	this.OnFlash(true)
	if this.Ctx.Input.URL() == "action" {
		this.Abort("500")
	} else {
		this.Redirect(action, 302)
	}
}

// executes redirect or OnJsonError
func (this *WebController) OnErrorAny(path string, format string, v ...interface{}) {

	//this.Log("** this.IsJson() %v", this.IsJson() )
	message := fmt.Sprintf(format, v...)
	if this.IsJson() {
		this.OnJsonError(message)
	} else {
		this.OnRedirectError(path, message)
	}
}

// executes redirect or OnJsonOk
func (this *WebController) OnOkAny(path string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if this.IsJson() {
		this.OnJsonOk(message)
	} else {
		this.Flash.Success(message)
		this.OnRedirect(path)
	}

}

// executes OnEntity or OnJsonValidationError
func (this *WebController) OnValidationErrorAny(view string, entity interface{}) {

	if this.IsJson() {
		this.OnJsonValidationError()
	} else {
		this.Rollback()
		this.OnEntity(view, entity)
	}

}

// executes OnEntity or OnJsonError
func (this *WebController) OnEntityErrorAny(view string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if this.IsJson() {
		this.OnJsonError(message)
	} else {
		this.Rollback()
		this.Flash.Error(message)
		this.OnEntity(view, entity)
	}

}

// executes OnEntity or OnJsonResultWithMessage
func (this *WebController) OnEntityAny(view string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if this.IsJson() {
		this.OnJsonResultWithMessage(entity, message)
	} else {
		this.Flash.Success(message)
		this.OnEntity(view, entity)
	}

}

// executes OnResults or OnJsonResults
func (this *WebController) OnResultsAny(viewName string, results interface{}) {

	if this.IsJson() {
		this.OnJsonResults(results)
	} else {
		this.OnResults(viewName, results)
	}

}

// executes  OnResultsWithTotalCount or OnJsonResultsWithTotalCount
func (this *WebController) OnResultsWithTotalCountAny(viewName string, results interface{}, totalCount int64) {

	if this.IsJson() {
		this.OnJsonResultsWithTotalCount(results, totalCount)
	} else {
		this.OnResultsWithTotalCount(viewName, results, totalCount)
	}

}

func (this *WebController) OnFlash(store bool) {
	if store {
		this.Flash.Store(&this.Controller)
	} else {
		this.Data["Flash"] = this.Flash.Data
		this.Data["flash"] = this.Flash.Data
	}
}

func (this *WebController) GetMessage(key string, args ...interface{}) string {
	return i18n.Tr(this.Lang, key, args)
}

func (this *WebController) OnValidate(entity interface{}, custonValidation func(validator *validation.Validation)) bool {

	result, _ := this.EntityValidator.IsValid(entity, custonValidation)

	if result.HasError {
		this.Flash.Error(this.GetMessage("cadastros.validacao"))
		this.EntityValidator.CopyErrorsToView(result, this.Data)
	}

	return result.HasError == false
}

func (this *WebController) OnParseForm(entity interface{}) {
	if err := this.ParseForm(entity); err != nil {
		logs.Error("*******************************************")
		logs.Error("***** ERROR parse FORM to JSON: %v", err.Error())
		logs.Error("*******************************************")
		this.Abort("500")
	}
}

func (this *WebController) OnJsonParseForm(entity interface{}) {
	this.Form2Json(entity)
}

/**
*
* use this.Form2JsonWithCnf(entity, map[string]string{
	* 	"FloatFieldName": "float",
	* 	"IntFieldName": "int",
	* 	"BoolFieldName": "bool",
	*   "DateFieldName": "date:layout",
	* })
*
*/
func (this *WebController) OnJsonParseFormWithFieldsConfigs(entity interface{}, configs map[string]string) {
	this.Form2JsonWithCnf(entity, configs)
}

func (this *WebController) Form2Json(entity interface{}) {
	this.Form2JsonWithCnf(entity, nil)
}

func (this *WebController) Form2JsonWithCnf(entity interface{}, configs map[string]string) {
	if err := this.FormToModelWithFieldsConfigs(this.Ctx, entity, configs); err != nil {
		logs.Error("*******************************************")
		logs.Error("***** ERROR parse FORM to JSON: %v ", err.Error())
		logs.Error("*******************************************")
		this.Abort("500")
	}
}

func (this *WebController) ParamParseMoney(s string) float64 {
	return this.ParamParseFloat(s)
}

// remove ,(virgula) do valor em params que vem como val de input com jquery money
// exemplo 45,000.00 vira 45000.00
func (this *WebController) ParamParseFloat(s string) float64 {
	var semic string = ","
	replaced := strings.Replace(s, semic, "", -1) // troca , por espaço
	precoFloat, err := strconv.ParseFloat(replaced, 64)
	var returnValue float64
	if err == nil {
		returnValue = precoFloat
	} else {
		logs.Error("*******************************************")
		logs.Error("****** ERROR parse string to float64 for stringv", s)
		logs.Error("*******************************************")
		this.Abort("500")
	}

	return returnValue
}

func (this *WebController) OnParseJson(entity interface{}) {
	if err := this.JsonToModel(this.Ctx, entity); err != nil {
		logs.Error("*******************************************")
		logs.Error("***** ERROR on parse json ", err.Error())
		logs.Error("*******************************************")
		this.Abort("500")
	}
}

func (this *WebController) RawBody() []byte {
	return this.Ctx.Input.RequestBody
}

func (this *WebController) NotFound() {
	this.Ctx.Output.SetStatus(404)
}

func (this *WebController) ServerError() {
	this.Ctx.Output.SetStatus(500)
}

func (this *WebController) BadRequest() {
	this.Ctx.Output.SetStatus(400)
}

func (this *WebController) Unauthorized() {
	this.Ctx.Output.SetStatus(401)
}

func (this *WebController) Forbidden() {
	this.Ctx.Output.SetStatus(403)
}

func (this *WebController) HasPath(paths ...string) bool {
	for _, it := range paths {
		if strings.HasPrefix(this.Ctx.Input.URL(), it) {
			return true
		}
	}
	return false
}

func (this *WebController) IsJson() bool {
	return this.Ctx.Input.AcceptsJSON()
}

func (this *WebController) IsAjax() bool {
	return this.Ctx.Input.IsAjax()
}

func (this *WebController) GetToken() string {
	return this.GetHeaderByName("X-Auth-Token")
}

func (this *WebController) GetHeaderByName(name string) string {
	return this.Ctx.Request.Header.Get(name)
}

func (this *WebController) GetHeaderByNames(names ...string) string {

	for _, name := range names {
		val := this.Ctx.Request.Header.Get(name)

		if len(val) > 0 {
			return val
		}
	}

	return ""
}

func (this *WebController) Log(format string, v ...interface{}) {
	logs.Debug(fmt.Sprintf(format, v...))
}

func (this *WebController) GetCurrentTimeUnix() int64 {
	return this.GetCurrentTime().Unix()
}

func (this *WebController) GetCurrentTime() time.Time {
	return time.Now().In(this.DefaultLocation)
}

func (this *WebController) GetPage() *db.Page {
	page := new(db.Page)

	var defaultLimit int64 = 25

	if this.IsJson() {

		if this.Ctx.Input.IsPost() {
			jsonMap, _ := this.JsonToMap(this.Ctx)

			if _, ok := jsonMap["limit"]; ok {
				page.Limit = optional.
					New[int64](this.GetJsonInt64(jsonMap, "limit")).
					OrElse(defaultLimit)

				page.Offset = this.GetJsonInt64(jsonMap, "offset")

				page.Sort = optional.
					New[string](this.GetJsonString(jsonMap, "order_column")).
					OrElse(this.GetJsonString(jsonMap, "sort"))

				page.Order = optional.
					New[string](this.GetJsonString(jsonMap, "order_sort")).
					OrElse(this.GetJsonString(jsonMap, "order"))

				page.Order = this.GetJsonString(jsonMap, "order_sort")
				page.Search = this.GetJsonString(jsonMap, "search")

				return page
			}
		}
	}

	page.Limit = optional.
		New[int64](this.GetIntByKey("limit")).
		OrElse(defaultLimit)

	page.Sort = optional.
		New[string](this.GetStringByKey("order_column")).
		OrElse(this.GetStringByKey("sort"))

	page.Order = optional.
		New[string](this.GetStringByKey("order_sort")).
		OrElse(this.GetStringByKey("order"))

	page.Offset = this.GetIntByKey("offset")
	page.Search = this.GetStringByKey("search")

	return page
}

func (this *WebController) StringToInt(text string) int {
	val, _ := strconv.Atoi(text)
	return val
}

func (this *WebController) StringToInt64(text string) int64 {
	val, _ := strconv.ParseInt(text, 10, 64)
	return val
}

func (this *WebController) IntToString(val int) string {
	return fmt.Sprintf("%v", val)
}

func (this *WebController) Int64ToString(val int64) string {
	return fmt.Sprintf("%v", val)
}

func (this *WebController) GetId() int64 {
	return this.GetIntParam(":id")
}

func (this *WebController) GetParam(key string) string {

	if !strings.HasPrefix(key, ":") {
		key = fmt.Sprintf(":%v", key)
	}

	return this.Ctx.Input.Param(key)
}

func (this *WebController) GetStringParam(key string) string {
	return this.GetParam(key)
}

func (this *WebController) GetIntParam(key string) int64 {
	id := this.GetParam(key)
	intid, _ := strconv.ParseInt(id, 10, 64)
	return intid
}

func (this *WebController) GetInt32Param(key string) int {
	val := this.GetParam(key)
	intid, _ := strconv.Atoi(val)
	return intid
}

func (this *WebController) GetBoolParam(key string) bool {
	val := this.GetParam(key)
	return val == "true"
}

func (this *WebController) GetIntByKey(key string) int64 {
	val := this.Ctx.Input.Query(key)
	intid, _ := strconv.ParseInt(val, 10, 64)
	return intid
}

func (this *WebController) GetBoolByKey(key string) bool {
	val := this.Ctx.Input.Query(key)
	boolean, _ := strconv.ParseBool(val)
	return boolean
}

func (this *WebController) GetStringByKey(key string) string {
	return this.Ctx.Input.Query(key)
}

func (this *WebController) GetDateByKey(key string) (time.Time, error) {
	date := this.Ctx.Input.Query(key)
	return this.ParseDate(date)
}

func (this *WebController) ParseDateByKey(key string, layout string) (time.Time, error) {
	date := this.Ctx.Input.Query(key)
	return time.ParseInLocation(layout, date, this.DefaultLocation)
}

// deprecated
func (this *WebController) ParseDate(date string) (time.Time, error) {
	return time.ParseInLocation(dateLayout, date, this.DefaultLocation)
}

// deprecated
func (this *WebController) ParseDateTime(date string) (time.Time, error) {
	return time.ParseInLocation(datetimeLayout, date, this.DefaultLocation)
}

// deprecated
func (this *WebController) ParseJsonDate(date string) (time.Time, error) {
	return time.ParseInLocation(jsonDateLayout, date, this.DefaultLocation)
}

func (this *WebController) NormalizePageSortKey(key string) string {
	if strings.Contains(key, ".") {
		return strings.Replace(key, ".", "__", -1)
	}
	return key
}

func (this *WebController) CheckboxToBool(key string) bool {
	arr := this.Ctx.Request.Form[key]
	return len(arr) > 0
}

func (this *WebController) WebControllerLoadModels() {
	if this.InheritedController != nil {
		if app, ok := this.InheritedController.(NestWebController); ok {
			app.WebControllerLoadModels()
		}
	}

	this.LoadModels()
}

func (this *WebController) LoadModels() {
	this.ModelAuditor = models.NewAuditor(this.Session)
	this.ModelCidade = models.NewCidade(this.Session)
	this.ModelEstado = models.NewEstado(this.Session)
	this.ModelRole = models.NewRole(this.Session)
	this.ModelTenant = models.NewTenant(this.Session)
	this.ModelUser = models.NewUser(this.Session)
	this.ModelTenantUser = models.NewTenantUser(this.Session)
	this.ModelUserRole = models.NewUserRole(this.Session)
}

func (this *WebController) LoadTenants() {

	if this.IsLoggedIn && !this.DoNotLoadTenantsOnSession {

		cacheKey := cache.CacheKey("tenants_user_", this.GetAuthUser().Id)
		this.DeleteCacheOnLogout(cacheKey)

		loader := func() ([]interface{}, error) {

			tenants := []*models.Tenant{}
			if this.Auth.IsRoot() {
				its, _ := this.ModelTenant.List()
				tenants = *its
			} else {
				list, _ := this.ModelTenantUser.ListByUserAdmin(this.GetAuthUser())

				for _, it := range *list {

					if !it.Enabled {
						continue
					}

					this.Session.Load(it.Tenant)
					tenants = append(tenants, it.Tenant)
				}
			}
			authorizeds := []interface{}{}
			for _, it := range tenants {
				authorizeds = append(authorizeds, it)
			}
			return authorizeds, nil
		}

		authorizeds, _ := cache.Memoize(this.CacheService, cacheKey, new([]interface{}), loader)

		this.Session.SetAuthorizedTenants(authorizeds)
		this.Data["AvailableTenants"] = authorizeds
	} else {
		this.Data["AvailableTenants"] = []*models.Tenant{}
	}
}

func (this *WebController) Audit(format string, v ...interface{}) {
	auditor := services.NewAuditorService(this.Session, this.Lang, this.GetAuditorInfo())
	auditor.OnAuditWithNewDbSession(format, v...)
}

func (this *WebController) GetAuditorInfo() *services.AuditorInfo {
	return &services.AuditorInfo{Tenant: this.GetAuthTenant(), User: this.GetAuthUser()}
}

func (this *WebController) GetLastUpdate() time.Time {
	lastUpdateUnix, _ := this.GetInt64("lastUpdate")
	var lastUpdate time.Time

	if lastUpdateUnix > 0 {
		lastUpdate = time.Unix(lastUpdateUnix, 0).In(this.DefaultLocation)
	}

	return lastUpdate
}

func (this *WebController) AppAuth() {

	token := this.GetToken()

	if strings.TrimSpace(token) != "" {

		auth := services.NewLoginService(this.Lang, this.Session)

		logs.Debug("Authenticate by token %v", token)

		loader := func() (*models.User, error) {
			return auth.AuthenticateToken(token)
		}

		user, err := cache.Memoize(this.CacheService, token, new(models.User), loader)

		if err != nil {
			logs.Error("LOGIN ERROR: %v", err)
			this.LogOut()
			return
		}

		if user == nil || !user.IsPersisted() {
			logs.Error("LOGIN ERROR: user not found!")
			this.LogOut()
			return
		}

		this.SetTokenLogin(user)
	}
}

func (this *WebController) GetLogin() *models.User {
	id, _ := this.GetSession("userinfo").(int64)
	return this.memoizeUser(id)
}

func (this *WebController) GetTokenLogin() *models.User {
	id, _ := this.GetSession("appuserinfo").(int64)
	return this.memoizeUser(id)
}

func (this *WebController) SessionLogOut() {
	this.LogOut()
}

func (this *WebController) LogOut() {
	this.CacheService.Delete(this.CacheKeysDeleteOnLogOut...)
	this.DelSession("userinfo")
	this.DelSession("appuserinfo")
	this.DelSession("authtenantid")
	this.DestroySession()
}

func (this *WebController) SetLogin(user *models.User) {
	this.SetSession("userinfo", user.Id)
}

func (this *WebController) SetTokenLogin(user *models.User) {
	this.SetSession("appuserinfo", user.Id)
}

func (this *WebController) LoginPath() string {
	return this.URLFor("LoginController.Login")
}

func (this *WebController) SetParams() {
	this.Data["Params"] = make(map[string]string)

	values, _ := this.Input()

	for k, v := range values {
		this.Data["Params"].(map[string]string)[k] = v[0]
	}
}

func (this *WebController) OnLoginRedirect() {
	path := this.Ctx.Input.URI()
	if !strings.Contains("?", path) {
		path = "?redirect=" + path
	}
	this.Ctx.Redirect(302, this.LoginPath()+path)
}

func (this *WebController) AuthCheck() {
	if !this.IsLoggedIn && !this.IsTokenLoggedIn {
		if this.IsJson() {
			this.OnJsonError(this.GetMessage("security.notLoggedIn"))
			this.Abort("401")
		} else {
			this.OnLoginRedirect()
		}
	}
}

func (this *WebController) AuthCheckRoot() {
	if !this.IsLoggedIn {
		if this.IsJson() {
			this.OnJsonError(this.GetMessage("security.notLoggedIn"))
			this.Abort("401")
		} else {
			this.OnLoginRedirect()
		}
	}

	if !this.Auth.IsRoot() {
		if this.IsJson() {
			this.OnJsonError(this.GetMessage("security.rootRequired"))
			this.Abort("401")
		} else {
			this.OnRedirect("/")
		}
	}
}

func (this *WebController) AuthCheckAdmin() {
	if !this.IsLoggedIn {
		if this.IsJson() {
			this.OnJsonError(this.GetMessage("security.notLoggedIn"))
			this.Abort("401")
		} else {
			this.OnLoginRedirect()
			this.OnRedirectError("/", this.GetMessage("security.rootRequired"))
		}
	}

	if !this.Auth.IsRoot() && !this.Auth.IsAdmin() {
		if this.IsJson() {
			this.OnJsonError(this.GetMessage("security.rootRequired"))
			this.Abort("401")
		} else {
			this.OnRedirectError("/", this.GetMessage("security.rootRequired"))
		}
	}
}

func (this *WebController) UpSecurityAuth() bool {

	roles := []string{}

	if this.Auth != nil {
		roles = this.Auth.GetUserRoles()
	}

	if !route.IsRouteAuthorized(this.Ctx, roles) {

		logs.Warn("WARN: path %v not authorized ", this.Ctx.Input.URL())

		if !this.IsLoggedIn && !this.IsTokenLoggedIn {
			if this.IsJson() {
				this.OnJsonError(this.GetMessage("security.notLoggedIn"))
				//this.Abort("401")
			} else {
				this.OnLoginRedirect()
			}
			return false
		}

		if this.IsJson() {
			this.OnJsonError(this.GetMessage("security.denied"))
			//this.Abort("401")
		} else {
			this.OnRedirect("/")
		}

		return false

	}

	return true
}

func (this *WebController) HasTenantAuth(tenant *models.Tenant) bool {
	if !this.Auth.IsRoot() {

		ModelTenantUser := this.ModelTenantUser

		loader := func() bool {
			item, _ := ModelTenantUser.FindByUserAndTenant(this.GetAuthUser(), tenant)
			return item != nil && item.IsPersisted()
		}
		parser := func(v string) bool {
			return v == "true"
		}
		cacheKey := cache.CacheKey("has_user", this.GetAuthUser().Id, "tenant", tenant.Id)
		return cache.MemoizeVal(this.CacheService, cacheKey, parser, loader)

	}
	return true
}

func (this *WebController) SetAuthTenantSession(tenant *models.Tenant) {

	if this.HasTenantAuth(tenant) {
		logs.Error("Set tenant session. user %v now is using tenant %v", this.GetAuthUser().Id, tenant.Id)
		this.SetSession("authtenantid", tenant.Id)
	} else {
		logs.Error("Cannot set tenant session. user %v not enable to use tenant %v", this.GetAuthUser().Id, tenant.Id)
	}

}

func (this *WebController) GetAuthTenantSession() *models.Tenant {
	var tenant *models.Tenant

	if id, ok := this.GetSession("authtenantid").(int64); ok && id > 0 {
		loader := func() (*models.Tenant, error) {
			tenant := models.Tenant{Id: int64(id)}
			this.Session.Load(&tenant)
			return &tenant, nil
		}
		tenant, _ = cache.Memoize(this.CacheService, cache.CacheKey("tenant_", id), new(models.Tenant), loader)
	}

	return tenant
}

func (this *WebController) GetAuthTenant() *models.Tenant {
	tenant := this.GetAuthTenantSession()
	if tenant != nil && tenant.IsPersisted() {
		return tenant
	}
	return this.tenant
}

func (this *WebController) SetAuthTenant(t *models.Tenant) {
	this.tenant = t
}

func (this *WebController) SetAuthUser(u *models.User) {
	this.userinfo = u
}

func (this *WebController) GetAuthUser() *models.User {
	return this.userinfo
}

func (this *WebController) DeleteCacheOnLogout(keys ...string) {
	this.CacheKeysDeleteOnLogOut = append(this.CacheKeysDeleteOnLogOut, keys...)
}

func (this *WebController) memoizeUser(id int64) *models.User {
	var user *models.User
	loader := func() (*models.User, error) {
		e, err := this.Session.FindById(new(models.User), id)
		if err != nil {
			return nil, err
		}
		user := e.(*models.User)
		this.ModelUser.LoadRelated(user)
		return user, nil
	}
	user, _ = cache.Memoize(this.CacheService, cache.CacheKey("user_", id), new(models.User), loader)
	return user
}

func (this *WebController) PrepareUploadedFile(fileOriginalName string, fileName string) (string, string, error) {

	path := this.UploadPathDestination + string(os.PathSeparator)

	if err := os.MkdirAll(path, 0777); err != nil {
		return "", "", err
	}

	splited := strings.Split(fileOriginalName, ".")

	if len(splited) == 0 {
		return "", "", errors.New("file extension not found")
	}

	ext := splited[len(splited)-1]

	if !strings.Contains(fileName, ".") {
		fileName += "." + ext
	}

	path += fileName

	logs.Trace("## save file on ", path)

	return path, fileName, nil
}

func (this *WebController) GetUploadedFileSavePath(fieldName string) string {

	if err := os.MkdirAll(this.UploadPathDestination, 0777); err != nil {
		logs.Error("Error on create uploaded file save path %v: %v", this.UploadPathDestination, err)
	}

	return fmt.Sprintf("%v/%v", this.UploadPathDestination, fieldName)
}

func (this *WebController) HasUploadedFile(fname string) (bool, error) {
	_, _, err := this.GetFile(fname)
	if err == http.ErrMissingFile {
		return false, nil
	} else if err != nil {
		return false, errors.New(fmt.Sprintf("Erro ao buscar documento: %v", err))
	}
	return true, nil
}

func (this *WebController) GetUploadedFileExt(fieldName string, required bool) (bool, string, error) {

	_, multipartFileHeader, err := this.GetFile(fieldName)

	if err == http.ErrMissingFile {

		if required {
			return false, "", fmt.Errorf("Selecione uma imagem para enviar")
		}

		return false, "", nil

	} else if err != nil {
		return false, "", fmt.Errorf("Erro ao enviar arquivo: %v", err)
	}

	originalName := this.Character.Transform(multipartFileHeader.Filename)
	splited := strings.Split(originalName, ".")

	if len(splited) == 0 {
		return false, "", fmt.Errorf("file extension not found")
	}

	ext := splited[len(splited)-1]
	uuid := uuid.NewV4()
	newFileName := fmt.Sprintf("%v.%v", uuid.String(), ext)

	return true, newFileName, nil
}
