package web

import (
  "github.com/mobilemindtec/go-utils/beego/validator"
  "github.com/mobilemindtec/go-utils/app/services"
  "github.com/mobilemindtec/go-utils/app/models"
  "github.com/mobilemindtec/go-utils/app/route"
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/mobilemindtec/go-utils/json"
  beego "github.com/beego/beego/v2/server/web"
  "github.com/mobilemindtec/go-utils/support"
  "github.com/beego/beego/v2/core/validation"
  "github.com/beego/beego/v2/core/logs"
  "github.com/beego/i18n"
  "html/template"
  "runtime/debug"
  "strings"
  "strconv"
  "time"
  "fmt"
)



type WebController struct {
  
  beego.Controller
  support.JsonParser
  i18n.Locale
  
  EntityValidator *validator.EntityValidator
  Flash *beego.FlashData
  Session *db.Session
  ViewPath string  

  // models
  ModelAuditor *models.Auditor
  ModelCidade *models.Cidade
  ModelEstado *models.Estado
  ModelRole *models.Role
  ModelTenant *models.Tenant
  ModelUser *models.User
  ModelTenantUser *models.TenantUser
  ModelUserRole *models.UserRole  

  defaultPageLimit int64
  
  // auth
  userinfo *models.User
  tenant *models.Tenant

  IsLoggedIn  bool
  IsTokenLoggedIn  bool

  Auth *services.AuthService

  UseJsonPackage bool 

  InheritedController interface{}


}


func init() {
  LoadIl8n()
  LoadFuncs()
}

func (this *WebController) SetUseJsonPackage() *WebController{
  this.UseJsonPackage = true
  return this
}

func (this *WebController) loadLang(){
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

  this.EntityValidator = validator.NewEntityValidator(this.Lang, this.ViewPath)
  this.DefaultLocation, _ = time.LoadLocation("America/Sao_Paulo")
  this.defaultPageLimit = 25

  this.loadLang()

  this.Flash = beego.NewFlash()
  this.FlashRead()


  // Set template level language option.
  this.Data["Lang"] = this.Lang
  this.Data["xsrfdata"]= template.HTML(this.XSRFFormHTML())
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
    this.Log("ERROR: db.NewSession: %v", err)
    this.Abort("505")
  }else{
    this.Log("INFO: Session created successful")
  }  

  return session
}

func (this *WebController) AuthPrepare(){
  // login
  this.AppAuth()
  this.SetParams()

  this.IsLoggedIn = this.GetSession("userinfo") != nil
  this.IsTokenLoggedIn = this.GetSession("appuserinfo") != nil

  var tenant *models.Tenant
  tenantUuid := this.GetHeaderByNames("tenant", "X-Auth-Tenant")

  if len(tenantUuid) > 0 {
    ModelTenant := this.ModelTenant
    tenant, _ = ModelTenant.GetByUuidAndEnabled(tenantUuid)
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
        ModelTenantUser := this.ModelTenantUser
        tenant, _ = ModelTenantUser.GetFirstTenant(this.GetAuthUser())
      }

    }

    if tenant == nil || !tenant.IsPersisted() {
      tenant = this.GetAuthUser().Tenant
      this.Session.Load(tenant)
    }


    if tenant == nil || !tenant.IsPersisted() {
      
      this.Log("ERROR: user does not have active tenant")

      if this.IsTokenLoggedIn && !this.IsJson() {
        this.OnJsonError("set header tenant")
      } else {
        this.OnErrorAny("/", "user does not has active tenant")
      }
      return
    }

    this.SetAuthTenant(tenant)

    this.Log("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++")
    this.Log("* User Id = %v", this.GetAuthUser().Id)
    this.Log("* User Name = %v", this.GetAuthUser().Name)
    this.Log("* Tenant Id = %v", this.GetAuthTenant().Id)
    this.Log("* Tenant Name = %v", this.GetAuthTenant().Name)
    this.Log("* User Authority = %v", this.GetAuthUser().Role.Authority)
    this.Log("* User Roles = %v", this.GetAuthUser().GetAuthorities())
    this.Log("* User IsLoggedIn = %v", this.IsLoggedIn)
    this.Log("* User IsTokenLoggedIn = %v", this.IsTokenLoggedIn)
    this.Log("* User Auth Token = %v", this.GetToken())    
    this.Log("+++++++++++++++++++++++++++++++++++++++++++++++++++++++++")

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

  for _, url := range pathList {
    if strings.HasPrefix(this.Ctx.Input.URL(), url) {
      this.EnableXSRF = false
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

  this.Log("* Controller.Finish, Commit")

  this.Session.Close()

  if app, ok := this.AppController.(NestFinisher); ok {
    app.NestFinish()
  }
}

func (this *WebController) Finally(){

  this.Log("* Controller.Finally, Rollback")

  if this.Session != nil {
    this.Session.OnError().Close()
  }
}

func (this *WebController) Recover(info interface{}){
  /*
  this.Log("--------------- Controller.Recover ---------------")
  this.Log("INFO: %v", info)
  this.Log("STACKTRACE: %v", string(debug.Stack()))
  this.Log("--------------- Controller.Recover ---------------")
  */
  if app, ok := this.AppController.(NestRecover); ok {
    info := &RecoverInfo{ Error: fmt.Sprintf("%v", info), StackTrace: string(debug.Stack()) }
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

func (this *WebController) OnResultsWithTotalCount(viewName string, results interface{}, totalCount int64) {
  this.Data["results"] = results
  this.Data["totalCount"] = totalCount
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *WebController) OnJsonResult(result interface{}) {
  this.Data["json"] = &support.JsonResult{ Result: result, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *WebController) OnJsonMessage(format string, v ...interface{}) {
  message := fmt.Sprintf(format, v...)
  this.Data["json"] = &support.JsonResult{ Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *WebController) OnJsonResultError(result interface{}, format string, v ...interface{}) {
  this.Rollback()
  message := fmt.Sprintf(format, v...)
  this.Data["json"] = &support.JsonResult{ Result: result, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *WebController) OnJsonResultWithMessage(result interface{}, format string, v ...interface{}) {
  message := fmt.Sprintf(format, v...)
  this.Data["json"] = &support.JsonResult{ Result: result, Error: false, Message: message, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *WebController) OnJsonResults(results interface{}) {
  this.Data["json"] = &support.JsonResult{ Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *WebController) OnJsonResultAndResults(result interface{}, results interface{}) {
  this.Data["json"] = &support.JsonResult{ Result: result, Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *WebController) OnJsonResultsWithTotalCount(results interface{}, totalCount int64) {
  this.Data["json"] = &support.JsonResult{ Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount }
  this.ServeJSON()
}

func (this *WebController) OnJsonResultAndResultsWithTotalCount(result interface{}, results interface{}, totalCount int64) {
  this.Data["json"] = &support.JsonResult{ Result: result, Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount }
  this.ServeJSON()
}

func (this *WebController) OnJsonResultsError(results interface{}, format string, v ...interface{}) {
  this.Rollback()
  message := fmt.Sprintf(format, v...)
  this.Data["json"] = &support.JsonResult{ Results: results, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() }
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
  result := &support.JsonResult{ Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.OnJson(result)
}

func (this *WebController) ServeJSON(){
  if this.UseJsonPackage {
    result := this.Data["json"]
    bdata, err := json.Encode(result)
    if err != nil {
      this.Data["json"] = &support.JsonResult{ Message: fmt.Sprintf("Error json.Encode: %v", err), Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() }
      this.Controller.ServeJSON()
    } else {
      this.Ctx.Output.Header("Content-Type", "application/json")
      this.Ctx.Output.Body(bdata)      
    }    
  }else{
    this.Controller.ServeJSON()
  }
}


func (this *WebController) OnJsonErrorNotRollback(format string, v ...interface{}) {  
  message := fmt.Sprintf(format, v...)
  this.OnJson(&support.JsonResult{ Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *WebController) OnJsonOk(format string, v ...interface{}) {
  message := fmt.Sprintf(format, v...)
  this.OnJson(&support.JsonResult{ Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *WebController) OnJson200() {
  this.OnJson(&support.JsonResult{ CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *WebController) OkAsJson(format string, v ...interface{}) {
  message := fmt.Sprintf(format, v...)
  this.OnJson(&support.JsonResult{ CurrentUnixTime: this.GetCurrentTimeUnix(), Message: message })
}

func (this *WebController) OkAsHtml(message string) {
  this.Ctx.Output.Body([]byte(message))
}

func (this *WebController) Ok() {
  this.Ctx.Output.SetStatus(200)
}

func (this *WebController) OnJsonValidationError() {
  this.Rollback()
  errors := this.Data["errors"].(map[string]string)
  this.OnJson(&support.JsonResult{  Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *WebController) OnJsonValidationWithErrors(errors map[string]string) {
  this.Rollback()  
  this.OnJson(&support.JsonResult{  Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *WebController) OnJsonValidationMessageWithErrors(message string, errors map[string]string) {
  this.Rollback()  
  this.OnJson(&support.JsonResult{  Message: message, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix() })
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
  }}

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

func (this *WebController) GetMessage(key string, args ...interface{}) string{
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
    this.Log("*******************************************")
    this.Log("***** ERROR on parse form ", err.Error())
    this.Log("*******************************************")
    this.Abort("500")
  }
}

func (this *WebController) OnJsonParseForm(entity interface{}) {
  this.OnJsonParseFormWithFieldsConfigs(entity, nil)
}

func (this *WebController) OnJsonParseFormWithFieldsConfigs(entity interface{}, configs map[string]string) {
  if err := this.FormToModelWithFieldsConfigs(this.Ctx, entity, configs)  ; err != nil {
    this.Log("*******************************************")
    this.Log("***** ERROR on parse form ", err.Error())
    this.Log("*******************************************")
    this.Abort("500")
  }
}

func (this *WebController) ParamParseMoney(s string) float64{
  return this.ParamParseFloat(s)
}

// remove ,(virgula) do valor em params que vem como val de input com jquery money
// exemplo 45,000.00 vira 45000.00
func (this *WebController) ParamParseFloat(s string) float64{
  var semic string = ","
  replaced := strings.Replace(s, semic, "", -1) // troca , por espaço
  precoFloat, err := strconv.ParseFloat(replaced, 64)
  var returnValue float64
  if err == nil {
    returnValue = precoFloat
  }else{
    this.Log("*******************************************")
    this.Log("****** ERROR parse string to float64 for stringv", s)
    this.Log("*******************************************")
    this.Abort("500")
  }

  return returnValue
}

func (this *WebController) OnParseJson(entity interface{}) {
  if err := this.JsonToModel(this.Ctx, entity); err != nil {
    this.Log("*******************************************")
    this.Log("***** ERROR on parse json ", err.Error())
    this.Log("*******************************************")
    this.Abort("500")
  }
}

func (this *WebController) HasPath(paths ...string) bool{
  for _, it := range paths {
    if strings.HasPrefix(this.Ctx.Input.URL(), it){
      return true
    }
  }
  return false
}

func (this *WebController) IsJson() bool{
  return  this.Ctx.Input.AcceptsJSON()
}

func (this *WebController) IsAjax() bool{
  return  this.Ctx.Input.IsAjax()
}

func (this *WebController) GetToken() string{
  return this.GetHeaderByName("X-Auth-Token")
}

func (this *WebController) GetHeaderByName(name string) string{
  return this.Ctx.Request.Header.Get(name)
}

func (this *WebController) GetHeaderByNames(names ...string) string{

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

func (this *WebController) GetPage() *db.Page{
  page := new(db.Page)

  if this.IsJson() {

    if this.Ctx.Input.IsPost() {
      jsonMap, _ := this.JsonToMap(this.Ctx)

      //this.Log(" page jsonMap = %v", jsonMap)

      if _, ok := jsonMap["limit"]; ok {
        page.Limit = this.GetJsonInt64(jsonMap, "limit")
        page.Offset = this.GetJsonInt64(jsonMap, "offset")
        page.Sort = this.GetJsonString(jsonMap, "order_column")
        page.Order = this.GetJsonString(jsonMap, "order_sort")
        page.Search = this.GetJsonString(jsonMap, "search")
      }
    } else {

        page.Limit = this.GetIntByKey("limit")
        page.Offset = this.GetIntByKey("offset")
        page.Sort = this.GetStringByKey("order_column")
        page.Order = this.GetStringByKey("order_sort")
        page.Search = this.GetStringByKey("search")

    }

  } else {

    page.Limit = this.GetIntByKey("limit")
    page.Offset = this.GetIntByKey("offset")
    page.Search = this.GetStringByKey("search")
    page.Order = this.GetStringByKey("order_sort")
    page.Sort = this.GetStringByKey("order_column")

  }

  if page.Limit <= 0 {
    page.Limit = this.defaultPageLimit
  }

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
    key = fmt.Sprintf(":", key)
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

func (this *WebController) GetIntByKey(key string) int64{
  val := this.Ctx.Input.Query(key)
  intid, _ := strconv.ParseInt(val, 10, 64)
  return intid
}

func (this *WebController) GetBoolByKey(key string) bool{
  val := this.Ctx.Input.Query(key)
  boolean, _ := strconv.ParseBool(val)
  return boolean
}

func (this *WebController) GetStringByKey(key string) string{
  return this.Ctx.Input.Query(key)
}

func (this *WebController) GetDateByKey(key string) (time.Time, error){
  date := this.Ctx.Input.Query(key)
  return this.ParseDate(date)
}

func (this *WebController) ParseDateByKey(key string, layout string) (time.Time, error){
  date := this.Ctx.Input.Query(key)
  return time.ParseInLocation(layout, date, this.DefaultLocation)
}

// deprecated
func (this *WebController) ParseDate(date string) (time.Time, error){
  return time.ParseInLocation(dateLayout, date, this.DefaultLocation)
}

// deprecated
func (this *WebController) ParseDateTime(date string) (time.Time, error){
  return time.ParseInLocation(datetimeLayout, date, this.DefaultLocation)
}

// deprecated
func (this *WebController) ParseJsonDate(date string) (time.Time, error){
  return time.ParseInLocation(jsonDateLayout, date, this.DefaultLocation)
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


func (this *WebController) LoadTenants(){
  tenants := []*models.Tenant{}
  
  if this.IsLoggedIn {

    if this.Auth.IsRoot() {
      its, _ := this.ModelTenant.List()
      tenants = *its
    } else {
      list, _ := this.ModelTenantUser.ListByUser(this.GetAuthUser())


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
    this.Session.SetAuthorizedTenants(authorizeds)
  }

  this.Data["AvailableTenants"] = tenants
}


func (this *WebController) Audit(format string, v ...interface{}) {
  auditor := services.NewAuditorService(this.Session, this.Lang, this.GetAuditorInfo())
  auditor.OnAuditWithNewDbSession(format, v...)
}

func (this *WebController) GetAuditorInfo() *services.AuditorInfo{
  return &services.AuditorInfo{ Tenant: this.GetAuthTenant(), User: this.GetAuthUser() }
}

func (this *WebController) GetLastUpdate() time.Time{
  lastUpdateUnix, _ := this.GetInt64("lastUpdate")
  var lastUpdate time.Time

  if lastUpdateUnix > 0 {
    lastUpdate = time.Unix(lastUpdateUnix, 0).In(this.DefaultLocation)
  }

  return lastUpdate
}

func (this *WebController) AppAuth(){

  token := this.GetToken()

  if strings.TrimSpace(token) != "" {

    auth := services.NewLoginService(this.Lang, this.Session)

    this.Log("Authenticate by token %v", token)

    user, err := auth.AuthenticateToken(token)

    if err != nil {
      this.Log("LOGIN ERROR: %v", err)
      this.LogOut()
      return
    }

    if user == nil {
      this.Log("LOGIN ERROR: user not found!")
      this.LogOut()
      return      
    }

    if user != nil && user.Id > 0{
      this.ModelUser.LoadRelated(user)
      this.SetTokenLogin(user)
    }
  }
}

func (this *WebController) GetLogin() *models.User {
  id, _ := this.GetSession("userinfo").(int64)
  e, err := this.Session.FindById(new(models.User), id)
  if err != nil {
    return nil
  }
  user := e.(*models.User)
  this.ModelUser.LoadRelated(user)
  return user
}

func (this *WebController) GetTokenLogin() *models.User {
  id, _ := this.GetSession("appuserinfo").(int64)
  e, err := this.Session.FindById(new(models.User), id)
  if err != nil {
    return nil
  }
  user := e.(*models.User)
  this.ModelUser.LoadRelated(user)
  return user
}


func (this *WebController) LogOut() {
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

  for k, v := range  values {
    this.Data["Params"].(map[string]string)[k] = v[0]
  }
}

func (this *WebController) OnLoginRedirect() {
  path := this.Ctx.Input.URI()
  if !strings.Contains("?", path) {
    path = "?redirect=" + path
  }
  this.Ctx.Redirect(302, this.LoginPath() + path)
}

func (this *WebController) AuthCheck() {
  if !this.IsLoggedIn && !this.IsTokenLoggedIn {
    if this.IsJson(){
      this.OnJsonError(this.GetMessage("security.notLoggedIn"))
      this.Abort("401")
    } else  {
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

    this.Log("WARN: path %v not authorized ", this.Ctx.Input.URL())

    if !this.IsLoggedIn && !this.IsTokenLoggedIn {
      if this.IsJson(){
        this.OnJsonError(this.GetMessage("security.notLoggedIn"))
        //this.Abort("401")
      } else  {
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

func (this *WebController) HasTenantAuth(tenant *models.Tenant) bool{
  if !this.Auth.IsRoot() {

    
    item, _ := this.ModelTenantUser.FindByUserAndTenant(this.GetAuthUser(), tenant)

    return item != nil && item.IsPersisted()

  }
  return true
}

func (this *WebController) SetAuthTenantSession(tenant *models.Tenant) {

  if(this.HasTenantAuth(tenant)){
    this.Log("Set tenant session. user %v now is using tenant %v", this.GetAuthUser().Id, tenant.Id)
    this.SetSession("authtenantid", tenant.Id)
  } else {
    this.Log("Cannot set tenant session. user %v not enable to use tenant %v", this.GetAuthUser().Id, tenant.Id)
  }

}

func (this *WebController) GetAuthTenantSession() *models.Tenant {
  if id, ok := this.GetSession("authtenantid").(int64); ok {
    if id > 0 {
      tenant := models.Tenant{ Id: int64(id) }
      this.Session.Load(&tenant)
      return &tenant
    }
  }

  return nil
}

func (this *WebController) GetAuthTenant() *models.Tenant {

  tenant := this.GetAuthTenantSession()
  if tenant != nil && tenant.IsPersisted() {
    return tenant
  }

  return this.tenant
}

func (this *WebController) SetAuthTenant(t *models.Tenant ) {
  this.tenant = t
}

func (this *WebController) SetAuthUser(u *models.User)  {
  this.userinfo = u
}

func (this *WebController) GetAuthUser() *models.User {
  return this.userinfo
}
