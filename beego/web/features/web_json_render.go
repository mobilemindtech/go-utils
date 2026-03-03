package features

import (
	"fmt"
	"time"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-utils/beego/web/response"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	"github.com/mobilemindtech/go-utils/v2/maps"
	"github.com/mobilemindtech/go-utils/v2/optional"
)

// WebResponses Deprecated.
type WebJsonRender struct {
	base trait.WebBaseInterface
}

func (this *WebJsonRender) InitWebJsonRender(base trait.WebBaseInterface) {
	this.base = base
}

/*
func (this *WebResponses) SetViewModel(name string, data interface{}) *WebController {
	this.Data[name] = data
	return this
}

func (this *WebResponses) SetResults(results interface{}) *WebController {
	this.Data["results"] = results
	this.Data["entities"] = results
	return this
}

func (this *WebResponses) SetResultsAndTotalCount(results interface{}, totalCount int64) *WebController {
	this.Data["results"] = results
	this.Data["entities"] = results
	this.Data["totalCount"] = totalCount
	return this
}

func (this *WebResponses) SetResult(result interface{}) *WebController {
	this.Data["result"] = result
	this.Data["entity"] = result
	return this
}
*/

func (this *WebJsonRender) RenderJsonResult(opt interface{}) {

	//logs.Debug("RenderJsonResult = %v type of %v", opt, reflect.TypeOf(opt).Kind())

	switch opt.(type) {
	case *optional.Some:
		someVal := opt.(*optional.Some).Item

		switch someVal.(type) {
		case *criteria.Page:
			page := someVal.(*criteria.Page)
			this.OnJsonResultsWithTotalCount(page.Data, page.Count())
			break
		case *optional.Ok:
			this.OnJson200()
			break
		default:

			if val, ok := criteria.TryExtractPageIfPegeOf(someVal); ok {
				this.RenderJsonResult(val)
				return
			}

			if optional.IsSlice(someVal) {
				this.OnJsonResults(someVal)
			} else {
				this.OnJsonResult(someVal)
			}
		}
		break
	case *optional.None:
		this.NotFoundAsJson()
		break
	case *optional.Fail:

		fail := opt.(*optional.Fail)
		err := opt.(*optional.Fail).Error

		if err.Error() == "validation error" {
			this.OnJsonValidationWithErrors(fail.Item.(map[string]string))
		} else {
			this.OnJsonError(fmt.Sprintf("%v", err))
		}
		break
	case *response.JsonResult:
		this.OnJson(opt.(*response.JsonResult))
		break
	case *criteria.Page:
		page := opt.(*criteria.Page)
		this.OnJsonResultsWithTotalCount(page.Data, page.Count())
		break
	case error:
		this.OnJsonError(fmt.Sprintf("%v", opt.(error).Error()))
		break

	default:

		if val, ok := optional.TryExtractValIfOptional(opt); ok {
			this.RenderJsonResult(val)
			return
		}

		if val, ok := criteria.TryExtractPageIfPegeOf(opt); ok {
			this.RenderJsonResult(val)
			return
		}

		if optional.IsSlice(opt) {
			logs.Debug("render as results")
			this.OnJsonResults(opt)
		} else {
			logs.Debug("render as result")
			this.OnJsonResult(opt)
		}

		//this.OnJsonError(fmt.Sprintf("unknow optional value: %v", opt))
		break
	}
}

func (this *WebJsonRender) OnJsonResultNil() {
	this.OnJsonResult(nil)
}

func (this *WebJsonRender) OnJsonResult(result interface{}) {
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRender) GetJsonResult() (*response.JsonResult, bool) {
	if this.base.GetData()["json"] != nil {
		if j, ok := this.base.GetData()["json"].(*response.JsonResult); ok {
			return j, ok
		}
	}
	return nil, false
}

func (this *WebJsonRender) OnJsonMessage(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.renderDadaAsJson(&response.JsonResult{
		Message:         message,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRender) OnJsonResultError(result interface{}, format string, v ...interface{}) {
	this.base.SetErrorState()
	message := fmt.Sprintf(format, v...)
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRender) OnJsonResultWithMessage(result interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Error:           false,
		Message:         message,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRender) OnJsonResults(results interface{}) {
	this.renderDadaAsJson(&response.JsonResult{
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRender) OnJsonResultAndResults(result interface{}, results interface{}) {
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRender) OnJsonResultsWithTotalCount(results interface{}, totalCount int64) {
	this.renderDadaAsJson(&response.JsonResult{
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
		TotalCount:      totalCount,
	})
}

func (this *WebJsonRender) OnJsonPage(page *criteria.Page) {
	this.renderDadaAsJson(&response.JsonResult{
		Results:         page.Data,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
		TotalCount:      int64(page.TotalCount),
	})
}

func (this *WebJsonRender) OnJsonResultAndResultsWithTotalCount(result interface{}, results interface{}, totalCount int64) {
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
		TotalCount:      totalCount,
	})
}

func (this *WebJsonRender) OnJsonResultsError(results interface{}, format string, v ...interface{}) {
	this.base.SetErrorState()
	message := fmt.Sprintf(format, v...)
	this.renderDadaAsJson(&response.JsonResult{
		Results:         results,
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRender) RenderJsonWithStatusCode(data interface{}, status int) {
	this.base.SetViewData("json", data)
	this.base.GetBeegoController().Ctx.Output.SetStatus(status)
	this.ServeJSON()
}

func (this *WebJsonRender) OnJson(json *response.JsonResult) {
	this.base.SetViewData("json", json)
	this.ServeJSON()
}

func (this *WebJsonRender) OnJsonMap(jsonMap map[string]interface{}) {
	this.base.SetViewData("json", jsonMap)
	this.ServeJSON()
}

func (this *WebJsonRender) OnJsonError(format string, v ...interface{}) {
	this.base.SetErrorState()
	message := fmt.Sprintf(format, v...)
	result := &response.JsonResult{
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.GetCurrentTimeUnix(),
	}
	this.OnJson(result)
}

func (this *WebJsonRender) ServeJSON() {

	cgf := this.base.GetWebConfigs()

	if cgf.CustomJsonEncoder != nil {
		result := this.base.GetData()["json"]
		jsonData, err := cgf.CustomJsonEncoder(result)
		if err != nil {
			this.base.SetViewData("json", &response.JsonResult{
				Message:         fmt.Sprintf("Error json.Encode: %v", err),
				Error:           true,
				CurrentUnixTime: this.GetCurrentTimeUnix(),
			})
			this.base.GetBeegoController().ServeJSON()
		} else {
			this.base.GetBeegoController().Ctx.Output.Header("Content-Type", "application/json")
			this.base.GetBeegoController().Ctx.Output.Body(jsonData)
		}
	} else if cgf.UseJsonPackage {
		result := this.base.GetData()["json"]

		encoder := func(interface{}) ([]byte, error) {
			if cgf.JsonPackageAsCamelCase {
				return json.EncodeAsCamelCase(result)
			}
			return json.Encode(result)
		}

		jsonData, err := encoder(result)
		if err != nil {
			this.base.SetViewData("json", &response.JsonResult{
				Message: fmt.Sprintf("Error json.Encode: %v", err),
				Error:   true, CurrentUnixTime: this.GetCurrentTimeUnix(),
			})
			this.base.GetBeegoController().ServeJSON()
		} else {
			this.base.GetBeegoController().Ctx.Output.Header("Content-Type", "application/json")
			this.base.GetBeegoController().Ctx.Output.Body(jsonData)
		}
	} else {
		this.base.GetBeegoController().ServeJSON()
	}
}

func (this *WebJsonRender) RenderJsonMap(jsonMap map[string]interface{}) {
	this.OnJsonMap(jsonMap)
}

/*
func (this *WebJsonRender) OnRender(data interface{}) {
	switch data.(type) {
	case string:
		this.OnTemplate(data.(string))
		break
	case map[string]interface{}:
		this.OnJsonMap(data.(map[string]interface{}))
		break
	case *response.JsonResult:
		this.OnJson(data.(*response.JsonResult))
		break
	default:
		panic("no render selected")
	}
}*/

func (this *WebJsonRender) RenderJsonError(format string, v ...interface{}) {
	this.OnJsonError(format, v...)
}

func (this *WebJsonRender) OnJsonErrorNotRollback(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&response.JsonResult{Message: message, Error: true, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebJsonRender) OnJsonOk(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&response.JsonResult{Message: message, Error: false, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebJsonRender) OnJson200() {
	this.OnJson(&response.JsonResult{CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebJsonRender) NotFoundAsJson() {
	this.base.SetViewData("json", maps.JSON("message", "not found"))
	this.base.GetBeegoController().Ctx.Output.SetStatus(404)
	this.ServeJSON()

}

func (this *WebJsonRender) OkAsJson(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&response.JsonResult{CurrentUnixTime: this.GetCurrentTimeUnix(), Message: message})
}

func (this *WebJsonRender) OnJsonValidationError() {
	this.base.SetErrorState()
	errors := this.base.GetData()["errors"].(map[string]string)
	this.OnJson(&response.JsonResult{Message: this.base.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebJsonRender) OnJsonValidationWithErrors(errors map[string]string) {
	this.base.SetErrorState()
	this.OnJson(&response.JsonResult{Message: this.base.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebJsonRender) OnJsonValidationMessageWithErrors(message string, errors map[string]string) {
	this.base.SetErrorState()
	this.OnJson(&response.JsonResult{Message: message, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebJsonRender) OnJsonValidationWithResultAndMessageAndErrors(result interface{}, message string, errors map[string]string) {
	this.base.SetErrorState()
	this.OnJson(&response.JsonResult{Message: message, Result: result, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebJsonRender) OnJsonValidationWithResultsAndMessageAndErrors(results interface{}, message string, errors map[string]string) {
	this.base.SetErrorState()
	this.OnJson(&response.JsonResult{Message: message, Results: results, Error: true, Errors: errors, CurrentUnixTime: this.GetCurrentTimeUnix()})
}

func (this *WebJsonRender) RenderJsonValidationError(message string) {
	if this.base.GetWebConfigs().ExitWithHttpCode {
		this.RenderJsonWithStatusCode(maps.JSON("message", message), 400)
	} else {
		this.OnJsonError(message)
	}
}

// permission error
func (this *WebJsonRender) RenderJsonForbidenError(message string, abort bool) {
	if this.base.GetWebConfigs().ExitWithHttpCode {
		this.RenderJsonWithStatusCode(maps.JSON("message", message), 403)
	} else {
		this.OnJsonError(message)
		if abort {
			this.base.GetBeegoController().Abort("403")
		}
	}
}

// logion error
func (this *WebJsonRender) RenderJsonUnauthorizedError(message string, abort bool) {
	if this.base.GetWebConfigs().ExitWithHttpCode {
		this.RenderJsonWithStatusCode(maps.JSON("message", message), 401)
	} else {
		this.OnJsonError(message)
		if abort {
			this.base.GetBeegoController().Abort("401")
		}
	}
}

func (this *WebJsonRender) GetCurrentTimeUnix() int64 {
	return this.GetCurrentTime().Unix()
}

func (this *WebJsonRender) GetCurrentTime() time.Time {
	return time.Now().In(this.base.GetDefaultLocation())
}

func (this *WebJsonRender) renderDadaAsJson(data interface{}) {
	this.base.SetViewData("json", data)
	this.ServeJSON()
}
