package v1

import (
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-utils/beego/web/response"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	"github.com/mobilemindtech/go-utils/v2/maps"
	"github.com/mobilemindtech/go-utils/v2/optional"
)

// WebResponses Deprecated.
type WebJsonRenderV1 struct {
	base trait.WebBaseInterface
}

func (this *WebJsonRenderV1) InitWebJsonRenderV1(base trait.WebBaseInterface) {
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

func (this *WebJsonRenderV1) RenderJsonResult(opt interface{}) {

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

func (this *WebJsonRenderV1) OnJsonResultNil() {
	this.OnJsonResult(nil)
}

func (this *WebJsonRenderV1) OnJsonResult(result interface{}) {
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Error:           false,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRenderV1) GetJsonResult() (*response.JsonResult, bool) {
	if this.base.GetData()["json"] != nil {
		if j, ok := this.base.GetData()["json"].(*response.JsonResult); ok {
			return j, ok
		}
	}
	return nil, false
}

func (this *WebJsonRenderV1) OnJsonMessage(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.renderDadaAsJson(&response.JsonResult{
		Message:         message,
		Error:           false,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRenderV1) OnJsonResultError(result interface{}, format string, v ...interface{}) {
	this.base.SetErrorState()
	message := fmt.Sprintf(format, v...)
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRenderV1) OnJsonResultWithMessage(result interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Error:           false,
		Message:         message,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRenderV1) OnJsonResults(results interface{}) {
	this.renderDadaAsJson(&response.JsonResult{
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRenderV1) OnJsonResultAndResults(result interface{}, results interface{}) {
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRenderV1) OnJsonResultsWithTotalCount(results interface{}, totalCount int64) {
	this.renderDadaAsJson(&response.JsonResult{
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
		TotalCount:      totalCount,
	})
}

func (this *WebJsonRenderV1) OnJsonPage(page *criteria.Page) {
	this.renderDadaAsJson(&response.JsonResult{
		Results:         page.Data,
		Error:           false,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
		TotalCount:      int64(page.TotalCount),
	})
}

func (this *WebJsonRenderV1) OnJsonResultAndResultsWithTotalCount(result interface{}, results interface{}, totalCount int64) {
	this.renderDadaAsJson(&response.JsonResult{
		Result:          result,
		Results:         results,
		Error:           false,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
		TotalCount:      totalCount,
	})
}

func (this *WebJsonRenderV1) OnJsonResultsError(results interface{}, format string, v ...interface{}) {
	this.base.SetErrorState()
	message := fmt.Sprintf(format, v...)
	this.renderDadaAsJson(&response.JsonResult{
		Results:         results,
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
	})
}

func (this *WebJsonRenderV1) RenderJsonWithStatusCode(data interface{}, status int) {
	this.base.SetViewData("json", data)
	this.base.GetBeegoController().Ctx.Output.SetStatus(status)
	this.ServeJSON()
}

func (this *WebJsonRenderV1) OnJson(json *response.JsonResult) {
	this.base.SetViewData("json", json)
	this.ServeJSON()
}

func (this *WebJsonRenderV1) OnJsonMap(jsonMap map[string]interface{}) {
	this.base.SetViewData("json", jsonMap)
	this.ServeJSON()
}

func (this *WebJsonRenderV1) OnJsonError(format string, v ...interface{}) {
	this.base.SetErrorState()
	message := fmt.Sprintf(format, v...)
	result := &response.JsonResult{
		Message:         message,
		Error:           true,
		CurrentUnixTime: this.base.GetCurrentTimeUnix(),
	}
	this.OnJson(result)
}

func (this *WebJsonRenderV1) ServeJSON() {

	cgf := this.base.GetWebConfigs()

	if cgf.CustomJsonEncoder != nil {
		result := this.base.GetData()["json"]
		jsonData, err := cgf.CustomJsonEncoder(result)
		if err != nil {
			this.base.SetViewData("json", &response.JsonResult{
				Message:         fmt.Sprintf("Error json.Encode: %v", err),
				Error:           true,
				CurrentUnixTime: this.base.GetCurrentTimeUnix(),
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
				Error:   true, CurrentUnixTime: this.base.GetCurrentTimeUnix(),
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

func (this *WebJsonRenderV1) RenderJsonMap(jsonMap map[string]interface{}) {
	this.OnJsonMap(jsonMap)
}

/*
func (this *WebJsonRenderV1) OnRender(data interface{}) {
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

func (this *WebJsonRenderV1) RenderJsonError(format string, v ...interface{}) {
	this.OnJsonError(format, v...)
}

func (this *WebJsonRenderV1) OnJsonErrorNotRollback(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&response.JsonResult{Message: message, Error: true, CurrentUnixTime: this.base.GetCurrentTimeUnix()})
}

func (this *WebJsonRenderV1) OnJsonOk(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&response.JsonResult{Message: message, Error: false, CurrentUnixTime: this.base.GetCurrentTimeUnix()})
}

func (this *WebJsonRenderV1) OnJson200() {
	this.OnJson(&response.JsonResult{CurrentUnixTime: this.base.GetCurrentTimeUnix()})
}

func (this *WebJsonRenderV1) NotFoundAsJson() {
	this.base.SetViewData("json", maps.JSON("message", "not found"))
	this.base.GetBeegoController().Ctx.Output.SetStatus(404)
	this.ServeJSON()

}

func (this *WebJsonRenderV1) OkAsJson(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.OnJson(&response.JsonResult{CurrentUnixTime: this.base.GetCurrentTimeUnix(), Message: message})
}

func (this *WebJsonRenderV1) OnJsonValidationError() {
	this.base.SetErrorState()
	errors := this.base.GetData()["errors"].(map[string]string)
	this.OnJson(&response.JsonResult{Message: this.base.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.base.GetCurrentTimeUnix()})
}

func (this *WebJsonRenderV1) OnJsonValidationWithErrors(errors map[string]string) {
	this.base.SetErrorState()
	this.OnJson(&response.JsonResult{Message: this.base.GetMessage("cadastros.validacao"), Error: true, Errors: errors, CurrentUnixTime: this.base.GetCurrentTimeUnix()})
}

func (this *WebJsonRenderV1) OnJsonValidationMessageWithErrors(message string, errors map[string]string) {
	this.base.SetErrorState()
	this.OnJson(&response.JsonResult{Message: message, Error: true, Errors: errors, CurrentUnixTime: this.base.GetCurrentTimeUnix()})
}

func (this *WebJsonRenderV1) OnJsonValidationWithResultAndMessageAndErrors(result interface{}, message string, errors map[string]string) {
	this.base.SetErrorState()
	this.OnJson(&response.JsonResult{Message: message, Result: result, Error: true, Errors: errors, CurrentUnixTime: this.base.GetCurrentTimeUnix()})
}

func (this *WebJsonRenderV1) OnJsonValidationWithResultsAndMessageAndErrors(results interface{}, message string, errors map[string]string) {
	this.base.SetErrorState()
	this.OnJson(&response.JsonResult{Message: message, Results: results, Error: true, Errors: errors, CurrentUnixTime: this.base.GetCurrentTimeUnix()})
}

func (this *WebJsonRenderV1) RenderJsonValidationError(message string) {
	if this.base.GetWebConfigs().ExitWithHttpCode {
		this.RenderJsonWithStatusCode(maps.JSON("message", message), 400)
	} else {
		this.OnJsonError(message)
	}
}

// permission error
func (this *WebJsonRenderV1) RenderJsonForbidenError(message string, abort bool) {
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
func (this *WebJsonRenderV1) RenderJsonUnauthorizedError(message string, abort bool) {
	if this.base.GetWebConfigs().ExitWithHttpCode {
		this.RenderJsonWithStatusCode(maps.JSON("message", message), 401)
	} else {
		this.OnJsonError(message)
		if abort {
			this.base.GetBeegoController().Abort("401")
		}
	}
}

func (this *WebJsonRenderV1) renderDadaAsJson(data interface{}) {
	this.base.SetViewData("json", data)
	this.ServeJSON()
}
