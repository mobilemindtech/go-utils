package filters

import (
  _"github.com/astaxie/beego"  
  "github.com/astaxie/beego/context"
)

var FilterMethod = func(ctx *context.Context) {

  if ctx.Input.Query("_method") != "" && ctx.Input.IsPost(){
    ctx.Request.Method = ctx.Input.Query("_method")
  }
}