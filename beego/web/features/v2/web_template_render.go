package v2

import (
	"fmt"
	"strings"

	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type WebTemplateRender struct {
	base trait.WebBaseInterface
}

func (this *WebTemplateRender) InitWebTemplateRender(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebTemplateRender) RenderTemplate(tplName string, data ...interface{}) trait.WebBaseInterface {
	tpl := tplName
	if !strings.Contains(tplName, "/") {
		tpl = fmt.Sprintf("%s/%s.tpl", this.base.GetViewPath(), tplName)
	}
	this.base.SetTemplateName(tpl)
	this.base.FlashEnd(false)
	this.base.SetViewData(data...)
	return this.base
}
