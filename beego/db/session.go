package db

import (
	"errors"
	"fmt"
	"reflect"
	"runtime/debug"
	"strings"

	"github.com/beego/beego/v2/client/orm"
	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type SessionState int

const (
	SessionStateOk SessionState = iota + 1
	SessionStateError
)

type Session struct {
	State                       SessionState
	Tenant                      interface{}
	AuthorizedTenants           []TenantModel
	IgnoreTenantFilter          bool
	IgnoreAuthorizedTenantCheck bool
	Debug                       bool
	DbName                      string

	deepSetDefault   map[string]int
	deepSaveOrUpdate map[string]int
	deepEager        map[string]int
	deepRemove       map[string]int

	openDbError bool

	database *DataBase
	tx       bool
}

func NewSession() *Session {
	return &Session{State: SessionStateOk, Debug: false, database: NewDataBase("default"), IgnoreAuthorizedTenantCheck: true}
}

func NewSessionWithDbName(dbName string) *Session {
	return &Session{State: SessionStateOk, Debug: false, database: NewDataBase(dbName), IgnoreAuthorizedTenantCheck: true}
}

func NewSessionWithTenant(tenant interface{}) *Session {
	return &Session{State: SessionStateOk, Tenant: tenant, Debug: false, database: NewDataBase("default"), IgnoreAuthorizedTenantCheck: true}
}

func NewSessionWithTenantAndDbName(tenant interface{}, dbName string) *Session {
	return &Session{State: SessionStateOk, Tenant: tenant, Debug: false, database: NewDataBase(dbName), IgnoreAuthorizedTenantCheck: true}
}

func RunTx(tenant interface{}, fn func(session *Session) error) error {
	session := NewSession().
		SetTenant(tenant)

	session.OpenTx()
	defer session.Close()

	return fn(session)
}

func RunNoTx(tenant interface{}, fn func(session *Session) error) error {
	session := NewSession().
		SetTenant(tenant)

	session.OpenNoTx()
	defer session.Close()

	return fn(session)
}

func (this *Session) SetDatabase(dt *DataBase) *Session {
	this.database = dt
	return this
}

func OrmVerbose(verbose bool) {
	orm.Debug = verbose
}

func (this *Session) GetDb() *DataBase {
	return this.database
}

func (this *Session) SetTenant(tenant interface{}) *Session {
	this.Tenant = tenant
	return this
}

func (this *Session) SetAuthorizedTenants(tenants []interface{}) *Session {

	this.AuthorizedTenants = []TenantModel{}

	for _, it := range tenants {
		if t, ok := it.(TenantModel); ok {
			this.AuthorizedTenants = append(this.AuthorizedTenants, t)
		}
	}

	return this
}

func (this *Session) SetNoAuthSession() *Session {
	this.IgnoreAuthorizedTenantCheck = true
	return this
}

func (this *Session) SetCheckAuthorizedTenant() *Session {
	this.IgnoreAuthorizedTenantCheck = false
	return this
}

func (this *Session) SetIgnoreTenantFilter() *Session {
	this.IgnoreTenantFilter = true
	return this
}

func (this *Session) SetUseTenantFilter(s bool) *Session {
	this.IgnoreTenantFilter = !s
	return this
}

func (this *Session) OnError() *Session {
	this.SetError()
	return this
}

func (this *Session) SetError() *Session {
	this.State = SessionStateError
	logs.Debug("DEBUG: Session.SetError")
	if this.Debug {
		logs.Debug("------------------------ DEBUG STACK TRACE BEGIN")
		debug.PrintStack()
		logs.Debug("------------------------ DEBUG STACK TRACE END")
	}
	return this
}

func (this *Session) RunWithTenant(tenant interface{}, runner func()) {

	tmp := this.Tenant
	this.Tenant = tenant

	defer func() {
		this.Tenant = tmp
	}()

	runner()
}

func (this *Session) WithTenant(tenant interface{}, runner func() error) error {

	tmp := this.Tenant
	this.Tenant = tenant

	defer func() {
		this.Tenant = tmp
	}()

	return runner()
}

func (this *Session) IsOpenDbError() bool {
	return this.openDbError
}

func (this *Session) Open(withTx bool) error {
	if withTx {
		return this.OpenWithTx()
	}
	return this.OpenWithoutTx()
}

func (this *Session) WithTx() (*Session, error) {
	return this, this.OpenTx()
}

func (this *Session) WithoutTx() (*Session, error) {
	return this, this.OpenNoTx()
}

func (this *Session) OpenTxOpt() interface{} {
	return optional.Make(this, this.OpenTx())
}

func (this *Session) OpenOpt() interface{} {
	return optional.Make(this, this.OpenNoTx())
}

func (this *Session) OpenTx() error {
	return this.OpenWithTx()
}

func (this *Session) OpenWithTx() error {
	this.tx = true
	return this.beginTx()
}

func (this *Session) OpenNoTx() error {
	return this.OpenWithoutTx()
}

func (this *Session) OpenWithoutTx() error {
	this.tx = false
	this.database.Open()
	return nil
}

func (this *Session) Close() {

	if this.tx {
		if this.State == SessionStateOk {
			this.Commit()
		} else {
			this.Rollback()
		}
	} else {
		this.database = nil
	}
}

func (this *Session) beginTx() error {

	if err := this.database.Begin(); err != nil {

		this.openDbError = true

		logs.Debug("****************************************************************")
		logs.Debug("****************************************************************")
		logs.Debug("****************************************************************")
		logs.Debug("************************ db begin error: %v", err.Error())
		logs.Debug("****************************************************************")
		logs.Debug("****************************************************************")
		logs.Debug("****************************************************************")

		return err
		//panic(err)
	}

	//logs.Debug("tx openned = %v", this.tx)

	return nil

}

func (this *Session) Commit() error {

	if this.Debug {
		logs.Debug("## session commit ")
	}

	if this.database != nil {
		if err := this.database.Commit(); err != nil {
			logs.Debug("****************************************************************")
			logs.Debug("****************************************************************")
			logs.Debug("****************************************************************")
			logs.Debug("************************ db commit error: %v", err.Error())
			logs.Debug("****************************************************************")
			logs.Debug("****************************************************************")
			logs.Debug("****************************************************************")
			this.Rollback()
			//panic(err)
			return err
		}
		this.database = nil
	}

	return nil
}

func (this *Session) Rollback() error {

	if this.Debug {
		logs.Debug("## session rollback ")
	}

	//logs.Debug("## session rollback ")

	logs.Debug("** Session Rollback ")

	if this.database != nil {
		if err := this.database.Rollback(); err != nil {
			logs.Debug("****************************************************************")
			logs.Debug("****************************************************************")
			logs.Debug("****************************************************************")
			logs.Debug("************************ db roolback error: %v", err.Error())
			logs.Debug("****************************************************************")
			logs.Debug("****************************************************************")
			logs.Debug("****************************************************************")
			return err
		}
		this.database = nil
	}

	return nil
}

func (this *Session) Save(entity interface{}) error {

	if !this.isTenantNil() && this.isSetTenant(entity) {
		if this.Debug {
			logs.Debug("## Save set tenant")
		}
		this.setTenant(entity)
	}

	if !this.checkIsAuthorizedTenant(entity, "Session.Save") {
		return errors.New("Tenant not authorized for entity data access. Operation: Session.Save.")
	}

	if hook, ok := entity.(ModelHookBeforeSave); ok {
		if err := hook.BeforeSave(); err != nil {
			return err
		}
	}

	_, err := this.GetDb().Insert(entity)

	if this.Debug {
		logs.Debug("## save data: %+v", entity)
	}

	if err != nil {
		logs.Debug("## Session: error on save: %v", err.Error())
		this.SetError()
		return err
	}

	if hook, ok := entity.(ModelHookAfterSave); ok {
		if err := hook.AfterSave(); err != nil {
			return err
		}
	}

	return nil
}

func (this *Session) Update(entity interface{}) error {

	if !this.isTenantNil() && this.isSetTenant(entity) {
		if this.Debug {
			logs.Debug("## Update set tenant")
		}
		this.setTenant(entity)

	}

	if !this.checkIsAuthorizedTenant(entity, "Session.Update") {
		return errors.New("Tenant not authorized for entity data access. Operation: Session.Update.")
	}

	if hook, ok := entity.(ModelHookBeforeUpdate); ok {
		if err := hook.BeforeUpdate(); err != nil {
			return err
		}
	}

	_, err := this.GetDb().Update(entity)

	if this.Debug {
		logs.Debug("## update data: %+v", entity)
	}

	if err != nil {
		logs.Debug("## Session: error on update: %v", err.Error())
		this.SetError()
		return err
	}

	if hook, ok := entity.(ModelHookAfterUpdate); ok {
		if err := hook.AfterUpdate(); err != nil {
			return err
		}
	}

	return nil
}

func (this *Session) Remove(entity interface{}) error {

	if !this.checkIsAuthorizedTenant(entity, "Session.Remove") {
		return errors.New("Tenant not authorized for entity data access. Operation: Session.Remove.")
	}

	if hook, ok := entity.(ModelHookBeforeRemove); ok {
		if err := hook.BeforeRemove(); err != nil {
			return err
		}
	}

	_, err := this.GetDb().Delete(entity)

	if err != nil {
		logs.Debug("## Session: error on remove: %v", err.Error())
		this.SetError()
		return err
	}

	if hook, ok := entity.(ModelHookAfterRemove); ok {
		if err := hook.AfterRemove(); err != nil {
			return err
		}
	}

	return nil
}

func (this *Session) Load(entity interface{}) (bool, error) {
	return this.Get(entity)
}

func (this *Session) Get(entity interface{}) (bool, error) {

	if model, ok := entity.(Model); ok {

		err := this.GetDb().Read(entity)

		if err == orm.ErrNoRows {
			//logs.Debug("## Session: error on load: %v", err.Error())
			//this.SetError()
			return false, nil
		}

		if model.IsPersisted() {

			if !this.checkIsAuthorizedTenant(model, "Session.Get") {
				return false, errors.New("Tenant not authorized for entity data access. Operation: Session.Get.")
			}

			return true, nil
		}

		if hook, ok := entity.(ModelHookAfterLoad); ok {
			if next, err := hook.AfterLoad(entity); !next || err != nil {

				if err != nil {
					return false, err
				}

				if !next {
					return false, nil
				}

			}
		}

		return false, nil

	}

	this.SetError()
	return false, errors.New("entity does not implements of Model")
}

func (this *Session) Count(entity interface{}) (int64, error) {

	if model, ok := entity.(Model); ok {

		query := this.GetDb().QueryTable(model.TableName())

		if !this.IgnoreTenantFilter {
			query = this.setTenantFilter(entity, query)
		}

		num, err := query.Count()

		if err != nil {
			logs.Debug("## Session: error on count table %v: %v", model.TableName(), err.Error())
			//this.SetError()
		}
		return num, err
	}

	this.SetError()
	return 0, errors.New("entity does not implements of Model")
}

func (this *Session) HasById(entity interface{}, id int64) (bool, error) {

	if model, ok := entity.(Model); ok {
		query := this.GetDb().QueryTable(model.TableName()).Filter("id", id)

		if !this.IgnoreTenantFilter {
			query = this.setTenantFilter(entity, query)
		}

		return query.Exist(), nil
	}

	this.SetError()
	return false, errors.New("entity does not implements of Model")
}

func (this *Session) FindById(entity interface{}, id int64) (interface{}, error) {

	if model, ok := entity.(Model); ok {
		query := this.GetDb().QueryTable(model.TableName()).Filter("id", id)

		if !this.IgnoreTenantFilter {
			query = this.setTenantFilter(entity, query)
		}

		err := query.One(entity)

		if err == orm.ErrNoRows {
			return nil, nil
		}

		if err != nil {
			logs.Debug("## Session: error on find by id table %v: %v", model.TableName(), err.Error())
			//this.SetError()
			return nil, err
		}

		if !model.IsPersisted() {
			return nil, nil
		}

		if hook, ok := entity.(ModelHookAfterLoad); ok {
			if next, err := hook.AfterLoad(entity); !next || err != nil {

				if err != nil {
					return nil, err
				}

				if !next {
					return nil, nil
				}

			}
		}

		return entity, nil
	}

	this.SetError()
	return false, errors.New("entity does not implements of Model")
}

func (this *Session) SaveOrUpdate(entity interface{}) error {

	if !this.isTenantNil() && this.isSetTenant(entity) {
		if this.Debug {
			logs.Debug("## SaveOrUpdate set tenant")
		}
		this.setTenant(entity)
	}

	if !this.checkIsAuthorizedTenant(entity, "Session.SaveOrUpdate") {
		return errors.New("Tenant not authorized for entity data access. Operation: Session.SaveOrUpdate.")
	}

	if model, ok := entity.(Model); ok {
		if model.IsPersisted() {
			if err := this.Update(entity); err != nil {
				return errors.New(fmt.Sprintf("error on update %v: %v", model.TableName(), err))
			}
			return nil
		}
		if err := this.Save(entity); err != nil {
			return errors.New(fmt.Sprintf("error on save %v: %v", model.TableName(), err))
		}
		return nil
	}

	this.SetError()
	return errors.New(fmt.Sprintf("entity %v does not implements of Model", entity))
}

func (this *Session) List(entity interface{}, entities interface{}) error {
	if model, ok := entity.(Model); ok {

		query := this.GetDb().QueryTable(model.TableName())

		if !this.IgnoreTenantFilter {
			query = this.setTenantFilter(entity, query)
		}

		if hook, ok := entity.(ModelHookBeforeQuery); ok {
			query = hook.BeforeQuery(query)
		}

		if _, err := query.All(entities); err != nil {
			logs.Debug("## Session: error on list: %v", err.Error())
			//this.SetError()
			return err
		}
		return nil
	}

	if hook, ok := entity.(ModelHookAfterList); ok {
		hook.AfterList(entities)
	}

	this.SetError()
	return errors.New("entity does not implements of Model 1")
}

func (this *Session) Page(entity interface{}, entities interface{}, page *Page) error {
	return this.PageQuery(nil, entity, entities, page)
}

func (this *Session) PageQuery(query orm.QuerySeter, entity interface{}, entities interface{}, page *Page) error {
	if model, ok := entity.(Model); ok {

		if query == nil {
			query = this.GetDb().QueryTable(model.TableName())
		}

		query = query.Limit(page.Limit).Offset(page.Offset)

		switch page.Sort {
		case "asc":
			query = query.OrderBy(fmt.Sprintf("%v", page.Sort))
		case "desc":
			query = query.OrderBy(fmt.Sprintf("-%v", page.Sort))
		}

		if page.FilterColumns != nil && len(page.FilterColumns) > 0 {

			if len(page.FilterColumns) == 1 {
				for k, v := range page.FilterColumns {
					query = query.Filter(k, v)
				}
			} else {
				cond := orm.NewCondition()
				for k, v := range page.FilterColumns {
					cond = cond.Or(k, v)
				}
				query = query.SetCond(orm.NewCondition().AndCond(cond))
			}

		}

		if page.AndFilterColumns != nil && len(page.AndFilterColumns) > 0 {
			for k, v := range page.AndFilterColumns {
				query = query.Filter(k, v)
			}
		}

		if !this.IgnoreTenantFilter {
			query = this.setTenantFilter(entity, query)
		}

		if hook, ok := entity.(ModelHookBeforeQuery); ok {
			query = hook.BeforeQuery(query)
		}

		if _, err := query.All(entities); err != nil {
			logs.Debug("## Session: error on page: %v", err.Error())
			//this.SetError()
			return err
		}

		if hook, ok := entity.(ModelHookAfterList); ok {
			hook.AfterList(entities)
		}

		return nil
	}

	this.SetError()
	return errors.New("entity does not implements of Model")
}

func (this *Session) Query(entity interface{}) (orm.QuerySeter, error) {
	if model, ok := entity.(Model); ok {
		query := this.GetDb().QueryTable(model.TableName())

		if !this.IgnoreTenantFilter {
			query = this.setTenantFilter(entity, query)
		}

		return query, nil
	}

	this.SetError()
	return nil, errors.New("entity does not implements of Model")
}

func (this *Session) ToList(querySeter orm.QuerySeter, entities interface{}) error {
	if _, err := querySeter.All(entities); err != nil {
		logs.Debug("## Session: error on to list: %v", err.Error())
		//this.SetError()
		return err
	}
	return nil
}

func (this *Session) ToOne(querySeter orm.QuerySeter, entity interface{}) error {
	if err := querySeter.One(entity); err != nil {
		if err != orm.ErrNoRows {
			logs.Debug("## Session: error on to one: %v", err.Error())
			return err
		}
		//this.SetError()
	}
	return nil
}

func (this *Session) ToPage(querySeter orm.QuerySeter, entities interface{}, page *Page) error {
	querySeter.Limit(page.Limit)
	querySeter.Offset(page.Offset)
	if _, err := querySeter.All(entities); err != nil {
		logs.Debug("## Session: error on to page: %v", err.Error())
		//this.SetError()
		return err
	}
	return nil
}

func (this *Session) ToCount(querySeter orm.QuerySeter) (int64, error) {

	var count int64
	var err error

	if count, err = querySeter.Count(); err != nil {
		logs.Debug("## Session: error on to count: %v", err.Error())
		//this.SetError()
		return count, err
	}
	return count, err
}

func (this *Session) ExecuteDelete(querySeter orm.QuerySeter) (int64, error) {

	var count int64
	var err error

	if count, err = querySeter.Delete(); err != nil {
		logs.Debug("## Session: error on to list: %v", err.Error())
		//this.SetError()
		return count, err
	}

	return count, err
}

func (this *Session) ExecuteUpdate(querySeter orm.QuerySeter, args map[string]interface{}) (int64, error) {

	var count int64
	var err error
	var params orm.Params

	for k, v := range args {
		params[k] = v
	}

	if _, err := querySeter.Update(params); err != nil {
		logs.Debug("## Session: error on to list: %v", err.Error())
		//this.SetError()
		return count, err
	}

	return count, err
}

func (this *Session) Eager(reply interface{}) error {
	this.deepEager = map[string]int{}
	return this.eagerDeep(reply, false)
}

func (this *Session) EagerForce(reply interface{}) error {
	this.deepEager = map[string]int{}
	return this.eagerDeep(reply, true)
}

func (this *Session) eagerDeep(reply interface{}, ignoreTag bool) error {

	if reply == nil {
		if this.Debug {
			logs.Debug("## reply is nil")
		}
		return nil
	}

	// value e type of pointer
	refValue := reflect.ValueOf(reply)
	//refType := reflect.TypeOf(reply)

	// value e type of instance
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)

		fieldStruct := fullValue.FieldByName(field.Name)
		fieldValue := fieldStruct.Interface()
		//fieldType := fieldStruct.Type()

		tags := this.getTags(field)

		if !ignoreTag {
			if tags == nil || len(tags) == 0 || !this.hasTag(tags, "eager") {
				continue
			}
		}

		if tags != nil && this.hasTag(tags, "ignore_eager") {
			continue
		}

		zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
		model, ok := fieldValue.(Model)

		if zero {

			if this.Debug {
				logs.Debug("## no eager zero field: ", field.Name)
			}

		} else if !ok {

			if this.Debug {
				logs.Debug("## no eager. field does not implemente model: ", field.Name)
			}

		} else {

			if model.IsPersisted() {

				if this.Debug {
					logs.Debug("## eager field: ", field.Name, fieldValue)
				}

				if _, err := this.GetDb().LoadRelated(reply, field.Name); err != nil {
					logs.Debug("********* eager field error ", fullType, field.Name, fieldValue, err.Error())
				} else {
					// reload loaded value of field reference
					refValue = reflect.ValueOf(reply)
					fullValue = refValue.Elem()
					fullType = fullValue.Type()
					field = fullType.Field(i)
					fieldStruct = fullValue.FieldByName(field.Name)
					fieldValue = fieldStruct.Interface()

					key := fmt.Sprintf("%v.%v", strings.Split(fullType.String(), ".")[1], field.Name)
					if count, ok := this.deepEager[key]; ok {

						if count >= 5 {
							continue
						}

						this.deepEager[key] = count + 1
					} else {
						this.deepEager[key] = 1
					}

					if this.Debug {
						logs.Debug("## eager field success: ", field.Name, fieldValue)
					}

				}
			} else {
				if this.Debug {
					logs.Debug("## not eager field not persisted: ", field.Name)
				}
			}

			if tags != nil && this.hasTag(tags, "ignore_eager_child") {
				continue
			}

			if this.Debug {
				logs.Debug("## eager next field: ", field.Name)
			}

			if err := this.eagerDeep(fieldValue, ignoreTag); err != nil {
				logs.Debug("## eager next field %v: %v", field.Name, err.Error())
				return err
			}
		}

	}

	return nil
}

func (this *Session) SaveCascade(reply interface{}) error {
	this.deepSaveOrUpdate = map[string]int{}
	return this.saveOrUpdateCascadeDeep(reply, true)
}

func (this *Session) SaveOrUpdateCascade(reply interface{}) error {
	this.deepSaveOrUpdate = map[string]int{}
	return this.saveOrUpdateCascadeDeep(reply, true)
}

func (this *Session) saveOrUpdateCascadeDeep(reply interface{}, firstTime bool) error {

	// value e type of pointer
	refValue := reflect.ValueOf(reply)
	//refType := reflect.TypeOf(reply)

	ignoreAuthorizedTenantCheckBkp := this.IgnoreAuthorizedTenantCheck

	// verifica apenas na entidade principal
	// nas filhas, ignora verificação
	if firstTime {
		defer func() { // em caso de erro, restaura confiuração
			//logs.Debug("DEFER IgnoreAuthorizedTenantCheck value")
			this.IgnoreAuthorizedTenantCheck = ignoreAuthorizedTenantCheckBkp
		}()
		this.IgnoreAuthorizedTenantCheck = true
	}

	// value e type of instance
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)

		fieldStruct := fullValue.FieldByName(field.Name)
		fieldValue := fieldStruct.Interface()
		//fieldType := fieldStruct.Type()

		tags := this.getTags(field)

		if tags == nil || len(tags) == 0 {
			continue
		}

		if tags == nil || !this.hasTag(tags, "save_or_update_cascade") {
			continue
		}

		zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
		_, ok := fieldValue.(Model)

		if zero {

			if this.Debug {
				logs.Debug("## no cascade zero field: ", field.Name)
			}

		} else if !ok {

			if this.Debug {
				logs.Debug("## no cascade. field does not implemente model: ", field.Name)
			}

		} else {

			if this.Debug {
				logs.Debug("## cascade field: ", field.Name)
			}

			key := fmt.Sprintf("%v.%v", strings.Split(fullType.String(), ".")[1], field.Name)
			if count, ok := this.deepSaveOrUpdate[key]; ok {

				if count >= 5 {
					continue
				}

				this.deepSaveOrUpdate[key] = count + 1
			} else {
				this.deepSaveOrUpdate[key] = 1
			}

			if err := this.saveOrUpdateCascadeDeep(fieldValue, false); err != nil {
				return err
			}
		}

	}

	if this.Debug {
		logs.Debug("## save or update: ", fullType)
	}

	// Volta a configuração para verificação da entidade principal.
	// No SaveOrUpdate é verificado
	if firstTime {
		this.IgnoreAuthorizedTenantCheck = ignoreAuthorizedTenantCheckBkp
	}

	return this.SaveOrUpdate(reply)
}

func (this *Session) RemoveCascade(reply interface{}) error {
	this.deepRemove = map[string]int{}
	return this.RemoveCascadeDeep(reply, true)
}

func (this *Session) RemoveCascadeDeep(reply interface{}, firstTime bool) error {

	// value e type of pointer
	refValue := reflect.ValueOf(reply)
	//refType := reflect.TypeOf(reply)

	// value e type of instance
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	var itensToRemove []interface{}

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)

		fieldStruct := fullValue.FieldByName(field.Name)
		fieldValue := fieldStruct.Interface()
		//fieldType := fieldStruct.Type()

		tags := this.getTags(field)

		if tags == nil || len(tags) == 0 {
			continue
		}

		if tags == nil || !this.hasTag(tags, "remove_cascade") {
			continue
		}

		zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
		_, ok := fieldValue.(Model)

		if zero {

			if this.Debug {
				logs.Debug("## no cascade remove zero field: ", field.Name)
			}

		} else if !ok {

			if this.Debug {
				logs.Debug("## no cascade remove. field does not implemente model: ", field.Name)
			}

		} else {

			//fieldFullType := reflect.TypeOf(fieldValue).Elem()

			key := fmt.Sprintf("%v.%v", strings.Split(fullType.String(), ".")[1], field.Name)

			if count, ok := this.deepRemove[key]; ok {

				if count >= 5 {
					continue
				}

				this.deepRemove[key] = count + 1
			} else {
				this.deepRemove[key] = 1
			}

			if this.Debug {
				logs.Debug("## cascade remove field: ", field.Name)
			}

			itensToRemove = append(itensToRemove, fieldValue)
		}

	}

	if this.Debug {
		logs.Debug("## remove: ", fullType)
	}

	if err := this.Remove(reply); err != nil {
		return err
	}

	for _, it := range itensToRemove {
		if err := this.RemoveCascadeDeep(it, false); err != nil {
			return err
		}
	}

	return nil
}

func (this *Session) SetDefaults(reply interface{}) error {

	this.deepSetDefault = map[string]int{}

	return this.setDefaultsDeep(reply)
}

func (this *Session) setDefaultsDeep(reply interface{}) error {

	if reply == nil {
		return nil
	}

	// value e type of pointer
	refValue := reflect.ValueOf(reply)
	//refType := reflect.TypeOf(reply)

	// value e type of instance
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)

		fieldStruct := fullValue.FieldByName(field.Name)
		fieldValue := fieldStruct.Interface()
		//fieldType := fieldStruct.Type()

		tags := this.getTags(field)

		if tags != nil && this.hasTag(tags, "ignore_set_default") {
			continue
		}

		zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
		_, ok := fieldValue.(Model)

		if ok {

			if this.Debug {
				logs.Debug("set defaults to ", fullValue, field.Name)
			}

			if zero {
				fieldFullType := reflect.TypeOf(fieldValue).Elem()
				newRefValue := reflect.New(fieldFullType)
				fieldStruct.Set(newRefValue)
				fieldValue = fieldStruct.Interface()
			}

			key := fmt.Sprintf("%v.%v", strings.Split(fullType.String(), ".")[1], field.Name)

			if count, ok := this.deepSetDefault[key]; ok {

				if count >= 5 {
					continue
				}

				this.deepSetDefault[key] = count + 1
			} else {
				this.deepSetDefault[key] = 1
			}

			if tags != nil && this.hasTag(tags, "ignore_set_default_child") {
				continue
			}

			this.setDefaultsDeep(fieldValue)
		}

	}

	return nil
}

func (this *Session) hasTag(tags []string, tagName string) bool {

	for _, tag := range tags {
		if tag == tagName {
			return true
		}
	}

	return false
}

func (this *Session) setTenant(reply interface{}) {

	// value e type of pointer
	refValue := reflect.ValueOf(reply)
	//refType := reflect.TypeOf(reply)

	// value e type of instance
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	if this.Debug {
		logs.Debug("## set Tenant to ", fullType)
	}

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)

		fieldStruct := fullValue.FieldByName(field.Name)
		//fieldValue := fieldStruct.Interface()
		//fieldType := fieldStruct.Type()

		tags := this.getTags(field)

		if this.hasTag(tags, "tenant") {

			value := reflect.ValueOf(this.Tenant)

			if this.Debug {
				logs.Debug("## field %v is tenant set tenant %v", field.Name, value)
			}

			// check is null
			//logs.Debug("## set Tenant to ", fullType)
			elem := fieldStruct.Elem()
			//logs.Debug("elem = %v", elem)
			//valueOf := reflect.ValueOf(elem)
			//logs.Debug("valueOf = %v, IsValid = %v", valueOf, elem.IsValid())
			if !elem.IsValid() {
				//logs.Debug("==== auto set tenant")
				fieldStruct.Set(value)
			} else {
				//logs.Debug("==== auto set tenant: tenant already")
			}

		} else {
			if this.Debug {
				logs.Debug("## field %v not is tenant ", field.Name)
			}
		}

	}

}

func (this *Session) getTags(field reflect.StructField) []string {

	tag := field.Tag.Get("goutils")
	var tags []string

	if len(strings.TrimSpace(tag)) > 0 {
		tags = strings.Split(tag, ";")
	}

	return tags
}

func (this *Session) setTenantFilter(entity interface{}, query orm.QuerySeter) orm.QuerySeter {

	if !this.isTenantNil() && this.HasFilterTenant(entity) {
		if model, ok := this.Tenant.(Model); ok {
			if model.IsPersisted() {
				query = query.Filter("Tenant", this.Tenant)
			}
		}
	}

	return query
}

func (this *Session) HasTenant() bool {
	return !this.isTenantNil()
}

func (this *Session) isTenantNil() bool {

	isNil := true

	if this.Tenant != nil {
		value := reflect.ValueOf(this.Tenant)
		isNil = value.IsNil()
	}

	return isNil
}

func (this *Session) isSetTenant(reply interface{}) bool {

	// value e type of pointer
	refValue := reflect.ValueOf(reply)
	//refType := reflect.TypeOf(reply)

	// value e type of instance
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	if this.Debug {
		logs.Debug("## set Tenant to ", fullType)
	}

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)

		if field.Name == "Tenant" {

			tags := this.getTags(field)

			if tags == nil || len(tags) == 0 {
				//logs.Debug("## set tenant")
				return true
			}

			set := !this.hasTag(tags, "no_set_tenant")

			//logs.Debug("## set tenant = %v", set)

			return set

		}

	}

	//logs.Debug("## not set tenant")
	return false
}

func (this *Session) checkIsAuthorizedTenant(reply interface{}, action string) bool {

	ignoreAuthorizedTenantCheckError := false

	if this.IgnoreAuthorizedTenantCheck {
		ignoreAuthorizedTenantCheckError = true
	}

	if !this.HasFilterTenant(reply) || this.isTenantNil() {
		ignoreAuthorizedTenantCheckError = true
	}

	// value e type of pointer
	refValue := reflect.ValueOf(reply)
	//refType := reflect.TypeOf(reply)

	// value e type of instance
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	if this.Debug {
		logs.Debug("## check is same tenant ", fullType)
	}

	// return true if entity not has tenant, not is manager security
	tenantFieldNotFound := true

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)

		if field.Name == "Tenant" {

			tenantFieldNotFound = false

			fieldStruct := fullValue.FieldByName(field.Name)
			fieldValue := fieldStruct.Interface()

			zero := reflect.Zero(reflect.TypeOf(fieldValue)).Interface() == fieldValue
			if !zero {
				if entityTenant, ok := fieldValue.(TenantModel); ok {

					nilEntityTenant := reflect.Zero(reflect.TypeOf(entityTenant)).Interface() == entityTenant

					if this.AuthorizedTenants != nil && !nilEntityTenant {
						for _, authorizedTenant := range this.AuthorizedTenants {
							if entityTenant.GetId() == authorizedTenant.GetId() {
								//logs.Debug("check entity tenant = ", entityTenant.GetId(), ", auth tenant = ", authorizedTenant.GetId())
								return true
							}
						}
					}

					if currentTenant, ok := this.Tenant.(TenantModel); ok && !nilEntityTenant {

						nilCurrentTenant := reflect.Zero(reflect.TypeOf(currentTenant)).Interface() == currentTenant

						if !nilCurrentTenant {

							authorized := currentTenant.GetId() == entityTenant.GetId()

							if ignoreAuthorizedTenantCheckError {
								return true
							}

							if !authorized {
								logs.Debug("+++ WARN: unautorized tenant ", currentTenant.GetId(), "to entity tenant = ", entityTenant.GetId(), " for type ", fullType, " content = ", reply, ", action = ", action)
							}

							return authorized

						}

					} else {
						logs.Debug("=======================================================")
						logs.Debug("=== WARN!!! Current Tenant id empty for entity type = ", fullType, " content = ", reply, ", action = ", action)
						logs.Debug("=======================================================")
					}

				}
			} else if !ignoreAuthorizedTenantCheckError {
				logs.Debug("=======================================================")
				logs.Debug("=== WARN!!! Tenant id empty for entity type = ", fullType, " content = ", reply, ", action = ", action)
				logs.Debug("=======================================================")
			}
		}
	}

	//logs.Debug("does not authorize data access")
	//logs.Debug("## not set tenant")
	return false || tenantFieldNotFound || ignoreAuthorizedTenantCheckError
}

func (this *Session) HasFilterTenant(reply interface{}) bool {

	// value e type of pointer
	refValue := reflect.ValueOf(reply)
	//refType := reflect.TypeOf(reply)

	// value e type of instance
	fullValue := refValue.Elem()
	fullType := fullValue.Type()

	for i := 0; i < fullType.NumField(); i++ {
		field := fullType.Field(i)

		if field.Name == "Tenant" {

			tags := this.getTags(field)

			if !this.hasTag(tags, "tenant") {
				//logs.Debug("## filter tenant")
				return false
			}

			if this.IgnoreTenantFilter {
				return false
			}

			noFilter := this.hasTag(tags, "no_filter_tenant")

			//logs.Debug("## filter tenant = %v", filter)

			if noFilter {
				return false
			}

			return true

		}

	}
	//logs.Debug("## filter tenant")
	return false
}
