package features

import (
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-utils/beego/web/response"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type WebRespUtil struct {
	base trait.WebBaseInterface
}

func (this *WebRespUtil) IniWebRespUtil(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebRespUtil) Ok() {
	this.base.GetBeegoController().Ctx.Output.SetStatus(200)
}

func (this *WebRespUtil) OkAsText(message string) {
	this.base.GetBeegoController().Ctx.Output.Body([]byte(message))
}

func (this *WebRespUtil) OnRedirect(action string, args ...interface{}) {
	this.base.FlashEnd(true)
	if this.base.GetBeegoController().Ctx.Input.URL() == action {
		logs.Error("redirect to same URL")
		this.base.GetBeegoController().CustomAbort(500, "redirect to same URL")
	} else {
		this.base.GetBeegoController().Redirect(fmt.Sprintf(action, args...), 302)
	}
}

func (this *WebRespUtil) OnRedirectError(action string, format string, v ...interface{}) {
	this.base.SetErrorState()
	message := fmt.Sprintf(format, v...)
	this.base.FlashError(message)
	this.base.FlashEnd(true)
	if this.base.GetBeegoController().Ctx.Input.URL() == action {
		this.base.GetBeegoController().Abort("500")
	} else {
		this.base.GetBeegoController().Redirect(action, 302)
	}
}

func (this *WebRespUtil) OnRedirectSuccess(action string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.base.FlashSuccess(message)
	this.base.FlashEnd(true)
	if this.base.GetBeegoController().Ctx.Input.URL() == action {
		this.base.GetBeegoController().Abort("500")
	} else {
		this.base.GetBeegoController().Redirect(action, 302)
	}
}

// executes redirect or OnJsonError
func (this *WebRespUtil) OnErrorAny(path string, format string, v ...interface{}) {

	//this.Log("** this.base.IsJson() %v", this.base.IsJson() )
	message := fmt.Sprintf(format, v...)
	if this.base.IsJson() {
		this.base.RenderJsonError(message)
	} else {
		this.OnRedirectError(path, message)
	}
}

// executes redirect or OnJsonOk
func (this *WebRespUtil) OnOkAny(path string, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if this.base.IsJson() {
		this.base.RenderJsonOk(message)
	} else {
		this.base.FlashSuccess(message)
		this.OnRedirect(path)
	}

}

// executes OnEntity or OnJsonValidationError
func (this *WebRespUtil) OnValidationErrorAny(view string, entity interface{}) {

	if this.base.IsJson() {
		this.base.RenderJsonWithValidationError()
	} else {
		this.base.SetErrorState()
		this.base.RenderEntity(view, entity)
	}

}

// executes OnEntity or OnJsonError
func (this *WebRespUtil) OnEntityErrorAny(view string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if this.base.IsJson() {
		this.base.RenderJsonError(message)
	} else {
		this.base.SetErrorState()
		this.base.FlashError(message)
		this.base.RenderEntity(view, entity)
	}

}

// executes OnEntity or OnJsonResultWithMessage
func (this *WebRespUtil) OnEntityAny(view string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	if this.base.IsJson() {
		this.base.RenderJsonResultWithMessage(entity, message)
	} else {
		this.base.FlashSuccess(message)
		this.base.RenderEntity(view, entity)
	}

}

// executes OnResults or OnJsonResults
func (this *WebRespUtil) OnResultsAny(viewName string, results interface{}) {

	if this.base.IsJson() {
		this.base.RenderJsonResults(results)
	} else {
		this.base.RenderResults(viewName, results)
	}

}

// executes  OnResultsWithTotalCount or OnJsonResultsWithTotalCount
func (this *WebRespUtil) OnResultsWithTotalCountAny(viewName string, results interface{}, totalCount int64) {

	if this.base.IsJson() {
		this.base.RenderJsonResults(results, totalCount)
	} else {
		this.base.RenderResults(viewName, results, totalCount)
	}
}

func (this *WebRespUtil) RenderResponse(resp *response.Response) {

	if resp.HasTemplate() {
		resp.ConfigureFlash(this.base.GetFlash())
		this.base.SetTemplateName(resp.GetTemplate(this.base.GetViewPath()))
		this.base.SetViewData("errors", resp.Errors)
		this.base.SetViewData("entity", resp.Entity)
		this.base.SetViewData("entities", resp.Entities)
		this.base.SetViewData("result", resp.Result)
		this.base.SetViewData("results", resp.Results)
		this.base.FlashEnd(false)
	} else {
		// is json result

		// use custom json render
		this.base.
			GetWebConfigs().
			SetUseJsonPackage(resp.JsonPackage).
			SetJsonPackageAsCamelCase(resp.JsonPackageAsCamelCase)

		if resp.JsonResult {
			this.base.SetViewData("json", resp.MkJsonResult())
		} else if resp.HasValue() {
			this.base.RenderJson(resp.Value)
		}

	}
}
