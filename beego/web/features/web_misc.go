package features

import (
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-utils/app/util"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/v2/maps"
)

type WebMisc struct {
	base trait.WebBaseInterface
}

func (this *WebMisc) InitWebMisc(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebMisc) DisableXSRF(pathList []string) {
	if os.Getenv("BEEGO_MODE") == "test" {
		this.base.GetBeegoController().EnableXSRF = false
		logs.Trace("DISABLE ALL XSRF IN TEST MODE")
	} else {
		for _, url := range pathList {
			if strings.HasPrefix(this.base.GetBeegoController().Ctx.Input.URL(), url) {
				logs.Debug("disable xsrf for this route: %v", url)
				this.base.GetBeegoController().EnableXSRF = false
			}
		}
	}
}

func (this *WebBase) SetDefaultValuesInDataModel() {
	this.SetViewData(
		"Lang", this.Lang,
		"xsrfdata", template.HTML(this.GetBeegoController().XSRFFormHTML()),
		"dateLayout", util.DateBrLayout,
		"datetimeLayout", util.DateTimeBrLayout,
		"timeLayout", util.TimeMinutesLayout,
		"today", time.Now().In(this.DefaultLocation).Format("02.01.2006"),
	)
}

func (this *WebMisc) GetRawBody() []byte {
	return this.base.GetBeegoController().Ctx.Input.RequestBody
}

func (this *WebMisc) GetBodyAsJson() (*json.Json, error) {
	return json.NewFromBytes(this.GetRawBody())
}

func (this *WebMisc) GetBodyAsJsonResult() *result.Result[*json.Json] {
	return result.Try(this.GetBodyAsJson)
}

func (this *WebMisc) NotFound() {
	this.base.GetBeegoController().Ctx.Output.SetStatus(404)
}

func (this *WebMisc) ServerError() {
	this.base.GetBeegoController().Ctx.Output.SetStatus(500)
}
func (this *WebMisc) BadRequest() {
	this.base.GetBeegoController().Ctx.Output.SetStatus(400)
}

func (this *WebMisc) Unauthorized() {
	this.base.GetBeegoController().Ctx.Output.SetStatus(401)
}

func (this *WebMisc) Forbidden() {
	this.base.GetBeegoController().Ctx.Output.SetStatus(403)
}

func (this *WebMisc) HasUriPath(paths ...string) bool {
	for _, it := range paths {
		if strings.HasPrefix(this.base.GetBeegoController().Ctx.Input.URL(), it) {
			return true
		}
	}
	return false
}

func (this *WebMisc) IsJson() bool {
	return this.base.GetBeegoController().Ctx.Input.AcceptsJSON() || this.base.GetBeegoController().Ctx.Input.Header("Content-Type") == "application/json"
}

func (this *WebMisc) IsAjax() bool {
	return this.base.GetBeegoController().Ctx.Input.IsAjax()
}

func (this *WebMisc) GetHeaderByName(name string) string {
	return this.base.GetBeegoController().Ctx.Request.Header.Get(name)
}

func (this *WebMisc) GetHeaderByNames(names ...string) string {

	for _, name := range names {
		val := this.base.GetBeegoController().Ctx.Request.Header.Get(name)

		if len(val) > 0 {
			return val
		}
	}

	return ""
}

func (this *WebMisc) GetLastUpdate() time.Time {
	lastUpdateUnix, _ := this.base.GetBeegoController().GetInt64("lastUpdate")
	var lastUpdate time.Time

	if lastUpdateUnix > 0 {
		lastUpdate = time.Unix(lastUpdateUnix, 0).In(this.base.GetDefaultLocation())
	}

	return lastUpdate
}

/*
 */

func (this *WebMisc) PreRender(ret interface{}) {
	if app, ok := this.base.GetBeegoController().AppController.(beego.PreRender); ok {
		app.PreRender(ret)
	}
}

func (this *WebMisc) SetViewData(values ...interface{}) *WebMisc {

	if len(values)%2 > 0 {
		panic("expect key/pair values")
	}

	beegoData := this.base.GetBeegoController().Data
	data := maps.Of[string, interface{}](values...)

	for k, v := range data {
		beegoData[k] = v
	}

	return this
}
