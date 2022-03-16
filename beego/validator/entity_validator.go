package validator

import (
  "github.com/mobilemindtec/go-utils/support"
  "github.com/beego/beego/v2/core/validation"
  "github.com/beego/i18n"
  "reflect"
  "fmt"
)

type EntityValidatorResult struct {
  Errors map[string]string
  ErrorsFields map[string]string
  HasError bool
}

func (this *EntityValidatorResult) Merge(result *EntityValidatorResult){
  for k, v := range result.Errors {
    this.Errors[k] = v
  }
  for k, v := range result.ErrorsFields {
    this.ErrorsFields[k] = v
  }

  if !this.HasError {
    this.HasError = result.HasError
  }
}

func NewEntityValidatorResult() *EntityValidatorResult {
  return &EntityValidatorResult{ Errors: make(map[string]string), ErrorsFields: make(map[string]string) }
}

type EntityValidator struct {
  Lang string
  ViewPath string
}

func NewEntityValidator(lang string, viewPath string) *EntityValidator{
  return &EntityValidator{ Lang: lang, ViewPath: viewPath }
}

func (this *EntityValidator) ValidMult(entities []interface{}, action func(validator *validation.Validation)) (*EntityValidatorResult, error) {

  result := NewEntityValidatorResult()

  funcApply := action

  for _, it := range entities {

    if it == nil {
      continue
    }

    ev, err := this.IsValid(it, funcApply)
    if err != nil {
      return nil, err
    }

    funcApply = nil // aplica apenas para a primeira validação

    result.Merge(ev)
  }

  return result, nil

}
func (this *EntityValidator) IsValid(entity interface{}, action func(validator *validation.Validation)) (*EntityValidatorResult, error) {
  return this.Valid(entity, action)
}

func (this *EntityValidator) Valid(entity interface{}, action func(validator *validation.Validation)) (*EntityValidatorResult, error) {

  result := NewEntityValidatorResult()

  localValid := validation.Validation{}
  callerValid := validation.Validation{}

  typeName := ""

  if entity != nil {

    typeName = reflect.TypeOf(entity).Elem().Name()

    typeName = support.Underscore(typeName)

    //fmt.Println("typeName = %v", typeName)

    ok, err := localValid.Valid(entity)

    if  err != nil {
      fmt.Println("## error on run validation = ", err.Error())
      return nil, err
    }

    if !ok {
      for _, err := range localValid.Errors {

        label := this.GetMessage(fmt.Sprintf("%s.%s", typeName, err.Field))

        if label == "" {
          label = this.GetMessage(fmt.Sprintf("%s.%s", this.ViewPath, err.Field))
        }

        if label != "" {
          result.Errors[label] = err.Message
        }else{
          result.Errors[err.Field] = err.Message
        }

        result.ErrorsFields[err.Field] = err.Message

        //fmt.Println("## ViewPath %v", this.ViewPath)
        //fmt.Println("## lebel %v", label)
        fmt.Println(fmt.Sprintf("* validator error field %v.%v error %v", typeName, err.Field, err))
      }

      result.HasError = true
    }
  }

  if action != nil {
    action(&callerValid)
  }

  if callerValid.HasErrors() {
    for _, err := range callerValid.Errors {

      label := this.GetMessage(fmt.Sprintf("%s.%s", typeName, err.Field))

      if label == "" {
        label = this.GetMessage(fmt.Sprintf("%s.%s", this.ViewPath, err.Field))
      }

      if label != "" {
        result.Errors[label] = err.Message
      }else{
        result.Errors[err.Field] = err.Message
      }

      result.ErrorsFields[err.Field] = err.Message

      fmt.Println(fmt.Sprintf("* validator error field %v.%v error %v", typeName, err.Field, err))
    }

    result.HasError = true
  }

  return result, nil
}

func (this *EntityValidator) GetValidationErrors(result *EntityValidatorResult) map[string]string{
  data := make(map[interface{}]interface{})
  this.CopyErrorsToView(result, data)
  return data["errors"].(map[string]string)
}

func (this *EntityValidator) CopyErrorsToView(result *EntityValidatorResult, data map[interface{}]interface{}) {
  
  if len(result.Errors) > 0 {

    if data["errorsFields"] == nil {
      
      data["errorsFields"] = result.ErrorsFields
      data["errors"] = result.Errors

    } else {

      mapItem := data["errorsFields"].(map[string]string)
      for k, v := range result.ErrorsFields {
        mapItem[k] = v
      }

      mapItem = data["errors"].(map[string]string)
      for k, v := range result.Errors {
        mapItem[k] = v      
      }

    }

  }
}

func (this *EntityValidator) GetMessage(key string, args ...interface{}) string{
  return i18n.Tr(this.Lang, key, args)
}
