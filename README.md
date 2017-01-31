# go-utils
Go lang utils beego

### tools and utils
```

validator/
  cpf - CPF validator
  cnpj - CPF validator
  
support/
  json_parser - json parser functions
  json_result - default json result model
  security - create and compare hash
 
beego
  db
    page - db paginator
    session - beego orm wrapper
  filters
    filter_method - enable beego put support
  validator
    entity_validator - beegoo entity validator
  web
    base_controller - beego controller base
    
```

## use document validator

```
import "github.com/mobilemindtec/go-utils/validator"

ok, err := cpf.IsValid("999.999.999-99")
ok, err :=  cnpj.IsValid("99.999.999/9999-99")

```

## use json parser

```
import "github.com/mobilemindtec/go-utils/support"

var jsonMap map[string]interface{},
var err error
parser := new(support.JsonParser)

// convert body request to string map
jsonMap, err = parser.JsonToMap(this.Ctx)

// get map in map
parser.GetJsonObject(jsonMap, "Foo")

// get int in map
parser.GetJsonInt(jsonMap, "Id")

// get int64 in map
parser.GetJsonInt64(jsonMap, "Id")

// get string in map
parser.GetJsonString(jsonMap, "Name")

// get date in map
parser.GetJsonDate(jsonMap, "Date", dateLayout)

// convert body request to entity
entity := new(MyEntity)
err = parser.JsonToModel(entity)

```

## use security

```
import "github.com/mobilemindtec/go-utils/support"

// create sha1 hash
hash := support.TextToSha1("password")

// create sha2 hash
hash := support.TextToSha256("password")

// compare hash sha1
ok := support.IsSomeHash(hash, "password")

// compare hash sha2
ok := support.IsSomeHashSha256(hash, "password")

```

## use entity validator

```
import "github.com/mobilemindtec/go-utils/beego/validator"

var result *validator.EntityValidatorResult
var err error

// lang - to select locale file
// viewPath - view name, to get field description in locale file
// code: this.GetMessage(fmt.Sprintf("%s.%s", this.ViewPath, err.Field))
// eg.: viewPath = User, field = Name.. so, location message is User.Name = "User Name"
// or set only field name in model
entityValidator = validator.NewEntityValidator(lang, viewPath)

// beego validation
result, err := entityValidator.IsValid(MyModel, func(validator *validation.Validation){
  // extra validation.. this func can be nil, so will only validate model
  if !mytest {
    validator.SetError("ModelColumnName", this.GetMessage("my.message"))
  }
})

// copy errors to view
if result.HasError {  
  this.EntityValidator.CopyErrorsToView(result, this.Data)
}


```

## use filter
```
// add at base controller init
import "github.com/mobilemindtec/go-utils/beego/filters"
beego.InsertFilter("*", beego.BeforeRouter, filters.FilterMethod) // enable put 
```

## use base controller
```
import "github.com/mobilemindtec/go-utils/beego/web"

type MyBaseController struct {
  web.BaseController  
}

func (this *BaseController) Prepare() {
  
  this.NestPrepareBase()
  
  // to disable XSRF
  urlsDisableXSRF := []string{
    "/api/auth_token",
  }
  this.DisableXSRF(urlsDisableXSRF)
  
  // implements web.NestPreparer in your controller
  if app, ok := this.AppController.(web.NestPreparer); ok {
    app.NestPrepare()
  }  
}

```
web.BaseController 

```
type BaseController struct {
  EntityValidator *validator.EntityValidator
  beego.Controller
  Flash *beego.FlashData  
  Session *db.Session
  support.JsonParser
  ViewPath string
  Db orm.Ormer
  i18n.Locale     
}

```
* Database transaction manager (open, commit, rollback, close)

### functions
```
OnEntity(viewName string, entity interface{})
OnEntityError(viewName string, entity interface{}, message string)
OnEntities(viewName string, entities interface{})
OnResult(viewName string, result interface{})
OnResults(viewName string, results interface{})
OnJsonResult(result interface{})
OnJsonResultWithMessage(result interface{}, message string)
OnJsonResults(results interface{})
OnJson(json support.JsonResult)
OnJsonMap(jsonMap map[string]interface{})
OnJsonError(message string)
OnJsonOk(message string)
OnJsonValidationError()
OnTemplate(viewName string)
OnRedirect(action string)
 OnRedirectError(action string, message string)
 OnRedirectSuccess(action string, message string)
 OnFlash(store bool)
 GetMessage(key string, args ...interface{}) string
 OnValidate(entity interface{}, plus func(validator *validation.Validation)) bool 
 OnParseForm(entity interface{})
 OnParseJson(entity interface{})
 IsJson() bool
 GetId() int64
 GetIntParam(key string) int64
 GetIntByKey(key string) int64
 GetStringByKey(key string) string
 GetDateByKey(key string) (time.Time, error)
 ParseDate(date string) (time.Time, error)
 ParseDateTime(date string) (time.Time, error)
 ParseJsonDate(date string) (time.Time, error)
 GetToken() string
 IsZeroDate(date time.Time) bool
 Log(format string, v ...interface{}) 
 GetLastUpdate() time.Time
```
  
## Session

Model

```
type Model interface {
  IsPersisted() bool
  TableName() string  
}


```

Page
```
type Page struct {
  Offset int64
  Limit int64
  Search string
  Order string
  Sort string   
}

AddtFilterDefaul(columnName string) *Page
AddFilter(columnName string, value interface{}) *Page
AddFilterAnd(columnName string, value interface{}) *Page
MakeDefaultSort() 

Eg:

page := &Page{ Offset: 0, Limit: 10, Search: 'john' }

page.AddFilterDefault("Name").MakeDefaultSort() // name like '%john%'

page.AddFilter("Name", "john") // // name = 'john'
page.AddFilter("Name__icontains", "john") // name like '%john%'

page.AddFilter("Name", "john").AddFilterColumn("Age", 10) // name = 'john' or age = '10'

page.AddFilterAnd("Name", "john").AddFilterAnd("Age", 10) // name = 'john' and age = '10'

```
functions - entity should be implemented Model

```
NewSession() *Session - create new session
NewSessionWithTenantId(tenantId int64) *Session - create new session with tenant filter
OnError() *Session - Set session error
Open() orm.Ormer - Open session
Close() - Close session (commit or rollback)
Begin() orm.Ormer - Init transaction
Commit() 
Rollback() 
Save(entity interface{}) error 
Update(entity interface{}) error
Remove(entity interface{}) error 
Load(entity interface{}) error 
Count(entity interface{}) (int64, error)
HasById(entity interface{}, id int64) (bool, error) 
FindById(entity interface{}, id int64) (interface{}, error) 
SaveOrUpdate(entity interface{}) error
List(entity interface{}, entities interface{}) error 
Page(entity interface{}, entities interface{}, page *Page) error 
Query(entity interface{}) (orm.QuerySeter, error) 
ToList(querySeter orm.QuerySeter, entities interface{}) error 
ToOne(querySeter orm.QuerySeter, entity interface{}) error 
ToPage(querySeter orm.QuerySeter, entities interface{}, page *Page) error 
```
