package features

import (
	beego "github.com/beego/beego/v2/server/web"
	"github.com/beego/i18n"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type WebMessages struct {
	i18n.Locale
	Flash *beego.FlashData
	base  trait.WebBaseInterface
}

func (this *WebMessages) InitWebMessages(base trait.WebBaseInterface) {
	this.base = base
	this.loadLang()
	this.Flash = beego.NewFlash()
	this.FlashStart()
}

func (this *WebMessages) loadLang() {
	// Reset language option.
	this.Lang = "" // This field is from i18n.Locale.

	// 1. Get language information from 'Accept-Language'.
	al := this.base.GetBeegoController().Ctx.Request.Header.Get("Accept-Language")
	if len(al) > 4 {
		al = al[:5] // Only compare first 5 letters.
		if i18n.IsExist(al) {
			this.Lang = al
		}
	}

	// 2. Default language is English.
	if len(this.Lang) == 0 {
		this.Lang = "pt-BR"
	}
}

func (this *WebMessages) FlashStart() {
	Flash := beego.ReadFromRequest(this.base.GetBeegoController())

	if n, ok := Flash.Data["notice"]; ok {
		this.Flash.Notice(n)
	}

	if n, ok := Flash.Data["error"]; ok {
		this.Flash.Error(n)
	}

	if n, ok := Flash.Data["warning"]; ok {
		this.Flash.Warning(n)
	}

	if n, ok := Flash.Data["success"]; ok {
		this.Flash.Success(n)
	}
}

func (this *WebMessages) FlashEnd(store bool) {
	if store {
		this.Flash.Store(this.base.GetBeegoController())
	} else {
		this.base.SetViewData("Flash", this.Flash.Data)
		this.base.SetViewData("flash", this.Flash.Data)
	}
}

func (this *WebMessages) GetMessage(key string, args ...interface{}) string {
	return i18n.Tr(this.Lang, key, args)
}

func (this *WebMessages) FlashError(msg string, args ...interface{}) {
	this.Flash.Error(msg, args...)
}

func (this *WebMessages) FlashSuccess(msg string, args ...interface{}) {
	this.Flash.Success(msg, args...)
}

func (this *WebMessages) FlashWarn(msg string, args ...interface{}) {
	this.Flash.Warning(msg, args...)
}

func (this *WebMessages) FlashNotice(msg string, args ...interface{}) {
	this.Flash.Notice(msg, args...)
}
