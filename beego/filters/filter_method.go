package filters

import (
  "github.com/beego/beego/v2/server/web/context"
)

var FilterMethod = func(ctx *context.Context) {

  if ctx.Input.Query("_method") != "" && ctx.Input.IsPost(){
    ctx.Request.Method = ctx.Input.Query("_method")
  }
}
