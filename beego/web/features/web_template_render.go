package features

import (
	"fmt"
	"strings"

	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/v2/maps"
	"github.com/mobilemindtech/go-utils/v2/optional"
)

type WebTemplateRender struct {
	base trait.WebBaseInterface
}

func (this *WebTemplateRender) InitWebTemplateRender(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebTemplateRender) OnEntity(viewName string, entity interface{}) {
	this.base.SetViewData("entity", entity)
	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) OnEntityFail(viewName string, entity interface{}, fail *optional.Fail) {
	this.base.SetViewData("entity", entity)

	if fail.Item != nil {
		switch fail.Item.(type) {
		case []map[string]string:
			errors := map[string]string{}
			for _, it := range fail.Item.([]map[string]string) {
				if _, ok := it["field"]; ok {
					errors[it["field"]] = it["message"]
				}
			}
			this.base.SetViewData("errors", errors)
		}
	}

	if strings.Contains(fail.ErrorString(), "validation") {
		this.base.FlashError(this.base.GetMessage("cadastros.validacao"))
	} else {
		this.base.FlashError(fail.ErrorString())
	}

	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) OnEntityError(viewName string, entity interface{}, format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)
	this.base.SetErrorState()
	this.base.FlashError(message)
	this.base.SetViewData("entity", entity)
	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) OnEntities(viewName string, entities interface{}) {
	this.base.SetViewData("entities", entities)
	this.OnTemplate(viewName)
	this.base.FlashEnd(false)
}

func (this *WebTemplateRender) OnEntitiesWithTotalCount(viewName string, entities interface{}, totalCount int64) {
	this.base.SetViewData("entities", entities, "totalCount", totalCount)
	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) OnResult(viewName string, result interface{}) {
	this.base.SetViewData("result", result)
	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) OnResults(viewName string, results interface{}) {
	this.base.SetViewData("results", results)
	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) OnResultsWithTotalCount(viewName string, results interface{}, totalCount int64) {
	this.base.SetViewData("results", results, "totalCount", totalCount)
	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) RenderTemplate(viewName string, data ...interface{}) {
	keyPars := maps.Of[string, interface{}](data...)
	for k, v := range keyPars {
		this.base.SetViewData(k, v)
	}
	this.OnTemplate(viewName)
}

func (this *WebTemplateRender) OnTemplate(viewName string) {
	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) SetTemplate(viewName string) {
	this.base.SetTemplateName(fmt.Sprintf("%s/%s.tpl", this.base.GetViewPath(), viewName))
	this.base.FlashEnd(false)
}

func (this *WebTemplateRender) OnTemplateWithData(viewName string, data map[string]interface{}) {
	if data != nil {
		for k, v := range data {
			this.base.SetViewData(k, v)
		}
	}
	this.SetTemplate(viewName)
}

func (this *WebTemplateRender) OnFullTemplate(tplName string) {
	this.base.SetTemplateName(fmt.Sprintf("%s.tpl", tplName))
	this.base.FlashEnd(false)
}

func (this *WebTemplateRender) OnPureTemplate(templateName string) {
	this.base.SetTemplateName(templateName)
	this.base.FlashEnd(false)
}
