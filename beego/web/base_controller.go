package web

import (
  "github.com/mobilemindtec/go-utils/beego/validator"
  "github.com/mobilemindtec/go-utils/beego/filters"
  "github.com/mobilemindtec/go-utils/beego/db"
  "github.com/mobilemindtec/go-utils/support"
  "github.com/astaxie/beego/validation"
  "github.com/astaxie/beego/orm"
  "github.com/astaxie/beego"
  "github.com/beego/i18n"
  "html/template"
  "strings"  
  "strconv"
  "time"
  "fmt" 
)

var (
  langTypes []string // Languages that are supported.  
  datetimeLayout = "02/01/2006 10:25:32"
  dateLayout = "02/01/2006"
  dateZero = "01/01/0001"
  jsonDateLayout = "2006-01-02T15:04:05-07:00"
)

type BaseController struct {
  EntityValidator *validator.EntityValidator
  beego.Controller
  Flash *beego.FlashData  
  Session *db.Session
  support.JsonParser
  ViewPath string
  Db orm.Ormer
  i18n.Locale

  defaultPageLimit int64     
}

type NestPreparer interface {
  NestPrepare()
}

type NestFinisher interface {
  NestFinish()
}

func init() {
  LoadIl8n()
  LoadFuncs(nil)
}

func LoadFuncs(controller *BaseController) {
  hasError := func(args map[string]string, key string) string{    
    if args[key] != "" {
      return "has-error"      
    }
    return ""
  }

  errorMsg := func(args map[string]string, key string) string{    
    return args[key]
  }

  currentYaer := func () string {
    return strconv.Itoa(time.Now().Year())
  }

  beego.AddFuncMap("has_error", hasError)
  beego.AddFuncMap("error_msg", errorMsg)  
  beego.AddFuncMap("current_yaer", currentYaer)
  beego.InsertFilter("*", beego.BeforeRouter, filters.FilterMethod) // enable put 
}

func LoadIl8n() {
  beego.AddFuncMap("i18n", i18n.Tr)
  beego.SetLevel(beego.LevelDebug)

  // Initialize language type list.
  langTypes = strings.Split(beego.AppConfig.String("lang_types"), "|")

  // Load locale files according to language types.
  for _, lang := range langTypes {    
    if err := i18n.SetMessage(lang, "conf/i18n/"+"locale_" + lang + ".ini"); err != nil {
      beego.Error("Fail to set message file:", err)
      return
    }  
  }  
}

// Prepare implemented Prepare() method for baseController.
// It's used for language option check and setting.
func (this *BaseController) NestPrepareBase () {
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


  beego.Trace("Accept-Language is " + al)
  // 2. Default language is English.
  if len(this.Lang) == 0 {
    this.Lang = "pt-BR"
  }
  
  this.Flash = beego.NewFlash()

  // Set template level language option.
  this.Data["Lang"] = this.Lang
  this.Data["xsrfdata"]= template.HTML(this.XSRFFormHTML())
  this.Data["dateLayout"] = dateLayout
  this.Data["datetimeLayout"] = datetimeLayout


  this.Session = db.NewSession()
  this.Db = this.Session.Open()

  this.FlashRead()

  this.EntityValidator = validator.NewEntityValidator(this.Lang, this.ViewPath)

  this.Log("use default time location America/Sao_Paulo")
  this.DefaultLocation, _ = time.LoadLocation("America/Sao_Paulo")
  
  this.defaultPageLimit = 25
}

func (this *BaseController) DisableXSRF(pathList []string) {

  for _, url := range pathList {
    if strings.HasPrefix(this.Ctx.Input.URL(), url) {
      this.EnableXSRF = false   
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
  
  this.Session.Close()

  if app, ok := this.AppController.(NestFinisher); ok {
    app.NestFinish()
  }  
}

func (this *BaseController) Finally(){
  this.Session.OnError().Close()
}

func (this *BaseController) Rollback() {
  this.Session.OnError()
}

func (this *BaseController) OnEntity(viewName string, entity interface{}) {  
  this.Data["entity"] = entity
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnEntityError(viewName string, entity interface{}, message string) {  
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
  this.Data["result"] = results
  this.OnTemplate(viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnJsonResult(result interface{}) {
  this.Data["json"] = support.JsonResult{ Result: result, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultError(result interface{}, message string) {
  this.Data["json"] = support.JsonResult{ Result: result, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultWithMessage(result interface{}, message string) {
  this.Data["json"] = support.JsonResult{ Result: result, Error: false, Message: message, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResults(results interface{}) {
  this.Data["json"] = support.JsonResult{ Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultsWithTotalCount(results interface{}, totalCount int64) {
  this.Data["json"] = support.JsonResult{ Results: results, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix(), TotalCount: totalCount }
  this.ServeJSON()
}

func (this *BaseController) OnJsonResultsError(results interface{}, message string) {
  this.Data["json"] = support.JsonResult{ Results: results, Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() }
  this.ServeJSON()
}

func (this *BaseController) OnJson(json support.JsonResult) {
  this.Data["json"] = json
  this.ServeJSON()
}

func (this *BaseController) OnJsonMap(jsonMap map[string]interface{}) {
  this.Data["json"] = jsonMap
  this.ServeJSON()
}

func (this *BaseController) OnJsonError(message string) {
  this.Rollback()
  this.OnJson(support.JsonResult{ Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OnJsonOk(message string) {
  this.OnJson(support.JsonResult{ Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OnJsonValidationError() {
  this.Rollback()
  errors := this.Data["errors"].(map[string]string)
  this.OnJson(support.JsonResult{  Message: this.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix() })
}

func (this *BaseController) OnTemplate(viewName string) {    
  this.TplName = fmt.Sprintf("%s/%s.tpl", this.ViewPath, viewName)
  this.OnFlash(false)
}

func (this *BaseController) OnRedirect(action string) {
  this.OnFlash(true)
  this.Redirect(action, 302)  
}

func (this *BaseController) OnRedirectError(action string, message string) {
  this.Rollback()
  this.Flash.Error(message)
  this.OnFlash(true)
  this.Redirect(action, 302)  
}

func (this *BaseController) OnRedirectSuccess(action string, message string) {
  this.Flash.Success(message)
  this.OnFlash(true)
  this.Redirect(action, 302)  
}

func (this *BaseController) OnFlash(store bool) {
  if store {
    this.Flash.Store(&this.Controller)    
  } else {
    this.Data["Flash"] = this.Flash.Data
    this.Data["flash"] = this.Flash.Data
  }  
}

func (this *BaseController) GetMessage(key string, args ...interface{}) string{
  return i18n.Tr(this.Lang, key, args)
}

func (this *BaseController) OnValidate(entity interface{}, plus func(validator *validation.Validation)) bool {
  
  result, _ := this.EntityValidator.IsValid(entity, plus)

  if result.HasError {
    this.Flash.Error(this.GetMessage("cadastros.validacao"))
    this.EntityValidator.CopyErrorsToView(result, this.Data)
  }
  
  return result.HasError == false
}

func (this *BaseController) OnParseForm(entity interface{}) {
  if err := this.ParseForm(entity); err != nil {
    beego.Error("## error on parse form ", err.Error())
    panic(err)
  }

  //beego.Debug(fmt.Sprintf("################################################"))
  //beego.Debug(fmt.Sprintf("## on parse form success %+v", entity))
  //beego.Debug(fmt.Sprintf("################################################"))
}

func (this *BaseController) OnJsonParseForm(entity interface{}) {
  if err := this.FormToModel(this.Ctx, entity)  ; err != nil {
    beego.Error("## error on parse form ", err.Error())
    panic(err)
  }

  //beego.Debug(fmt.Sprintf("################################################"))
  //beego.Debug(fmt.Sprintf("## on parse form success %+v", entity))
  //beego.Debug(fmt.Sprintf("################################################"))
}



func (this *BaseController) OnParseJson(entity interface{}) {

  if err := this.JsonToModel(this.Ctx, entity); err != nil {
    beego.Error("## error on parse json ", err.Error())
    panic(err)
  }

  //beego.Debug(fmt.Sprintf("################################################"))
  //beego.Debug(fmt.Sprintf("## on parse json success %+v", entity))
  //beego.Debug(fmt.Sprintf("################################################"))
}

func (this *BaseController) IsJson() bool{
  return this.Ctx.Request.Header.Get("Content-Type") == "application/json" || this.Ctx.Request.Header.Get("contentType") == "application/json"
}

func (this *BaseController) GetId() int64 {
  return this.GetIntParam(":id")
}

func (this *BaseController) GetIntParam(key string) int64 {
  id := this.Ctx.Input.Param(key)
  intid, _ := strconv.ParseInt(id, 10, 64)  
  return intid
}

func (this *BaseController) GetIntByKey(key string) int64{
  val := this.Ctx.Input.Query(key)
  intid, _ := strconv.ParseInt(val, 10, 64)  
  return intid
}

func (this *BaseController) GetBoolByKey(key string) bool{
  val := this.Ctx.Input.Query(key)
  boolean, _ := strconv.ParseBool(val)
  return boolean
}

func (this *BaseController) GetStringByKey(key string) string{
  return this.Ctx.Input.Query(key)
}

func (this *BaseController) GetDateByKey(key string) (time.Time, error){
  date := this.Ctx.Input.Query(key)
  return this.ParseDate(date)
}

func (this *BaseController) ParseDate(date string) (time.Time, error){  
  return time.ParseInLocation(dateLayout, date, this.DefaultLocation)
}

func (this *BaseController) ParseDateTime(date string) (time.Time, error){  
  return time.ParseInLocation(datetimeLayout, date, this.DefaultLocation)
}

func (this *BaseController) ParseJsonDate(date string) (time.Time, error){  
  return time.ParseInLocation(jsonDateLayout, date, this.DefaultLocation)
}

func (this *BaseController) GetToken() string{
  return this.Ctx.Request.Header.Get("X-Auth-Token")
}

func (this *BaseController) IsZeroDate(date time.Time) bool{
  return date.Format(dateLayout) == dateZero || date.IsZero()
}

func (this *BaseController) Log(format string, v ...interface{}) {
  beego.Debug(fmt.Sprintf(format, v...))
}

func (this *BaseController) GetCurrentTimeUnix() int64 {
  return this.GetCurrentTime().Unix()
}

func (this *BaseController) GetCurrentTime() time.Time {
  return time.Now().In(this.DefaultLocation)
}

func (this *BaseController) GetPage() *db.Page{  
  page := new(db.Page)

  if this.IsJson() {
    
    jsonMap, _ := this.JsonToMap(this.Ctx)
    
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
    page.Search = this.GetStringByKey("search")
    page.Order = this.GetStringByKey("order_sort")
    page.Sort = this.GetStringByKey("order_column")

  }
  
  if page.Limit <= 0 {
    page.Limit = this.defaultPageLimit
  }

  return page
}