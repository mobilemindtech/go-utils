package response

import (
	"fmt"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/mobilemindtech/go-utils/v2/maps"
)

type RenderResponse interface {
	RenderResponse(resp *Response)
}

type Response struct {
	TemplateName     string
	FullTemplateName string
	Data             map[string]interface{} `jsonp:""`
	Result           interface{}            `jsonp:""`
	Results          interface{}            `jsonp:""`
	TotalCount       int64                  `jsonp:""`
	Entity           interface{}
	Entities         interface{}
	Error            bool              `jsonp:""`
	Message          string            `jsonp:""`
	Errors           map[string]string `jsonp:""`

	Value interface{}

	JsonResult             bool
	JsonPackage            bool
	JsonPackageAsCamelCase bool

	validation    *Validation
	validationMap *ValidationMap
	err           error

	flashError   string
	flashSuccess string
	flashNotice  string
	flashWarning string
}

func New() *Response {
	return &Response{}
}

func WithTpl(templateName string) *Response {
	return &Response{TemplateName: templateName}
}

func WithJSON(data map[string]interface{}) *Response {
	return &Response{Data: data}
}

func (this *Response) HasTemplate() bool {
	return len(this.TemplateName) > 0 || len(this.FullTemplateName) > 0
}

func (this *Response) GetTemplate(path string) string {
	if len(this.FullTemplateName) > 0 {
		return this.FullTemplateName
	}
	return fmt.Sprintf("%s/%s.tpl", path, this.TemplateName)
}

func (this *Response) SetValue(value interface{}) *Response {
	this.Value = value
	return this
}

func (this *Response) HasValue() bool {
	return this.Value != nil
}

func (this *Response) UseJsonResult() *Response {
	this.JsonResult = true
	return this
}

func (this *Response) UseJsonPackage() *Response {
	this.JsonPackage = true
	return this
}

func (this *Response) UseJsonPackageAsCamelCase() *Response {
	this.JsonPackageAsCamelCase = true
	return this
}

func (this *Response) MkJsonResult() *JsonResult {
	return &JsonResult{
		Result:     this.Result,
		Results:    this.Results,
		Error:      this.Error,
		Message:    this.Message,
		TotalCount: this.TotalCount,
		Errors:     this.Errors,
	}
}

func (this *Response) SetTemplate(tpl string) *Response {
	this.TemplateName = tpl
	return this
}

func (this *Response) SetFullTemplate(tpl string) *Response {
	this.FullTemplateName = tpl
	return this
}

func (this *Response) SetData(data map[string]interface{}) *Response {
	this.Data = data
	return this
}

func (this *Response) SetKV(args ...interface{}) *Response {
	this.Data = maps.JSON(args...)
	return this
}

func (this *Response) SetResult(result interface{}) *Response {
	this.Result = result
	return this
}

func (this *Response) SetResults(results interface{}) *Response {
	this.Results = results
	return this
}

func (this *Response) SetEntity(entity interface{}) *Response {
	this.Entity = entity
	return this
}

func (this *Response) SetEntities(entities interface{}) *Response {
	this.Entities = entities
	return this
}

func (this *Response) SetTotalCount(totalCount int64) *Response {
	this.TotalCount = totalCount
	return this
}

func (this *Response) SetError() *Response {
	this.Error = true
	return this
}

func (this *Response) SetMessage(msg string, args ...interface{}) *Response {
	this.Message = fmt.Sprintf(msg, args...)
	return this
}

func (this *Response) SetErrors(errors map[string]string) *Response {
	this.Errors = errors
	return this
}

func (this *Response) FlashError(msg string, args ...interface{}) *Response {
	this.flashError = fmt.Sprintf(msg, args...)
	return this
}
func (this *Response) FlashSuccess(msg string, args ...interface{}) *Response {
	this.flashSuccess = fmt.Sprintf(msg, args...)
	return this
}
func (this *Response) FlashNotice(msg string, args ...interface{}) *Response {
	this.flashNotice = fmt.Sprintf(msg, args...)
	return this
}
func (this *Response) FlashWarning(msg string, args ...interface{}) *Response {
	this.flashWarning = fmt.Sprintf(msg, args...)
	return this
}

func (this *Response) ConfigureFlash(flashData *beego.FlashData) *Response {
	if len(this.flashError) > 0 {
		flashData.Error(this.flashError)
	} else if len(this.flashNotice) > 0 {
		flashData.Notice(this.flashError)
	} else if len(this.flashSuccess) > 0 {
		flashData.Success(this.flashError)
	} else if len(this.flashWarning) > 0 {
		flashData.Warning(this.flashError)
	}
	return this
}

func (this *Response) Err(err error) *Response {
	this.Error = true
	if v, ok := err.(*Validation); ok {
		this.validation = v
	} else if v, ok := err.(*ValidationMap); ok {
		this.validationMap = v
	} else {
		err = this.err
	}
	return this
}

func (this *Response) HasErr() bool {
	return this.err != nil
}

func (this *Response) GetErr() error {
	return this.err
}

func (this *Response) HasValidation() bool {
	return this.validation != nil
}

func (this *Response) GetValidation() *Validation {
	return this.validation
}

func (this *Response) HasValidationMap() bool {
	return this.validationMap != nil
}

func (this *Response) GetValidationMap() *ValidationMap {
	return this.validationMap
}

func (this *Response) On(controller RenderResponse) {
	controller.RenderResponse(this)
}
