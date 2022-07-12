package db

import (
	"github.com/beego/beego/v2/client/orm"

	"fmt"
)


type RawQuery struct {

  Query string
  Args []interface{}

  Session *Session

  Error error
  RowsAffected int64

  valuesTransformFunc func(map[string]interface{}) (interface{}, error)
  valuesListTransformFunc func([]interface{}) (interface{}, error)

  values []orm.Params
  valuesList []orm.ParamsList
  valuesFlat orm.ParamsList

}

func NewRawQuerySession(session *Session) *RawQuery {
	return &RawQuery { Session: session }
}

func NewRawQuery(session *Session, query string) *RawQuery {
	q := &RawQuery { Session: session }
	return q.WithQuery(query)
}

func NewRawQueryArgs(session *Session, query string, args ...interface{}) *RawQuery {
	q := &RawQuery { Session: session }
	return q.WithQuery(query).WithArgs(args...)
}

func (this *RawQuery) HasError() bool {
	return this.Error != nil
}

func (this *RawQuery) WithQuery(query string) *RawQuery {
	this.Query = query
	return this
}

func (this *RawQuery) WithArgs(args ...interface{}) *RawQuery{
	this.Args = args
	return this
}

func (this *RawQuery) WithValuesTransformer(tr func(map[string]interface{}) (interface{}, error)) *RawQuery{
	this.valuesTransformFunc = tr
	return this
}

func (this *RawQuery) WithValuesListTransformer(tr func([]interface{}) (interface{}, error)) *RawQuery{
	this.valuesListTransformFunc = tr
	return this
}

func (this *RawQuery) RawSeter() orm.RawSeter{  
  return this.Session.GetDb().Raw(this.Query, this.Args...)
}

func (this *RawQuery) Execute() (int64, error){  
  res, err  := this.Session.GetDb().Raw(this.Query, this.Args...).Exec()

  if err != nil {
  	this.Error = err
  	return 0, err
  }

  this.RowsAffected, this.Error = res.RowsAffected()

	return this.RowsAffected, this.Error
}

func (this *RawQuery) ToResutl(result interface{}) error{  
  this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).QueryRow(result)
  return this.Error
}

func (this *RawQuery) ToResutls(results interface{}) error{  
  this.RowsAffected, this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).QueryRows(results)
  return this.Error
}

func (this *RawQuery) Values() *RawQuery{   
  this.RowsAffected, this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).Values(&this.values)
  return this
}

func (this *RawQuery) ValuesList() *RawQuery{   
  this.RowsAffected, this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).ValuesList(&this.valuesList)
  return this
}

func (this *RawQuery) ValuesFlat() *RawQuery{   
  this.RowsAffected, this.Error = this.Session.GetDb().Raw(this.Query, this.Args...).ValuesFlat(&this.valuesFlat)
  return this
}

func (this *RawQuery) GetValuesFlat() []interface{} {
	results := []interface{}{}

	for _, x := range this.valuesFlat {
		results = append(results, x)
	}

	return results
}

func (this *RawQuery) GetValues() []map[string]interface{} {
  results := []map[string]interface{}{}

  for _, item := range this.values {

    itemMap := make(map[string]interface{})

    for k, v := range item {
      itemMap[k] = v      
    }

    results = append(results, itemMap)
  }

  return results

}

func (this *RawQuery) GetValuesList() [][]interface{} {  

	results := [][]interface{}{}

  for _, item := range this.valuesList {

    newItem := []interface{}{}

    for _, v := range item {
      newItem = append(newItem, v)
    }

    results = append(results, newItem)
  }

  return results
}


func (this *RawQuery) TransformValues() ([]interface{}, error){ 

	if this.valuesListTransformFunc == nil && this.valuesTransformFunc == nil {
		return nil, fmt.Errorf("use values or value list transformer function")
	}

	if this.valuesListTransformFunc != nil && this.valuesTransformFunc != nil {
		return nil, fmt.Errorf("use values or value list transformer function")
	}

	results := []interface{}{}

	if this.valuesTransformFunc != nil {

		values := this.GetValues()

		for _, it := range values {
			result, err := this.valuesTransformFunc(it)

			if err != nil {
				return nil, err
			}

			results = append(results, result)
		}


	} else {

	  values := this.GetValuesList()

	  for _, item := range values {
	    result, err := this.valuesListTransformFunc(item)

	    if err != nil {
	      return nil, err
	    }

	    results = append(results, result)
	  }

	}

 

  return results, nil

}