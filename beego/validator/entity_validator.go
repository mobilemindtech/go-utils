package validator

import (
	"fmt"
	"reflect"

	"github.com/beego/beego/v2/core/logs"
	"github.com/beego/beego/v2/core/validation"
	"github.com/beego/i18n"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-io/types/unit"
	iov "github.com/mobilemindtech/go-io/validation"
	"github.com/mobilemindtech/go-utils/support"
	"github.com/mobilemindtech/go-utils/v2/maps"
	"github.com/mobilemindtech/go-utils/v2/optional"
	"strings"
)

type ValidationError struct {
	Message string
	List    []string
	Map     map[string]string
}

func NewValidationError() *ValidationError {
	return new(ValidationError)
}

func (this *ValidationError) Error() string {
	return this.Message
}

type EntityValidatorResult struct {
	Errors       map[string]string
	ErrorsFields map[string]string
	HasError     bool
}

func (this *EntityValidatorResult) Error() string {
	if len(this.Errors) == 0 && len(this.ErrorsFields) == 0 {
		return ""
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
type FuncValidation func(validator *Validation)
type ValidationForType func(entity interface{}, validator *Validation)

type ValidatorForType struct {
	Fn  ValidationForType
	Typ reflect.Type
}

func NewEntityValidatorResult() *EntityValidatorResult {
	return &EntityValidatorResult{Errors: make(map[string]string), ErrorsFields: make(map[string]string)}
}

type EntityValidator struct {
	Lang              string
	ViewPath          string
	valActionsFuncs   []FuncValidation
	valActionsForType []*ValidatorForType
	values            []interface{}
}

func NewEntityValidator(lang string, viewPath string) *EntityValidator {
	return &EntityValidator{Lang: lang, ViewPath: viewPath}
}

// New Create new validator with default lang pt-BR
func New() *EntityValidator {
	return &EntityValidator{
		Lang:   "pt-BR",
		values: []interface{}{}}
	//valActions: []CustomValidation{}}
}

// WithPath Configure path to find message, eg.:
// path = tenant and error on field Name
// so search by message tenant.Name
func (this *EntityValidator) WithPath(path string) *EntityValidator {
	this.ViewPath = path
	return this
}

func (this *EntityValidator) AddFuncValidation(acs ...FuncValidation) *EntityValidator {
	for _, ac := range acs {
		this.valActionsFuncs = append(this.valActionsFuncs, ac)
	}
	return this
}

func (this *EntityValidator) AddValidationForType(t reflect.Type, ac ValidationForType) *EntityValidator {
	this.valActionsForType = append(this.valActionsForType, &ValidatorForType{ac, t})
	return this
}

func (this *EntityValidator) AddEntities(vs ...interface{}) *EntityValidator {
	for _, it := range vs {
		if it != nil {
			this.values = append(this.values, it)
		}
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
		err := NewValidationError()
		err.Message = "validation error"
		err.Map = results
		err.List = maps.ToSliceKV(results)
		return optional.NewFail(err)
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

// ValidateResult Return a Result error if the error occurred, or return sucess. Success is a validation
// result, and has a IsFailure func to validation error sinalize
func (this *EntityValidator) ValidateResult(entities ...interface{}) *result.Result[iov.Validation] {
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

// ValidateOrError retorn a Result error if validation not pass or a error occurried.
func (this *EntityValidator) ValidateOrError(entities ...interface{}) *result.Result[*unit.Unit] {
	return result.Try(func() (*unit.Unit, error) {
		val, err := this.ValidMult(entities, nil)
		if err != nil {
			return nil, err
		}
		if val.HasError {
			return nil, &ValidationError{
				Message: val.Error(),
				Map:     val.Errors,
				List:    maps.ToSliceKV(val.Errors),
			}
		}
		return unit.OfUnit(), nil
	})
}

func (this *EntityValidator) ValidMult(entities []interface{}, action func(validator *Validation)) (*EntityValidatorResult, error) {

	this.AddEntities(entities...)

	result := NewEntityValidatorResult()

	//customApplyDone := false

	for i, it := range this.values {

		if it == nil {
			logs.Warning("skip validation for nil value")
			continue
		}

		//if !customApplyDone {

		var ev *EntityValidatorResult
		var err error

		if i == 0 { // execute action only one time
			ev, err = this.IsValid(it, action)
		} else {
			ev, err = this.IsValid(it, nil)
		}

		if err != nil {
			return nil, fmt.Errorf("ValidMult(0) error validating %v: %v", it, err)
		}
		result.Merge(ev)

		for _, ac := range this.valActionsForType {
			if ac.Typ == reflect.TypeOf(it) {
				ev, err := this.IsValid(it, func(v *Validation) {
					ac.Fn(it, v)
				})
				if err != nil {
					return nil, fmt.Errorf("ValidMult(1) error validating %v: %v", it, err)
				}
				result.Merge(ev)
			}
		}
		//customApplyDone = true
		//}

	}

	for _, ac := range this.valActionsFuncs {

		ev, err := this.IsValid(nil, func(v *Validation) {
			ac(v)
		})
		if err != nil {
			return nil, fmt.Errorf("ValidMult(2) error validating: %v", err)
		}
		result.Merge(ev)
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
			return nil, fmt.Errorf("Valid(0) error validating: %v", err)
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

				//logs.Error("Validatiuon error for %v.%v: %v", typeName, err.Field, err)
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
