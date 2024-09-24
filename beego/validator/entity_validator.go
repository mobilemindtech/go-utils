package validator

import (
	"fmt"
	"reflect"

	"errors"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/core/validation"
	"github.com/beego/i18n"
	"github.com/mobilemindtec/go-utils/support"
	"github.com/mobilemindtec/go-utils/v2/optional"
	"github.com/mobilemindtec/go-io/result"
	iov "github.com/mobilemindtec/go-io/validation"
	"strings"
)



type EntityValidatorResult struct {
	Errors       map[string]string
	ErrorsFields map[string]string
	HasError     bool
}


func (this *EntityValidatorResult) Error() string {
	if len(this.Errors) == 0 && len(this.ErrorsFields)  == 0 {
		return  ""
	}

	var lst []string

	for k, v := range this.Errors {
		lst = append(lst, fmt.Sprintf("%v: %v", k, v))
	}

	return strings.Join(lst, ", ")
}

func (this *EntityValidatorResult) Merge(result *EntityValidatorResult) {
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

type Validation = validation.Validation
type CustomAction func(validator *Validation)
type CustomValidation func(entity interface{}, validator *Validation)

type ValidatorForType struct {
	Fn  CustomValidation
	Typ reflect.Type
}

func NewEntityValidatorResult() *EntityValidatorResult {
	return &EntityValidatorResult{Errors: make(map[string]string), ErrorsFields: make(map[string]string)}
}

type EntityValidator struct {
	Lang              string
	ViewPath          string
	valActions        []CustomValidation
	valActionsForType []*ValidatorForType
	values            []interface{}
}

func NewEntityValidator(lang string, viewPath string) *EntityValidator {
	return &EntityValidator{Lang: lang, ViewPath: viewPath}
}
func New() *EntityValidator {
	return &EntityValidator{values: []interface{}{}, valActions: []CustomValidation{}}
}

func (this *EntityValidator) AddValidation(acs ...CustomValidation) *EntityValidator {
	for _, ac := range acs {
		this.valActions = append(this.valActions, ac)
	}
	return this
}

func (this *EntityValidator) AddValidationForType(t reflect.Type, ac CustomValidation) *EntityValidator {
	this.valActionsForType = append(this.valActionsForType, &ValidatorForType{ac, t})
	return this
}

func (this *EntityValidator) AddEntities(vs ...interface{}) *EntityValidator {
	for _, it := range vs {
		this.values = append(this.values, it)
	}
	return this
}

func (this *EntityValidator) AddEntity(vs interface{}) *EntityValidator {
	this.values = append(this.values, vs)
	return this
}

func (this *EntityValidator) ValidateOpt(entities ...interface{}) *optional.Optional[*EntityValidator] {
	return optional.Of[*EntityValidator](
		this.Validate(entities...))
}

func (this *EntityValidator) Validate(entities ...interface{}) interface{} {

	result, err := this.ValidMult(entities, nil)

	if err != nil {
		logs.Error("err = %v", err)
		return optional.NewFail(err)
	}

	if result.HasError {
		results := this.GetValidationErrors(result)
		return optional.NewFailWithItem(errors.New("validation error"), results)
	}

	return optional.SomeOk()
}


func (this *EntityValidator) ValidateResultWith(entity interface{}, f func(validator *Validation)) *result.Result[iov.Validation] {
	return result.Try(func() (iov.Validation, error) {
		val, err := this.ValidMult([]interface{}{entity}, f)
		if err != nil {
			return nil, err
		}
		if val.HasError {
			return iov.WithErrors(val.Errors), nil
		}
		return iov.NewSuccess(), nil
	})
}
func (this *EntityValidator) ValidatetResult(entities ...interface{}) *result.Result[iov.Validation] {
	return result.Try(func() (iov.Validation, error) {
		val, err := this.ValidMult(entities, nil)
		if err != nil {
			return nil, err
		}
		if val.HasError {
			return iov.WithErrors(val.Errors), nil
		}
		return iov.NewSuccess(), nil
	})
}

func (this *EntityValidator) ValidMult(entities []interface{}, action func(validator *Validation)) (*EntityValidatorResult, error) {

	this.AddEntities(entities...)

	result := NewEntityValidatorResult()

	customApplyDone := false

	for _, it := range this.values {

		if it == nil {
			continue
		}

		if !customApplyDone {

			ev, err := this.IsValid(it, action)

			if err != nil {
				return nil, err
			}
			result.Merge(ev)

			for _, ac := range this.valActions {

				ev, err := this.IsValid(it, func(v *Validation) {
					ac(it, v)
				})
				if err != nil {
					return nil, err
				}
				result.Merge(ev)
			}

			for _, ac := range this.valActionsForType {
				if ac.Typ == reflect.TypeOf(it) {
					ev, err := this.IsValid(it, func(v *Validation) {
						ac.Fn(it, v)
					})
					if err != nil {
						return nil, err
					}
					result.Merge(ev)
				}
			}
			customApplyDone = true
		}

	}

	return result, nil

}
func (this *EntityValidator) IsValid(entity interface{}, action CustomAction) (*EntityValidatorResult, error) {
	return this.Valid(entity, action)
}

func (this *EntityValidator) ValidSimple(entity interface{}) (*EntityValidatorResult, error) {
	return this.Valid(entity, nil)
}

func (this *EntityValidator) Valid(entity interface{}, action CustomAction) (*EntityValidatorResult, error) {

	result := NewEntityValidatorResult()

	localValid := Validation{}
	callerValid := Validation{}

	typeName := ""

	if entity != nil {

		typeName = reflect.TypeOf(entity).Elem().Name()

		typeName = support.Underscore(typeName)

		//logs.Debug("typeName = %v", typeName)

		ok, err := localValid.Valid(entity)

		if err != nil {
			return nil, err
		}

		if !ok {
			for _, err := range localValid.Errors {

				lbl := this.ViewPath

				if lbl == "" {
					lbl = typeName
				}

				if lbl != "" {
					label := this.GetMessage(fmt.Sprintf("%s.%s", lbl, err.Field))
					result.Errors[label] = err.Message
				} else {
					result.Errors[err.Field] = err.Message
				}

				result.ErrorsFields[err.Field] = err.Message

				//logs.Debug("## ViewPath %v", this.ViewPath)
				//logs.Debug("## lebel %v", label)
				logs.Debug(fmt.Sprintf("* validator error field %v.%v error %v", typeName, err.Field, err))
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
			} else {
				result.Errors[err.Field] = err.Message
			}

			result.ErrorsFields[err.Field] = err.Message

			logs.Debug(fmt.Sprintf("* validator error field %v.%v error %v", typeName, err.Field, err))
		}

		result.HasError = true
	}

	return result, nil
}

func (this *EntityValidator) GetValidationErrors(result *EntityValidatorResult) map[string]string {
	data := make(map[interface{}]interface{})
	this.CopyErrorsToView(result, data)
	return data["errors"].(map[string]string)
}

func (this *EntityValidator) GetValidationResults(result *EntityValidatorResult) []map[string]string {
	data := make(map[interface{}]interface{})
	this.CopyErrorsToView(result, data)
	validations := data["errors"].(map[string]string)
	results := []map[string]string{}
	for k, v := range validations {
		results = append(results, map[string]string{
			"field":   k,
			"message": v,
		})
	}
	return results
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

func (this *EntityValidator) GetMessage(key string, args ...interface{}) string {
	return i18n.Tr(this.Lang, key, args)
}
