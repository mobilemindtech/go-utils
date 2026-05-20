package v2

import (
	"github.com/beego/beego/v2/core/logs"
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

func (this *WebRespUtil) RedirectTo(uri string) {
	this.base.FlashEnd(true)
	if this.base.GetBeegoController().Ctx.Input.URL() == uri {
		logs.Error("redirect to same URL")
		this.base.GetBeegoController().CustomAbort(500, "redirect to same URL")
	} else {
		this.base.GetBeegoController().Redirect(uri, 302)
	}
}

func (this *WebRespUtil) RedirectWithError(uri string, msg string, args ...interface{}) {
	this.base.FlashError(msg, args...)
	this.RedirectTo(uri)
}

func (this *WebRespUtil) RedirectWithSuccess(uri string, msg string, args ...interface{}) {
	this.base.FlashSuccess(msg, args...)
	this.RedirectTo(uri)
}

func (this *WebRespUtil) Respond(ifJson func() interface{}, orElse func()) {
	if this.base.IsJson() {
		this.base.RenderJson(ifJson())
	} else {
		orElse()
	}
}
