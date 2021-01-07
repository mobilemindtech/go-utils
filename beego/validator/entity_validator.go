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

type EntityValidator struct {
  Lang string
  ViewPath string
}

func NewEntityValidator(lang string, viewPath string) *EntityValidator{
  return &EntityValidator{ Lang: lang, ViewPath: viewPath }
}

func (this *EntityValidator) IsValid(entity interface{}, action func(validator *validation.Validation)) (*EntityValidatorResult, error) {

  result := new(EntityValidatorResult)

  localValid := validation.Validation{}
  callerValid := validation.Validation{}
  result.Errors = make(map[string]string)
  result.ErrorsFields = make(map[string]string)

  typeName := ""

  if entity != nil {

    typeName = reflect.TypeOf(entity).Elem().Name()

    typeName = support.Underscore(typeName)

    fmt.Println("typeName = %v", typeName)

    ok, err := localValid.Valid(entity)

    if  err != nil {
      fmt.Println("## error on run validation %v", err.Error())
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
        fmt.Println("## validator error field %v.%v error %v", typeName, err.Field, err)
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

      fmt.Println("## validator error field %v.%v error %v", typeName, err.Field, err)
    }

    result.HasError = true
  }

  return result, nil
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
