package web

import (
	"errors"
	"fmt"
	"github.com/mobilemindtech/go-utils/app/util"
	"github.com/mobilemindtech/go-utils/beego/web/response"
	"html/template"
	"mime/multipart"
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
	"github.com/mobilemindtech/go-io/option"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-utils/app/models"
	"github.com/mobilemindtech/go-utils/app/route"
	"github.com/mobilemindtech/go-utils/app/services"
	"github.com/mobilemindtech/go-utils/beego/db"
	"github.com/mobilemindtech/go-utils/beego/validator"
	"github.com/mobilemindtech/go-utils/cache"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/support"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	"github.com/mobilemindtech/go-utils/v2/ioc"
	"github.com/mobilemindtech/go-utils/v2/lists"
	"github.com/mobilemindtech/go-utils/v2/maps"
	"github.com/mobilemindtech/go-utils/v2/inline"
	"github.com/mobilemindtech/go-utils/v2/optional"
	uuid "github.com/satori/go.uuid"
)

type Multipart struct {
	FileHeader *multipart.FileHeader
	File       *multipart.File
	Key        string
}

func (this *Multipart) FileName() string {
	return this.FileHeader.Filename
}

func (this *Multipart) FileExtension() string {
	ex := ""
	splited := strings.Split(this.FileName(), ".")

	if len(splited) > 0 {
		ex = splited[len(splited)-1]
	}
	return strings.ToLower(ex)
}

type WebController struct {
	beego.Controller
	support.JsonParser
	i18n.Locale

	EntityValidator *validator.EntityValidator `inject:""`
	Flash           *beego.FlashData
	Session         *db.Session `inject:""`
	ViewPath        string

	// models
	ModelAuditor    *models.Auditor    `inject:""`
	ModelCidade     *models.Cidade     `inject:""`
	ModelEstado     *models.Estado     `inject:""`
	ModelRole       *models.Role       `inject:""`
	ModelTenant     *models.Tenant     `inject:""`
	ModelUser       *models.User       `inject:""`
	ModelTenantUser *models.TenantUser `inject:""`
	ModelUserRole   *models.UserRole   `inject:""`

	defaultPageLimit int64

	// auth
	userinfo *models.User
	tenant   *models.Tenant

	IsWebLoggedIn       bool
	IsTokenLoggedIn     bool
	IsCustomAppLoggedIn bool

	Auth *services.AuthService

	UseJsonPackage         bool
	JsonPackageAsCamelCase bool
	ExitWithHttpCode       bool

	CustonJsonEncoder func(interface{}) ([]byte, error)

	NewJSON func() *json.JSON

	InheritedController interface{}

	DoNotLoadTenantsOnSession bool

	CacheService            *cache.CacheService `inject:""`
	Character               *support.Character  `inject:""`
	CacheKeysDeleteOnLogOut []string
	UploadPathDestination   string

	CustomAppAuthenticator func(*models.App) (*models.User, error)

	Container *ioc.Container
}

func init() {
	LoadIl8n()
	LoadFuncs()

}

func (this *WebController) SetUseJsonPackage() *WebController {
	this.UseJsonPackage = true
	return this
}

func (this *WebController) SetJsonPackageAsCamelCase() *WebController {
	this.JsonPackageAsCamelCase = true
	return this
}

func (this *WebController) SetCustonJsonEncoder(f func(interface{}) ([]byte, error)) *WebController {
	this.CustonJsonEncoder = f
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
	this.Data["dateLayout"] = util.DateBrLayout
	this.Data["datetimeLayout"] = util.DateTimeBrLayout
	this.Data["timeLayout"] = util.TimeMinutesLayout
	this.Data["today"] = time.Now().In(this.DefaultLocation).Format("02.01.2006")

	this.Session = this.WebControllerCreateSession()
	this.WebControllerLoadModels()

	this.Auth = services.NewAuthService(this.Session)

	this.AuthPrepare()

	this.Session.Tenant = this.GetAuthTenant()

	this.LoadTenants()
}

func (this *WebController) SetUp() {

}

func (this *WebController) WebControllerCreateSession() *db.Session {
	if this.InheritedController != nil {
		if app, ok := this.InheritedController.(NestWebController); ok {
			return app.WebControllerCreateSession()
		}
	}
	return this.CreateSession()
}

func (this *WebController) tryAppAuthenticate(token string) (*models.User, error) {

	if this.CustomAppAuthenticator == nil {
		return nil, nil
	}

	first := db.RunWithIgnoreTenantFilter(
		this.Session,
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
		user, err := this.CustomAppAuthenticator(app)

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

	this.IsWebLoggedIn = this.GetSession("userinfo") != nil
	this.IsTokenLoggedIn = this.GetSession("appuserinfo") != nil
	this.IsCustomAppLoggedIn = this.GetSession("customappuserinfo") != nil

	var tenant *models.Tenant

	tenantUuid := this.GetHeaderTenant()

	if len(tenantUuid) > 0 {
		loader := func() (*models.Tenant, error) {
			return this.ModelTenant.GetByUuidAndEnabled(tenantUuid)

		}
		tenant, _ = cache.Memoize(this.CacheService, tenantUuid, new(models.Tenant), loader)
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
				this.renderJsonValidationError("tenant not configured")
			} else {
				this.OnErrorAny("/", "user does not has active tenant")
			}
			return
		}

		if !tenant.Enabled && !services.IsRootUser(this.GetAuthUser()) {
			logs.Error("ERROR: tenant ", tenant.Id, " - ", tenant.Name, " is disabled")

			if this.IsTokenLoggedIn || this.IsJson() {
				this.renderJsonForbidenError("operation not permitted", false)
			} else {
				this.LogOut()
				this.OnErrorAny("/", "operation not permitted")
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

		this.Data["UserInfo"] = this.GetAuthUser()
		this.Data["Tenant"] = this.GetAuthTenant()

		//ioc.Get[services.AuthService](this.Container)
		this.Auth.SetUserInfo(this.GetAuthUser())
	}

	this.Data["IsLoggedIn"] = this.IsLoggedIn()

	if this.IsLoggedIn() {
		this.Data["IsAdmin"] = this.Auth.IsAdmin()
		this.Data["IsRoot"] = this.Auth.IsRoot()
	}

	this.UpSecurityAuth()
}

func (this *WebController) IsLoggedIn() bool {
	return this.IsWebLoggedIn || this.IsTokenLoggedIn || this.IsCustomAppLoggedIn
}

func (this *WebController) IsWebOrTokenLoggerIn() bool {
	return this.IsWebLoggedIn || this.IsTokenLoggedIn
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
	this.SetTemplate(viewName)
}

func (this *WebController) OnEntityFail(viewName string, entity interface{}, fail *optional.Fail) {
	this.Data["entity"] = entity

	if fail.Item != nil {
		switch fail.Item.(type) {
		case []map[string]string:
			errors := map[string]string{}
			for _, it := range fail.Item.([]map[string]string) {
				if _, ok := it["field"]; ok {
					errors[it["field"]] = it["message"]
				}
			}
			this.Data["errors"] = errors
		}
	}

	if strings.Contains(fail.ErrorString(), "validation") {
		this.Flash.Error(this.GetMessage("cadastros.validacao"))
	} else {
		this.Flash.Error(fail.ErrorString())
	}

	this.SetTemplate(viewName)
}

func (this *WebController) OnEntityError(viewName string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Rollback()
	this.Flash.Error(message)
	this.Data["entity"] = entity
	this.SetTemplate(viewName)
}

func (this *WebController) OnEntities(viewName string, entities interface{}) {
	this.Data["entities"] = entities
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *WebController) OnEntitiesWithTotalCount(viewName string, entities interface{}, totalCount int64) {
	this.Data["entities"] = entities
	this.Data["totalCount"] = totalCount
	this.SetTemplate(viewName)
}

func (this *WebController) OnResult(viewName string, result interface{}) {
	this.Data["result"] = result
	this.SetTemplate(viewName)
}

func (this *WebController) OnResults(viewName string, results interface{}) {
	this.Data["results"] = results
	this.SetTemplate(viewName)
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

func (this *WebController) SetData(values ...interface{}) *WebController {

	if len(values)%2 > 0 {
		panic("set data expect key pair values")
	}

	data := maps.Of[string, interface{}](values...)

	for k, v := range data {
		this.Data[k] = v
	}

	return this
}

func (this *WebController) OnResultsWithTotalCount(viewName string, results interface{}, totalCount int64) {
	this.Data["results"] = results
	this.Data["totalCount"] = totalCount
	this.SetTemplate(viewName)
}

func (this *WebController) RenderJsonResult(opt interface{}) {

	//logs.Debug("RenderJsonResult = %v type of %v", opt, reflect.TypeOf(opt).Kind())

	switch opt.(type) {
	case *optional.Some:
		someVal := opt.(*optional.Some).Item

		switch someVal.(type) {
		case *criteria.Page:
			page := someVal.(*criteria.Page)
			this.OnJsonResultsWithTotalCount(page.Data, page.Count())
			break
		case *optional.Ok:
			this.OnJson200()
			break
		default:

			if val, ok := criteria.TryExtractPageIfPegeOf(someVal); ok {
				this.RenderJsonResult(val)
				return
			}

			if optional.IsSlice(someVal) {
				this.OnJsonResults(someVal)
			} else {
				this.OnJsonResult(someVal)
			}
		}
		break
	case *optional.None:
		this.NotFoundAsJson()
		break
	case *optional.Fail:

		fail := opt.(*optional.Fail)
		err := opt.(*optional.Fail).Error

		if err.Error() == "validation error" {
			this.OnJsonValidationWithErrors(fail.Item.(map[string]string))
		} else {
			this.OnJsonError(fmt.Sprintf("%v", err))
		}
		break
	case *response.JsonResult:
		this.OnJson(opt.(*response.JsonResult))
		break
	case *criteria.Page:
		page := opt.(*criteria.Page)
		this.OnJsonResultsWithTotalCount(page.Data, page.Count())
		break
	case error:
		this.OnJsonError(fmt.Sprintf("%v", opt.(error).Error()))
		break

	default:

		if val, ok := optional.TryExtractValIfOptional(opt); ok {
			this.RenderJsonResult(val)
			return
		}

		if val, ok := criteria.TryExtractPageIfPegeOf(opt); ok {
			this.RenderJsonResult(val)
			return
		}

		if optional.IsSlice(opt) {
			logs.Debug("render as results")
			this.OnJsonResults(opt)
		} else {
			logs.Debug("render as result")
			this.OnJsonResult(opt)
		}

		//this.OnJsonError(fmt.Sprintf("unknow optional value: %v", opt))
		break
	}

}

func (this *WebController) NewRawJson(value interface{}) *response.RawJson {
	return &response.RawJson{value}
}

func (this *WebController) RenderJson(opt interface{}) {

	var dataResult interface{}
	var statusCodeResult = 200

	switch opt.(type) {
	case *criteria.Page:
		dataResult = opt
		break
	case *optional.Some:
		someVal := opt.(*optional.Some).Item

		switch someVal.(type) {
		case *criteria.Page:
			dataResult = someVal
			break
		case *optional.Ok:
			dataResult = map[string]interface{}{}
			break
		default:
			if val, ok := criteria.TryExtractPageIfPegeOf(someVal); ok {
				this.RenderJson(val)
				return
			}
			dataResult = map[string]interface{}{
				"data": someVal,
			}
		}

		break
	case *optional.None:
		statusCodeResult = 404
		dataResult = maps.JSON("error", true, "message", "not found")
		break
	case *response.RawJson:
		dataResult = opt.(*response.RawJson).Value
		break
	case *optional.Fail:

		f := opt.(*optional.Fail)
		err := f.Error
		statusCode := 500

		data := maps.JSON("error", true, "message", fmt.Sprintf("%v", err))

		switch err.(type) {
		case *validator.ValidationError:
			data["validation"] = err.(*validator.ValidationError).Map
			data["validations"] = err.(*validator.ValidationError).List
			statusCode = 400
			break
		default:
			if err.Error() == "validation error" {
				data["validation"] = f.Item
				statusCode = 400
			}
			break
		}
		dataResult = data
		statusCodeResult = statusCode
		break
	case *validator.ValidationError:
		data := maps.JSON("error", true, "message", fmt.Sprintf("%v", opt))
		data["validation"] = opt.(*validator.ValidationError).Map
		data["validations"] = opt.(*validator.ValidationError).List
		dataResult = data
		statusCodeResult = 400
		break
	case error:
		statusCodeResult = 500
		dataResult = maps.JSON("error", true, "message", fmt.Sprintf("%v", opt.(error).Error()))
		break
	default:

		if val, ok := optional.TryExtractValIfOptional(opt); ok {
			this.RenderJson(val)
			return
		}

		if val, ok := criteria.TryExtractPageIfPegeOf(opt); ok {
			this.RenderJson(val)
			return
		}

		if val, ok := opt.(result.IResult); ok {
			if val.HasError() {
				this.RenderJson(val.GetError())
			} else {
				this.RenderJson(val.GetValue())
			}
			return
		}

		dataResult = maps.JSON("data", opt)
		break
	}

	var je *json.JSON
	if this.NewJSON != nil {
		je = this.NewJSON()
	} else {
		je = json.NewJSON()
	}

	j, err := je.Encode(dataResult)

	//logs.Debug("JSON = %v", string(j))

	if err != nil {
		logs.Error("ERROR JSON ENCODE: %v", err)
		this.Ctx.Output.SetStatus(500)
		this.Ctx.Output.Body([]byte(fmt.Sprint(`{ "error": true, "message": "%v" }`, err.Error())))
	} else {
		logs.Trace("REPONSE STATUS CODE = %v", statusCodeResult)
		this.Ctx.Output.SetStatus(statusCodeResult)
		this.Ctx.Output.Body(j)
	}

	//this.ServeJSON()
}
func (this *WebController) OnJsonResultNil() {
	this.OnJsonResult(nil)
}

func (this *WebController) OnJsonResult(result interface{}) {
	this.Data["json"] = &response.JsonResult{
		Result:          result,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.ServeJSON()
}

func (this *WebController) GetJsonResult() (*response.JsonResult, bool) {
	if this.Data["json"] != nil {
		if j, ok := this.Data["json"].(*response.JsonResult); ok {
			return j, ok
		}
	}
	return nil, false
}

func (this *WebController) OnJsonMessage(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = &response.JsonResult{
		Message:         message,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultError(result interface{}, format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = &response.JsonResult{
		Result:          result,
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultWithMessage(result interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = &response.JsonResult{
		Result:          result,
		Error:           false,
		Message:         message,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.ServeJSON()
}

func (this *WebController) OnJsonResults(results interface{}) {
	this.Data["json"] = &response.JsonResult{
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultAndResults(result interface{}, results interface{}) {
	this.Data["json"] = &response.JsonResult{
		Result:          result,
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultsWithTotalCount(results interface{}, totalCount int64) {
	this.Data["json"] = &response.JsonResult{
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
		TotalCount:      totalCount,
	}
	this.ServeJSON()
}

func (this *WebController) OnJsonPage(page *criteria.Page) {
	this.Data["json"] = &response.JsonResult{
		Results:         page.Data,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
		TotalCount:      int64(page.TotalCount),
	}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultAndResultsWithTotalCount(result interface{}, results interface{}, totalCount int64) {
	this.Data["json"] = &response.JsonResult{
		Result:          result,
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
		TotalCount:      totalCount,
	}
	this.ServeJSON()
}

func (this *WebController) OnJsonResultsError(results interface{}, format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = &response.JsonResult{
		Results:         results,
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.ServeJSON()
}

func (this *WebController) RenderJsonWithStatusCode(status int, data maps.JsonData) {
	this.Data["json"] = data
	this.Ctx.Output.SetStatus(status)
	this.ServeJSON()
}

func (this *WebController) OnJson(json *response.JsonResult) {
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
	result := &response.JsonResult{
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.OnJson(result)
}

func (this *WebController) ServeJSON() {

	if this.CustonJsonEncoder != nil {
		result := this.Data["json"]
		bdata, err := this.CustonJsonEncoder(result)
		if err != nil {
			this.Data["json"] = &response.JsonResult{
				Message:         fmt.Sprintf("Error json.Encode: %v", err),
				Error:           true,
				CurrentUnixTime: this.GetCurrentTimeUnix(),
			}
			this.Controller.ServeJSON()
		} else {
			this.Ctx.Output.Header("Content-Type", "application/json")
			this.Ctx.Output.Body(bdata)
		}
	} else if this.UseJsonPackage {
		result := this.Data["json"]

		encoder := func(interface{}) ([]byte, error) {
			if this.JsonPackageAsCamelCase {
				return json.EncodeAsCamelCase(result)
			}
			return json.Encode(result)
		}

		bdata, err := encoder(result)
		if err != nil {
			this.Data["json"] = &response.JsonResult{
				Message: fmt.Sprintf("Error json.Encode: %v", err),
				Error:   true, CurrentUnixTime: this.GetCurrentTimeUnix(),
			}
			this.Controller.ServeJSON()
		} else {
			this.Ctx.Output.Header("Content-Type", "application/json")
			this.Ctx.Output.Body(bdata)
		}
	} else {
		this.Controller.ServeJSON()
	}
}

func (this *WebController) RenderTemplate(viewName string, data ...interface{}) {
	keyPars := maps.Of[string, interface{}](data...)
	for k, v := range keyPars {
		this.Data[k] = v
	}
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
	case *response.JsonResult:
		this.OnJson(data.(*response.JsonResult))
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
	this.OnJson(&response.JsonResult{Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonOk(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&response.JsonResult{Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJson200() {
	this.OnJson(&response.JsonResult{CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) NotFoundAsJson() {
	this.Data["json"] = maps.JSON("message", "not found")
	this.NotFound()
	this.ServeJSON()

}

func (this *WebController) OkAsJson(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&response.JsonResult{CurrentUnixTime: this.GetCurrentTimeUnix(), Message: message})
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
	this.OnJson(&response.JsonResult{Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonValidationWithErrors(errors map[string]string) {
	this.Rollback()
	this.OnJson(&response.JsonResult{Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonValidationMessageWithErrors(message string, errors map[string]string) {
	this.Rollback()
	this.OnJson(&response.JsonResult{Message: message, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonValidationWithResultAndMessageAndErrors(result interface{}, message string, errors map[string]string) {
	this.Rollback()
	this.OnJson(&response.JsonResult{Message: message, Result: result, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnJsonValidationWithResultsAndMessageAndErrors(results interface{}, message string, errors map[string]string) {
	this.Rollback()
	this.OnJson(&response.JsonResult{Message: message, Results: results, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebController) OnTemplate(viewName string) {
	this.SetTemplate(viewName)
}

func (this *WebController) SetTemplate(viewName string) {
	this.TplName = fmt.Sprintf("%s/%s.tpl", this.ViewPath, viewName)
	this.OnFlash(false)
}

func (this *WebController) OnTemplateWithData(viewName string, data map[string]interface{}) {
	if data != nil {
		for k, v := range data {
			this.Data[k] = v
		}
	}
	this.SetTemplate(viewName)
}

func (this *WebController) OnFullTemplate(tplName string) {
	this.TplName = fmt.Sprintf("%s.tpl", tplName)
	this.OnFlash(false)
}

func (this *WebController) OnPureTemplate(templateName string) {
	this.TplName = templateName
	this.OnFlash(false)
}

func (this *WebController) OnRedirect(action string, args ...interface{}) {
	this.OnFlash(true)
	if this.Ctx.Input.URL() == action {
		logs.Error("redirect to same URL")
		this.CustomAbort(500, "redirect to same URL")
	} else {
		this.Redirect(fmt.Sprintf(action, args...), 302)
	}
}

func (this *WebController) OnRedirectError(action string, format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.Flash.Error(message)
	this.OnFlash(true)
	if this.Ctx.Input.URL() == action {
		this.Abort("500")
	} else {
		this.Redirect(action, 302)
	}
}

func (this *WebController) OnRedirectSuccess(action string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Flash.Success(message)
	this.OnFlash(true)
	if this.Ctx.Input.URL() == action {
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

func (this *WebController) Validate(entity interface{}, f ...func(validator *validation.Validation)) bool {

	var fn func(validator *validation.Validation) = nil

	if len(f) > 0 {
		fn = f[0]
	}

	result, _ := this.EntityValidator.IsValid(entity, fn)

	if result.HasError {
		this.Flash.Error(this.GetMessage("cadastros.validacao"))
		this.EntityValidator.CopyErrorsToView(result, this.Data)
	}

	return result.HasError == false

}

// Deprecated: use Validate
func (this *WebController) OnValidate(entity interface{}, f func(validator *validation.Validation)) bool {
	return this.Validate(entity, f)
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

func (this *WebController) BodyAsJson() (*json.Json, error) {
	return json.NewFromBytes(this.RawBody())
}

func (this *WebController) BodyAsJsonResult() *result.Result[*json.Json] {
	return result.Try(this.BodyAsJson)
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
	return this.Ctx.Input.AcceptsJSON() || this.Ctx.Input.Header("Content-Type") == "application/json"
}

func (this *WebController) IsAjax() bool {
	return this.Ctx.Input.IsAjax()
}

func (this *WebController) GetHeaderToken() string {
	token := this.GetHeaderByName("X-Auth-Token")
	if len(token) == 0 {
		token = this.GetHeaderByName("Authorization")
	}
	return token
}

func (this *WebController) IsBearerToken() bool {
	return this.Auth.IsBearerToken(this.GetHeaderToken())
}

func (this *WebController) GetHeaderTenant() string {
	return this.GetHeaderByNames("tenant", "X-Auth-Tenant")
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
	filter := func(i int64) bool {
		return i > 0
	}

	if this.IsJson() {

		if this.Ctx.Input.IsPost() {
			jsonMap, _ := this.JsonToMap(this.Ctx)

			if _, ok := jsonMap["limit"]; ok {
				page.Limit = optional.
					New[int64](this.GetJsonInt64(jsonMap, "limit")).
					Filter(filter).
					GetOr(defaultLimit)

				page.Offset = this.GetJsonInt64(jsonMap, "offset")

				page.Sort = optional.
					New[string](this.GetJsonString(jsonMap, "order_column")).
					GetOr(this.GetJsonString(jsonMap, "sort"))

				page.Order = optional.
					New[string](this.GetJsonString(jsonMap, "order_sort")).
					GetOr(this.GetJsonString(jsonMap, "order"))

				page.Order = this.GetJsonString(jsonMap, "order_sort")
				page.Search = this.GetJsonString(jsonMap, "search")

				return page
			}
		}
	}

	page.Limit = optional.
		New[int64](this.GetIntByKey("limit")).
		Filter(filter).
		GetOr(defaultLimit)

	page.Sort = optional.
		New[string](this.GetStringByKey("order_column")).
		GetOr(this.GetStringByKey("sort"))

	page.Order = optional.
		New[string](this.GetStringByKey("order_sort")).
		GetOr(this.GetStringByKey("order"))

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

func (this *WebController) GetCheckbox(key string) bool {
	val := this.GetString(key)
	return strings.ToLower(val) == "on"
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
	return time.ParseInLocation(util.DateBrLayout, date, this.DefaultLocation)
}

// deprecated
func (this *WebController) ParseDateTime(date string) (time.Time, error) {
	return time.ParseInLocation(util.DateTimeBrLayout, date, this.DefaultLocation)
}

// deprecated
func (this *WebController) ParseJsonDate(date string) (time.Time, error) {
	return time.ParseInLocation(util.DateTimeDbLayout, date, this.DefaultLocation)
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

	if this.IsWebLoggedIn && !this.DoNotLoadTenantsOnSession {

		cacheKey := cache.CacheKey("tenants_user_", this.GetAuthUser().Id)
		this.DeleteCacheOnLogout(cacheKey)

		loader := func() ([]*models.Tenant, error) {

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

			return tenants, nil
		}

		authorizeds, _ := cache.Memoize(this.CacheService, cacheKey, new([]*models.Tenant), loader)

		this.Session.SetAuthorizedTenants(lists.MapToInterface(authorizeds))
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

	token := this.GetHeaderToken()

	if strings.TrimSpace(token) != "" {

		auth := services.NewLoginService(this.Lang, this.Session)

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

			user, err := cache.Memoize(this.CacheService, token, new(models.User), loader)

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

func (this *WebController) GetLogin() *models.User {
	id, _ := this.GetSession("userinfo").(int64)
	return this.memoizeUser(id)
}

func (this *WebController) GetTokenLogin() *models.User {
	id, _ := this.GetSession("appuserinfo").(int64)
	return this.memoizeUser(id)
}

func (this *WebController) GetCustomAppLogin() *models.User {
	id, _ := this.GetSession("customappuserinfo").(int64)
	return this.memoizeUser(id)
}

func (this *WebController) GetCustomAppName() string {
	name, _ := this.GetSession("customappname").(string)
	return name
}

func (this *WebController) GetCustomAppId() int64 {
	id, _ := this.GetSession("customappid").(int64)
	return id
}

func (this *WebController) GetCustomApp() (*models.App, error) {
	id := this.GetCustomAppId()
	return criteria.New[*models.App](this.Session).FindById(id)
}

func (this *WebController) SessionLogOut() {
	this.LogOut()
}

func (this *WebController) LogOut() {
	this.CacheService.Delete(this.CacheKeysDeleteOnLogOut...)
	this.DelSession("userinfo")
	this.DelSession("appuserinfo")
	this.DelSession("customappuserinfo")
	this.DelSession("authtenantid")
	this.DelSession("customappname")
	this.DelSession("customappid")
	this.DestroySession()
}

func (this *WebController) SetCustomAppName(appname string) {
	this.SetSession("customappname", appname)
}

func (this *WebController) SetCustomAppId(id int64) {
	this.SetSession("customappid", id)
}

func (this *WebController) SetLogin(user *models.User) {
	this.SetSession("userinfo", user.Id)
}

func (this *WebController) SetTokenLogin(user *models.User) {
	this.SetSession("appuserinfo", user.Id)
}

// set login from models.App (custom app login)
func (this *WebController) SetCustomAppLogin(user *models.User) {
	this.SetSession("customappuserinfo", user.Id)
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
	if !this.IsLoggedIn() {
		if this.IsJson() {
			msgKey := "security.notLoggedIn"
			if this.ExitWithHttpCode {
				msgKey = "security.unauthorized"
			}
			this.renderJsonUnauthorizedError(this.GetMessage(msgKey), true)
		} else {
			this.OnLoginRedirect()
		}
	}
}

func (this *WebController) AuthCheckRoot() {
	if !this.IsWebLoggedIn {
		if this.IsJson() {
			msgKey := "security.notLoggedIn"
			if this.ExitWithHttpCode {
				msgKey = "security.unauthorized"
			}
			this.renderJsonUnauthorizedError(this.GetMessage(msgKey), true)
		} else {
			this.OnLoginRedirect()
		}
	}

	if !this.Auth.IsRoot() {
		if this.IsJson() {
			this.renderJsonForbidenError(this.GetMessage("security.rootRequired"), true)
		} else {
			this.OnRedirect("/")
		}
	}
}

func (this *WebController) AuthCheckAdmin() {
	if !this.IsWebLoggedIn {
		if this.IsJson() {
			msgKey := "security.notLoggedIn"
			if this.ExitWithHttpCode {
				msgKey = "security.unauthorized"
			}
			this.renderJsonUnauthorizedError(this.GetMessage(msgKey), true)
		} else {
			this.OnLoginRedirect()
			this.OnRedirectError("/", this.GetMessage("security.rootRequired"))
		}
	}

	if !this.Auth.IsRoot() && !this.Auth.IsAdmin() {
		if this.IsJson() {
			this.renderJsonForbidenError(this.GetMessage("security.rootRequired"), true)
		} else {
			this.OnRedirectError("/", this.GetMessage("security.rootRequired"))
		}
	}
}

func (this *WebController) UpSecurityAuth() bool {

	roles := []string{}

	if this.Auth.IsAuthenticated() {
		roles = this.Auth.GetUserRoles()
	}

	if !route.IsRouteAuthorized(this.Ctx, roles) {

		logs.Warn("WARN: path %v not authorized ", this.Ctx.Input.URL())

		if !this.IsLoggedIn() {

			if this.IsBearerToken() {
				this.RenderJsonWithStatusCode(401,
					maps.JSON(
						"message", "unauthorized"))
			} else {
				if this.IsJson() {
					msgKey := "security.notLoggedIn"
					if this.ExitWithHttpCode {
						msgKey = "security.unauthorized"
					}
					this.renderJsonUnauthorizedError(
						this.GetMessage(msgKey), false)
				} else {
					this.OnLoginRedirect()
				}
			}
			return false
		}

		if this.IsBearerToken() {
			this.RenderJsonWithStatusCode(403,
				maps.JSON(
					"message", "forbidden"))
		} else {
			if this.IsJson() {
				this.renderJsonUnauthorizedError(this.GetMessage("security.denied"), false)
			} else {
				this.OnRedirectError("/", this.GetMessage("security.denied"))
			}
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

func (this *WebController) GetFileOpt(key string) *optional.Optional[*Multipart] {
	file, fileHeader, err := this.GetFile(key)

	if err != nil {
		if err == http.ErrMissingFile {
			return optional.OfNone[*Multipart]()
		}
		return optional.OfFail[*Multipart](err)
	}

	return optional.Of[*Multipart](&Multipart{File: &file, FileHeader: fileHeader, Key: key})
}

func (this *WebController) FlashError(msg string, args ...interface{}) *WebController {
	this.Flash.Error(msg, args...)
	return this
}

func (this *WebController) FlashSuccess(msg string, args ...interface{}) *WebController {
	this.Flash.Success(msg, args...)
	return this
}

func (this *WebController) FlashWarn(msg string, args ...interface{}) *WebController {
	this.Flash.Warning(msg, args...)
	return this
}

func (this *WebController) FlashNotice(msg string, args ...interface{}) *WebController {
	this.Flash.Notice(msg, args...)
	return this
}

func (this *WebController) RenderResponse(resp *response.Response) {

	if resp.HasTemplate() {
		resp.ConfigureFlash(this.Flash)
		this.TplName = resp.GetTemplate(this.ViewPath)
		this.Data["errors"] = resp.Errors
		this.Data["entity"] = resp.Entity
		this.Data["entities"] = resp.Entities
		this.Data["result"] = resp.Result
		this.Data["results"] = resp.Results
		this.OnFlash(false)
	} else {
		// is json result

		// use custom json render
		this.UseJsonPackage = resp.JsonPackage
		this.JsonPackageAsCamelCase = resp.JsonPackageAsCamelCase

		if resp.JsonResult {
			this.Data["json"] = resp.MkJsonResult()
		} else if resp.HasValue() {
			this.RenderJson(resp.Value)
		}

	}
}

func (this *WebController) PreRender(ret interface{}) {
	if app, ok := this.AppController.(beego.PreRender); ok {
		app.PreRender(ret)
	}
}

func (this *WebController) renderJsonValidationError(message string) {
	if this.ExitWithHttpCode {
		this.RenderJsonWithStatusCode(400, maps.JSON("message", message))
	} else {
		this.OnJsonError(message)
	}
}

// permission error
func (this *WebController) renderJsonForbidenError(message string, abort bool) {
	if this.ExitWithHttpCode {
		this.RenderJsonWithStatusCode(403, maps.JSON("message", message))
	} else {
		this.OnJsonError(message)
		if abort {
			this.Abort("403")
		}
	}
}

// logion error
func (this *WebController) renderJsonUnauthorizedError(message string, abort bool) {
	if this.ExitWithHttpCode {
		this.RenderJsonWithStatusCode(401, maps.JSON("message", message))
	} else {
		this.OnJsonError(message)
		if abort {
			this.Abort("401")
		}
	}
}
