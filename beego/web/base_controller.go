package web

import (
	"fmt"
	"github.com/mobilemindtec/go-utils/app/util"
	"html/template"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/core/validation"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/i18n"
	"github.com/mobilemindtec/go-utils/beego/db"
	"github.com/mobilemindtec/go-utils/beego/validator"
	"github.com/mobilemindtec/go-utils/beego/web/response"
	"github.com/mobilemindtec/go-utils/cache"
	"github.com/mobilemindtec/go-utils/json"
	"github.com/mobilemindtec/go-utils/support"
	"github.com/mobilemindtec/go-utils/v2/criteria"
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type BaseController struct {
	EntityValidator *validator.EntityValidator
	beego.Controller
	Flash   *beego.FlashData
	Session *db.Session
	support.JsonParser
	ViewPath string
	i18n.Locale

	defaultPageLimit int64

	CacheService            *cache.CacheService
	CacheKeysDeleteOnLogOut []string
}

func init() {
	LoadIl8n()
	LoadFuncs()
}

// Prepare implemented Prepare() method for baseController.
// It's used for language option check and setting.
func (this *BaseController) NestPrepareBase() {

	//this.Log("** web.BaseController.NestPrepareBase")

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

	//this.Log(" ** use language %v", this.Lang)

	this.Flash = beego.NewFlash()

	// Set template level language option.
	this.Data["Lang"] = this.Lang
	this.Data["xsrfdata"] = template.HTML(this.XSRFFormHTML())

	this.Data["dateLayout"] = util.DateBrLayout
	this.Data["datetimeLayout"] = util.DateTimeBrLayout
	this.Data["timeLayout"] = util.TimeMinutesLayout

	this.Session = db.NewSession()
	var err error
	err = this.Session.OpenTx()

	this.CacheService = cache.New()

	if err != nil {
		this.Log("***************************************************")
		this.Log("***************************************************")
		this.Log("***** erro ao iniciar conexão com banco de dados: %v", err)
		this.Log("***************************************************")
		this.Log("***************************************************")

		this.Abort("505")
		return
	}

	this.FlashRead()

	this.EntityValidator = validator.NewEntityValidator(this.Lang, this.ViewPath)

	//this.Log("use default time location America/Sao_Paulo")
	this.DefaultLocation, _ = time.LoadLocation("America/Sao_Paulo")

	this.defaultPageLimit = 25
}

func (this *BaseController) DisableXSRF(pathList []string) {

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

func (this *BaseController) FlashRead() {
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

func (this *BaseController) Finish() {

	this.Log("* Controller.Finish, Commit")

	this.Session.Close()

	if app, ok := this.AppController.(NestFinisher); ok {
		app.NestFinish()
	}
}

func (this *BaseController) Finally() {

	this.Log("* Controller.Finally, Rollback")

	if this.Session != nil {
		this.Session.OnError().Close()
	}
	this.CacheService.Close()
}

func (this *BaseController) Recover(info interface{}) {
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

func (this *BaseController) Rollback() {
	if this.Session != nil {
		this.Session.OnError()
	}
}

func (this *BaseController) OnEntity(viewName string, entity interface{}) {
	this.Data["entity"] = entity
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *BaseController) OnEntityError(viewName string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Rollback()
	this.Flash.Error(message)
	this.Data["entity"] = entity
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *BaseController) OnEntities(viewName string, entities interface{}) {
	this.Data["entities"] = entities
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *BaseController) OnEntitiesWithTotalCount(viewName string, entities interface{}, totalCount int64) {
	this.Data["entities"] = entities
	this.Data["totalCount"] = totalCount
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *BaseController) OnResult(viewName string, result interface{}) {
	this.Data["result"] = result
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *BaseController) OnResults(viewName string, results interface{}) {
	this.Data["results"] = results
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *BaseController) OnResultsWithTotalCount(viewName string, results interface{}, totalCount int64) {
	this.Data["results"] = results
	this.Data["totalCount"] = totalCount
	this.OnTemplate(viewName)
	this.OnFlash(false)
}

func (this *BaseController) RenderJsonResult(opt interface{}) {

	switch opt.(type) {
	case *optional.Some:

		if optional.IsOk(opt) {
			this.OnJson200()
		} else {

			it := opt.(*optional.Some).Item
			if optional.IsSlice(it) {
				this.OnJsonResults(it)
			} else {
				this.OnJsonResult(it)
			}
		}
		break
	case *optional.None:
		this.OnJson200()
		break
	case *optional.Fail:
		this.OnJsonError(fmt.Sprintf("%v", opt.(*optional.Fail).Error))
		break
	default:
		this.OnJsonError(fmt.Sprintf("unknow optional value: %v", opt))
		break
	}

}

func (this *BaseController) RenderJson(opt interface{}) {

	var dataResult interface{}
	var statusCodeResult = 200

	switch opt.(type) {
	case *optional.Some:

		someVal := opt.(*optional.Some).Item

		if optional.IsOk(someVal) {
			statusCodeResult = 404
		} else {

			switch someVal.(type) {
			case *criteria.Page:
				dataResult = someVal
				break
			default:
				dataResult = map[string]interface{}{
					"data": someVal,
				}
			}
		}

		break
	case *optional.None:
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
		this.Log("ERROR JSON ENCODE")
		this.Ctx.Output.SetStatus(500)
		this.Ctx.Output.Body([]byte(fmt.Sprint(`{ "error": true, "message": "%v" }`, err.Error())))
	} else {
		this.Log("STATUS = %v", statusCodeResult)
		this.Ctx.Output.SetStatus(statusCodeResult)
		this.Ctx.Output.Body(j)
	}

	this.ServeJSON()
}

func (this *BaseController) OnJsonResult(result interface{}) {
	this.Data["json"] = response.JsonResult{Result: result, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *BaseController) OnJsonMessage(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = response.JsonResult{Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *BaseController) OnJsonResultError(result interface{}, format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = response.JsonResult{Result: result, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *BaseController) OnJsonResultWithMessage(result interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = response.JsonResult{Result: result, Error: false, Message: message, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *BaseController) OnJsonResults(results interface{}) {
	this.Data["json"] = response.JsonResult{Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *BaseController) OnJsonResultAndResults(result interface{}, results interface{}) {
	this.Data["json"] = response.JsonResult{Result: result, Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *BaseController) OnJsonResultsWithTotalCount(results interface{}, totalCount int64) {
	this.Data["json"] = response.JsonResult{Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount}
	this.ServeJSON()
}

func (this *BaseController) OnJsonResultAndResultsWithTotalCount(result interface{}, results interface{}, totalCount int64) {
	this.Data["json"] = response.JsonResult{Result: result, Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount}
	this.ServeJSON()
}

func (this *BaseController) OnJsonResultsError(results interface{}, format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.Data["json"] = response.JsonResult{Results: results, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()}
	this.ServeJSON()
}

func (this *BaseController) OnJson(json response.JsonResult) {
	this.Data["json"] = json
	this.ServeJSON()
}

func (this *BaseController) OnJsonMap(jsonMap map[string]interface{}) {
	this.Data["json"] = jsonMap
	this.ServeJSON()
}

func (this *BaseController) OnJsonError(format string, v ...interface{}) {
	this.Rollback()
	message := fmt.Sprintf(format, v...)
	this.OnJson(response.JsonResult{Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *BaseController) OnJsonErrorNotRollback(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(response.JsonResult{Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *BaseController) OnJsonOk(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(response.JsonResult{Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *BaseController) OnJson200() {
	this.OnJson(response.JsonResult{CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *BaseController) OkAsJson(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(response.JsonResult{CurrentUnixTime: this.GetCurrentTimeUnix(), Message: message})
}

func (this *BaseController) OkAsHtml(message string) {
	this.Ctx.Output.Body([]byte(message))
}

func (this *BaseController) OkAsText(message string) {
	this.Ctx.Output.Body([]byte(message))
}

func (this *BaseController) Ok() {
	this.Ctx.Output.SetStatus(200)
}

func (this *BaseController) OnJsonValidationError() {
	this.Rollback()
	errors := this.Data["errors"].(map[string]string)
	this.OnJson(response.JsonResult{Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *BaseController) OnJsonValidationWithErrors(errors map[string]string) {
	this.Rollback()
	this.OnJson(response.JsonResult{Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *BaseController) OnJsonValidationWithResultAndMessageAndErrors(result interface{}, message string, errors map[string]string) {
	this.Rollback()
	this.OnJson(response.JsonResult{Message: message, Result: result, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *BaseController) OnJsonValidationWithResultsAndMessageAndErrors(results interface{}, message string, errors map[string]string) {
	this.Rollback()
	this.OnJson(response.JsonResult{Message: message, Results: results, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *BaseController) OnTemplate(viewName string) {
	this.TplName = fmt.Sprintf("%s/%s.tpl", this.ViewPath, viewName)
	this.OnFlash(false)
}

func (this *BaseController) OnPureTemplate(templateName string) {
	this.TplName = templateName
	this.OnFlash(false)
}

func (this *BaseController) OnRedirect(action string) {
	this.OnFlash(true)
	if this.Ctx.Input.URL() == "action" {
		this.Abort("500")
	} else {
		this.Redirect(action, 302)
	}
}

func (this *BaseController) OnRedirectError(action string, format string, v ...interface{}) {
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

func (this *BaseController) OnRedirectSuccess(action string, format string, v ...interface{}) {
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
func (this *BaseController) OnErrorAny(path string, format string, v ...interface{}) {

	//this.Log("** this.IsJson() %v", this.IsJson() )
	message := fmt.Sprintf(format, v...)
	if this.IsJson() {
		this.OnJsonError(message)
	} else {
		this.OnRedirectError(path, message)
	}
}

// executes redirect or OnJsonOk
func (this *BaseController) OnOkAny(path string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if this.IsJson() {
		this.OnJsonOk(message)
	} else {
		this.Flash.Success(message)
		this.OnRedirect(path)
	}

}

// executes OnEntity or OnJsonValidationError
func (this *BaseController) OnValidationErrorAny(view string, entity interface{}) {

	if this.IsJson() {
		this.OnJsonValidationError()
	} else {
		this.Rollback()
		this.OnEntity(view, entity)
	}

}

// executes OnEntity or OnJsonError
func (this *BaseController) OnEntityErrorAny(view string, entity interface{}, format string, v ...interface{}) {
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
func (this *BaseController) OnEntityAny(view string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if this.IsJson() {
		this.OnJsonResultWithMessage(entity, message)
	} else {
		this.Flash.Success(message)
		this.OnEntity(view, entity)
	}

}

// executes OnResults or OnJsonResults
func (this *BaseController) OnResultsAny(viewName string, results interface{}) {

	if this.IsJson() {
		this.OnJsonResults(results)
	} else {
		this.OnResults(viewName, results)
	}

}

// executes  OnResultsWithTotalCount or OnJsonResultsWithTotalCount
func (this *BaseController) OnResultsWithTotalCountAny(viewName string, results interface{}, totalCount int64) {

	if this.IsJson() {
		this.OnJsonResultsWithTotalCount(results, totalCount)
	} else {
		this.OnResultsWithTotalCount(viewName, results, totalCount)
	}

}

func (this *BaseController) OnFlash(store bool) {
	if store {
		this.Flash.Store(&this.Controller)
	} else {
		this.Data["Flash"] = this.Flash.Data
		this.Data["flash"] = this.Flash.Data
	}
}

func (this *BaseController) GetMessage(key string, args ...interface{}) string {
	return i18n.Tr(this.Lang, key, args)
}

func (this *BaseController) OnValidate(entity interface{}, custonValidation func(validator *validation.Validation)) bool {

	result, _ := this.EntityValidator.IsValid(entity, custonValidation)

	if result.HasError {
		this.Flash.Error(this.GetMessage("cadastros.validacao"))
		this.EntityValidator.CopyErrorsToView(result, this.Data)
	}

	return result.HasError == false
}

func (this *BaseController) OnParseForm(entity interface{}) {
	if err := this.ParseForm(entity); err != nil {
		this.Log("*******************************************")
		this.Log("***** ERROR on parse form ", err.Error())
		this.Log("*******************************************")
		this.Abort("500")
	}
}

func (this *BaseController) OnJsonParseForm(entity interface{}) {
	this.OnJsonParseFormWithFieldsConfigs(entity, nil)
}

func (this *BaseController) OnJsonParseFormWithFieldsConfigs(entity interface{}, configs map[string]string) {
	if err := this.FormToModelWithFieldsConfigs(this.Ctx, entity, configs); err != nil {
		this.Log("*******************************************")
		this.Log("***** ERROR on parse form ", err.Error())
		this.Log("*******************************************")
		this.Abort("500")
	}
}

func (this *BaseController) ParamParseMoney(s string) float64 {
	return this.ParamParseFloat(s)
}

// remove ,(virgula) do valor em params que vem como val de input com jquery money
// exemplo 45,000.00 vira 45000.00
func (this *BaseController) ParamParseFloat(s string) float64 {
	var semic string = ","
	replaced := strings.Replace(s, semic, "", -1) // troca , por espaço
	precoFloat, err := strconv.ParseFloat(replaced, 64)
	var returnValue float64
	if err == nil {
		returnValue = precoFloat
	} else {
		this.Log("*******************************************")
		this.Log("****** ERROR parse string to float64 for stringv", s)
		this.Log("*******************************************")
		this.Abort("500")
	}

	return returnValue
}

func (this *BaseController) OnParseJson(entity interface{}) {
	if err := this.JsonToModel(this.Ctx, entity); err != nil {
		this.Log("*******************************************")
		this.Log("***** ERROR on parse json ", err.Error())
		this.Log("*******************************************")
		this.Abort("500")
	}
}

func (this *BaseController) HasPath(paths ...string) bool {
	for _, it := range paths {
		if strings.HasPrefix(this.Ctx.Input.URL(), it) {
			return true
		}
	}
	return false
}

func (this *BaseController) IsJson() bool {
	return this.Ctx.Input.AcceptsJSON()
}

func (this *BaseController) IsAjax() bool {
	return this.Ctx.Input.IsAjax()
}

func (this *BaseController) GetToken() string {
	return this.GetHeaderByName("X-Auth-Token")
}

func (this *BaseController) GetHeaderByName(name string) string {
	return this.Ctx.Request.Header.Get(name)
}

func (this *BaseController) GetHeaderByNames(names ...string) string {

	for _, name := range names {
		val := this.Ctx.Request.Header.Get(name)

		if len(val) > 0 {
			return val
		}
	}

	return ""
}

func (this *BaseController) Log(format string, v ...interface{}) {
	logs.Debug(fmt.Sprintf(format, v...))
}

func (this *BaseController) GetCurrentTimeUnix() int64 {
	return this.GetCurrentTime().Unix()
}

func (this *BaseController) GetCurrentTime() time.Time {
	return time.Now().In(this.DefaultLocation)
}

func (this *BaseController) GetPage() *db.Page {
	return this.GetPageWithDefaultLimit(this.defaultPageLimit)
}

func (this *BaseController) GetPageWithDefaultLimit(defaultLimit int64) *db.Page {
	page := new(db.Page)

	if this.IsJson() {

		if this.Ctx.Input.IsPost() {
			jsonMap, _ := this.JsonToMap(this.Ctx)

			if _, ok := jsonMap["limit"]; ok {
				page.Limit = optional.
					New[int64](this.GetJsonInt64(jsonMap, "limit")).
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

func (this *BaseController) StringToInt(text string) int {
	val, _ := strconv.Atoi(text)
	return val
}

func (this *BaseController) StringToInt64(text string) int64 {
	val, _ := strconv.ParseInt(text, 10, 64)
	return val
}

func (this *BaseController) IntToString(val int) string {
	return fmt.Sprintf("%v", val)
}

func (this *BaseController) Int64ToString(val int64) string {
	return fmt.Sprintf("%v", val)
}

func (this *BaseController) GetId() int64 {
	return this.GetIntParam(":id")
}

func (this *BaseController) GetParam(key string) string {

	if !strings.HasPrefix(key, ":") {
		key = fmt.Sprintf(":", key)
	}

	return this.Ctx.Input.Param(key)
}

func (this *BaseController) GetStringParam(key string) string {
	return this.GetParam(key)
}

func (this *BaseController) GetIntParam(key string) int64 {
	id := this.GetParam(key)
	intid, _ := strconv.ParseInt(id, 10, 64)
	return intid
}

func (this *BaseController) GetInt32Param(key string) int {
	val := this.GetParam(key)
	intid, _ := strconv.Atoi(val)
	return intid
}

func (this *BaseController) GetBoolParam(key string) bool {
	val := this.GetParam(key)
	return val == "true"
}

func (this *BaseController) GetIntByKey(key string) int64 {
	val := this.Ctx.Input.Query(key)
	intid, _ := strconv.ParseInt(val, 10, 64)
	return intid
}

func (this *BaseController) GetBoolByKey(key string) bool {
	val := this.Ctx.Input.Query(key)
	boolean, _ := strconv.ParseBool(val)
	return boolean
}

func (this *BaseController) GetStringByKey(key string) string {
	return this.Ctx.Input.Query(key)
}

func (this *BaseController) GetDateByKey(key string) (time.Time, error) {
	date := this.Ctx.Input.Query(key)
	return this.ParseDate(date)
}

func (this *BaseController) ParseDateByKey(key string, layout string) (time.Time, error) {
	date := this.Ctx.Input.Query(key)
	return time.ParseInLocation(layout, date, this.DefaultLocation)
}

// deprecated
func (this *BaseController) ParseDate(date string) (time.Time, error) {
	return time.ParseInLocation(util.DateBrLayout, date, this.DefaultLocation)
}

// deprecated
func (this *BaseController) ParseDateTime(date string) (time.Time, error) {
	return time.ParseInLocation(util.DateTimeBrLayout, date, this.DefaultLocation)
}

// deprecated
func (this *BaseController) ParseJsonDate(date string) (time.Time, error) {
	return time.ParseInLocation(util.DateTimeDbLayout, date, this.DefaultLocation)
}

func (this *BaseController) RawBody() []byte {
	return this.Ctx.Input.RequestBody
}

func (this *BaseController) NotFound() {
	this.Ctx.Output.SetStatus(404)
}

func (this *BaseController) ServerError() {
	this.Ctx.Output.SetStatus(500)
}

func (this *BaseController) BadRequest(data interface{}) {
	this.Ctx.Output.SetStatus(400)
}

func (this *BaseController) DeleteCacheOnLogout(keys ...string) {
	this.CacheKeysDeleteOnLogOut = append(this.CacheKeysDeleteOnLogOut, keys...)
}

func (this *BaseController) LogoutHanlder() {
	this.CacheService.Delete(this.CacheKeysDeleteOnLogOut...)
}
