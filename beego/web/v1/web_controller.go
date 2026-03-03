package v1

import (
	"fmt"
	"runtime/debug"

	"github.com/beego/beego/v2/core/logs"
	beego "github.com/beego/beego/v2/server/web"
	"github.com/mobilemindtech/go-utils/beego/web/features"
	"github.com/mobilemindtech/go-utils/beego/web/misc"
)

type WebController struct {
	beego.Controller
	features.WebBase
	// Deprecated, use GetWebConfigs().SetViewPath() instead.
	ViewPath string
}

func init() {
	misc.LoadIl8n()
	misc.LoadFuncs()
}

// Prepare implemented Prepare() method for WebController.
// It's used for language option check and setting.
func (this *WebController) Prepare() {
	this.InitWebBase(&this.Controller)
}

func (this *WebController) Finish() {
	logs.Trace("Finish http call, commit db session")
	this.ReleaseDbSession()
	this.ReleaseCacheService()
	if app, ok := this.AppController.(misc.NestFinisher); ok {
		app.NestFinish()
	}
}

func (this *WebController) Finally() {
	logs.Trace("Finally http call, Rollback db session")
	this.ReleaseDbSessionWithError()
	this.ReleaseCacheService()
}

func (this *WebController) Recover(info interface{}) {
	if app, ok := this.AppController.(misc.NestRecover); ok {
		info := &misc.RecoverInfo{Error: fmt.Sprintf("%v", info), StackTrace: string(debug.Stack())}
		app.NextOnRecover(info)
	}
}
