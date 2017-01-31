package db

import (
  "github.com/astaxie/beego/orm"
  "errors"
  "fmt" 
)

type SessionState int

const (
  SessionStateOk SessionState = iota + 1  
  SessionStateError
)

type Model interface {
  IsPersisted() bool
  TableName() string  
}

type Session struct {
  Db orm.Ormer
  State SessionState
  TenantId int64
}


func NewSession() *Session{
  return &Session{ State: SessionStateOk }
}

func NewSessionWithTenantId(tenantId int64) *Session{
  return &Session{ State: SessionStateOk, TenantId: tenantId }
}

func (this *Session) OnError() *Session {
  this.State = SessionStateError
  return this
}

func (this *Session) Open() orm.Ormer{
  return this.Begin()
}

func (this *Session) Close() {
  if this.State == SessionStateOk {
    this.Commit()
  } else {
    this.Rollback()
  }
}

func (this *Session) Begin() orm.Ormer{
  this.Db = orm.NewOrm()
  this.Db.Using("default")    

  err := this.Db.Begin()
  if err != nil {
    fmt.Println("## db begin error: %v", err.Error())
    panic(err)
  }

  return this.Db
}

func (this *Session) Commit() {
  fmt.Println("## session commit ")

  if this.Db != nil{
    err := this.Db.Commit()
    if err != nil {
      fmt.Println("## db commit error: %v", err.Error())
      this.Rollback()
      panic(err)
    }
    this.Db = nil
  }
}

func (this *Session) Rollback() {
  fmt.Println("## session rollback ")

  if this.Db != nil{
    err := this.Db.Rollback() 
    if err != nil {
      fmt.Println("## db rollback error: %v", err.Error())
      panic(err)
    }
    this.Db = nil
  }
}

func (this *Session) Save(entity interface{}) error {  
  num, err := this.Db.Insert(entity)
  
  if err != nil {
    fmt.Println("## Session: error on save: %v", err.Error())
    this.OnError()
    return err
  }

  if num == 0 {
    this.OnError()
    return errors.New("save row count is zero")
  }


  return nil
}

func (this *Session) Update(entity interface{}) error {
  num, err := this.Db.Update(entity)

  if err != nil {
    fmt.Println("## Session: error on update: %v", err.Error())
    this.OnError()
    return err
  }

  if num == 0 {
    this.OnError()
    return errors.New("update row count is zero")
  }


  return nil
}

func (this *Session) Remove(entity interface{}) error {
  
  num, err := this.Db.Delete(entity)

  if err != nil {
    fmt.Println("## Session: error on remove: %v", err.Error())
    this.OnError()
    return err
  }

  if num == 0 {
    this.OnError()
    return errors.New("update row count is zero")
  }


  return nil
}

func (this *Session) Load(entity interface{}) error {
  if err := this.Db.Read(entity); err != nil {
    fmt.Println("## Session: error on load: %v", err.Error())
    //this.OnError()
    return err
  }
  return nil
}

func (this *Session) Count(entity interface{}) (int64, error){  
  
  if model, ok := entity.(Model); ok {
    fmt.Println("## count by table %v", model.TableName())
    num, err := this.Db.QueryTable(model.TableName()).Count()
    if err != nil {
      fmt.Println("## Session: error on count: %v", err.Error())
      //this.OnError()
    }
    return num, err
  }

  this.OnError()
  return 0, errors.New("entity does not implements of Model")
}

func (this *Session) HasById(entity interface{}, id int64) (bool, error) {  
  
  if model, ok := entity.(Model); ok {
    return this.Db.QueryTable(model.TableName()).Filter("id", id).Exist(), nil
  }
  
  this.OnError()
  return false, errors.New("entity does not implements of Model")
}

func (this *Session) FindById(entity interface{}, id int64) (interface{}, error) {  
  
  if model, ok := entity.(Model); ok {
    err := this.Db.QueryTable(model.TableName()).Filter("id", id).One(entity)

    if err != nil{
      fmt.Println("## Session: error on find by id: %v", err.Error())
      //this.OnError()
      return entity, err
    }

    if model.IsPersisted() {
      return entity, nil
    }

    return entity, nil
  }
  
  this.OnError()
  return false, errors.New("entity does not implements of Model")
}

func (this *Session) SaveOrUpdate(entity interface{}) error{

  if model, ok := entity.(Model); ok {
    if model.IsPersisted() {
      return this.Update(entity)
    }
    return this.Save(entity)
  }

  this.OnError()
  return errors.New("entity does not implements of Model")
}

func (this *Session) List(entity interface{}, entities interface{}) error { 
  if model, ok := entity.(Model); ok {

    query := this.Db.QueryTable(model.TableName())
    
    if this.TenantId > 0 {
      query.Filter("Tenant__Id", this.TenantId)
    }

    if _, err := query.All(entities); err != nil {
      fmt.Println("## Session: error on list: %v", err.Error())
      //this.OnError()
      return err
    }
    return nil
  }

  this.OnError()
  return errors.New("entity does not implements of Model 1")
}

func (this *Session) Page(entity interface{}, entities interface{}, page *Page) error { 
  if model, ok := entity.(Model); ok {
    query := this.Db.QueryTable(model.TableName())
    
    query = query.Limit(page.Limit).Offset(page.Offset)   

    if page.Sort != "" { 
      query = query.OrderBy(fmt.Sprintf("%v%v", page.Order, page.Sort))
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
          fmt.Println("### cond %v=%v", k, v)
        }        
        query = query.SetCond(orm.NewCondition().AndCond(cond))
      }

    }

    if page.AndFilterColumns != nil && len(page.AndFilterColumns) > 0 {      
      for k, v := range page.AndFilterColumns {
        query = query.Filter(k, v)
      }      
    }

    if this.TenantId > 0 {
      query.Filter("Tenant__Id", this.TenantId)
    }    
 
    if _, err := query.All(entities); err != nil {
      fmt.Println("## Session: error on page: %v", err.Error())
      //this.OnError()
      return err
    }


    return nil
  }

  this.OnError()
  return errors.New("entity does not implements of Model")
}

func (this *Session) Query(entity interface{}) (orm.QuerySeter, error) { 
  if model, ok := entity.(Model); ok {
    query := this.Db.QueryTable(model.TableName())
    
    if this.TenantId > 0 {
      query.Filter("Tenant__Id", this.TenantId)
    }

    return query, nil
  }

  this.OnError()
  return nil, errors.New("entity does not implements of Model")
}

func (this *Session) ToList(querySeter orm.QuerySeter, entities interface{}) error {
  if _, err := querySeter.All(entities); err != nil {
    fmt.Println("## Session: error on to list: %v", err.Error())
    //this.OnError()
    return err
  }
  return nil
}

func (this *Session) ToOne(querySeter orm.QuerySeter, entity interface{}) error {
  if err := querySeter.One(entity); err != nil {
    fmt.Println("## Session: error on to one: %v", err.Error())
    //this.OnError()
    return err
  }
  return nil
}

func (this *Session) ToPage(querySeter orm.QuerySeter, entities interface{}, page *Page) error {
  querySeter.Limit(page.Limit)
  querySeter.Offset(page.Offset)  
  if _, err := querySeter.All(entities); err != nil {
    fmt.Println("## Session: error on to page: %v", err.Error())
    //this.OnError()
    return err
  }
  return nil
}