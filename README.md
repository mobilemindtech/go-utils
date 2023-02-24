# go-utils
Go lang utils beego

### tools and utils
```

validator/
  cpf - CPF validator
  cnpj - CPF validator
  
support/
  document_converter - converter document base64 to file
  json_parser - json parser functions
  json_result - default json result model
  security - create and compare hash
  utils - utils
 
beego
  db
    page - db paginator
    session - beego orm wrapper
  filters
    filter_method - enable beego put support
  validator
    entity_validator - beegoo entity validator
    validator - beego validator defaults configure
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

// convert form to json map.. so you can use at form
// input(name="Id") 
// input(name="Name") 
// input(name="Adreess.Id") 
// input(name="Adreess.Street")
// input(name="Adreess.Country.Id")
// result of parse is a map like this
// json = { Id: "", Name: "", Adrress: { Id: "", Street:"", Country: { Id: "" } } }
jsonMap = parser.FormToJson(this.Ctx) 

// convert form to map to json to model.. so you can use at form
// input(name="Id") 
// input(name="Name") 
// input(name="Adreess.Id") 
// input(name="Adreess.Street")
// input(name="Adreess.Country.Id")
// result of parse is a json like this:
// json = { Id: "", Name: "", Adrress: { Id: "", Street:"", Country: { Id: "" } } }
// so the parse of json to model is did
err := jsonMap = parser.FormToModel(this.Ctx, model) 

// get map in map
parser.GetJsonObject(jsonMap, "Foo")

// get array of json map.. return []map[string]interface{}
arrayOfMap := parser.GetJsonArray(jsomMap, "fones")

// get int in map
parser.GetJsonInt(jsonMap, "Id")

// get int64 in map
parser.GetJsonInt64(jsonMap, "Id")

// get string in map
parser.GetJsonString(jsonMap, "Name")

// get bool in map
parser.GetJsonBool(jsonMap, "Enabled")

// get date in map
parser.GetJsonDate(jsonMap, "Date", dateLayout)

// convert body request to entity
entity := new(MyEntity)
err = parser.JsonToModel(entity)

```

## use utils
```
import "github.com/mobilemindtec/go-utils/support"

result := support.FilterNumber("aa13") // rerutn 13
result := support.IsEmpty("") // rerutn true

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


validator.SetDefaultMessages() // set default validator messages pt-br
validator.AddCnpjValidator() // add custom func to CPF validator
validator.AddCpfValidator() // add custom func to CNPJ validator
validator.AddRelationValidator() // // add custom func to valid relations.. uses IsPersisted method of db.Model

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
  
## Session

Model
-----

```
type Model interface {
  IsPersisted() bool
  TableName() string  
}

```

Page
-----

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

// set all default Relations (FK) values (implements db.Model)
// use tag model `goutils:"ignore_set_default;ignore_set_default_child"`
// ignore_set_default: ignore relation
// ignore_set_default_child: ignore relation chields (fields)
// by default set all relations default value
SetDefaults(reply interface{}) error

// remove all relation that has tag `goutils:"remove_cascade"`
// bydefault not remove the relation without tag
RemoveCascade(reply interface{}) error

// saver or update relations that has tag `goutils:"save_or_update_cascade"`
// bydefault not save or update the relation without tag
SaveOrUpdateCascade(reply interface{}) error

// load all relations calling db.LoadRelated when relation has tag `goutils:"eager"`
// bydefault not load the relation without tag
Eager(reply interface{}) error

// load all relations calling db.LoadRelated and ignore the tag `goutils:"eager"`
// but look for tag `goutils:"ignore_eager;ignore_eager_child"`
// ignore_eager: ignore relation
// ignore_eager_child: ignore relation chields (fields)
EagerForce(reply interface{}) error
```

### Criteria



### Optional


		func doAnything() interface{} {
		
		  p, err := // do stuff
		
		  if err != nil {
		    return optional.NewFail(fmt.Errorf("error"))
		  }
		
		  return optional.NewNone()
		}
		
		opt := doAnything()
		
		switch opt.(type) {
		  case optional.Some:
		    //optiona.GetItem(opt)
		    // do stuff
		  case optional.Nome:
		    // do stuff
		  case optional.Fail:
		    //GetFail(opt).ErrorString()
		    // do stuff
		}


## Pipeline

		render := func (v interface{}) {
			logs.Info("%v", v)
		}
		
		pipe := pipeline.New()
	
		pipe.
			ErrorHandler(func(f *optional.Fail){
				// do anything on error
			}).
			ExitHandler(func() {
				// do anything on exit none
			}).
			SuccessHandler(func(){
				// do anything on success
			}).
			Next(func() interface{} {
				return json.New(this.RawBody()) // next need return optional Some or Fail
			}).
			Next(func() interface{} {
				
				// get last result from pipe line ctx, or with next param
				// each result is saved on ctx with step index. eg.: $0, $1, $2, $3
				// Optionaly you can return optional.NewEmpty() to ignore step result
				j := pipeline.GetCtxPtr[json.Json](pipe, "$0") 
				return criteria.
					New[Customer](this.Session).
					Eq("Id", j.GetInt64("id")).
					Rx().
					First().
					Get() // next need return optional Some or Fail
			}).
			Next(func() interface{} {
				cupom := pipeline.GetCtxPtr[Cupom](pipe, "$1")
				return session.New[CupomPtr](this.Session).Remove(cupom).Val()
			}).
			Run()
		
With validation
---------------

		type Cupom = models.Cupom
		type CupomPtr = *Cupom
		type Cupons = []CupomPtr
		
		func (this *BeegoController) Save() {
		
			render := this.RenderJson
			pipe := pipeline.New()
			ev := validator.New()
		
			ev.AddCustom(func(v *validation.Validation) {
				m := pipeline.GetCtxPtr[Cupom](pipe, "$1")
		
				if m.PercentDiscount <= 0 && m.CashDiscount <= 0 {
					v.SetError("Desconto", "Entre com um desconto válido")
				}
			})
		
			pipe.
				ErrorHandler(render).
				ExitHandler(func() {
					render(optional.NewNone())
				}).
				SuccessHandler(func() {
					cupom := pipeline.GetCtxPtr[Cupom](pipe, "$1")
					render(optional.NewSome(cupom))
				}).
				Next(func() interface{} {
					return json.New(this.RawBody())
				}).
				Next(func(j interface{}) interface{} {
					opt := criteria.
						New[Cupom](this.Session).
						Eq("Id", j.(*json.Json).GetInt64("id")).
						Rx().
						First().
						Get()
					return optional.OrElseSome(opt, new(Cupom))
				}).
				Next(func(j interface{}, m interface{}) interface{} {
					p := json.NewParser[Cupom]().
						AddConverter(func(j *json.Json) {
							// any json converter rule
						})
					return p.ParseJsonInto(j.(*json.Json), m.(CupomPtr)).Val()
				}).
				Next(func() interface{} {
					cupom := pipeline.GetCtxPtr[Cupom](pipe, "$1")
					return ev.Validate(cupom)
				}).
				Next(func() interface{} {
					cupom := pipeline.GetCtxPtr[Cupom](pipe, "$1")				
					return session.New[CupomPtr](this.Session).Persist(cupom).Val()
				}).
				Run()
		}
		

## Foreatch	

		foreach.
			New[CupomPtr]().
			By(100).
			For(func(offset int) *optional.Optional[Cupons] {
				return criteria.
					New[Cupom](localSession).
					Eq("Enabled", true).
					Lt("LimitDate", util.DateNow().Format("2006-01-2")).
					SetOffset(int64(offset)).
					SetLimit(100).
					OptAll()
		
			}).
			EachOne(func(c CupomPtr) bool {
		
				c.Enabled = false
		
				session.
					New[CupomPtr](localSession).
					Update(c).
					Val()
		
				return true // next?
			}).
			Do() // or return to pipeline Next. Pipile line call Foreach Do
		

## Task		

		type Provider struct {
			// stuff
		}
		
		type Result struct {
			Message String
		}
		
		providers := []*Provider{ new(Provider), new(Provider), }
		
	
		worker := func worker(provider *Provider, tg *task.TaskGroup[*Provider, *Result]) {
			tg.AddResult(&Result{ Message: "result to collect each datasource item" })	
		}
		
		tg := task.
			New[*ProviderItem, *Result]().
			SetDataSource(providers).
			SetTask(worker).
			SetOnReceive(func(r *Result){
				// receive processed item
			}).
			Start() // wait task be completed
	

