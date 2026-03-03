package features

import (
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-io/option"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-utils/beego/validator"
	"github.com/mobilemindtech/go-utils/beego/web/misc"
	"github.com/mobilemindtech/go-utils/beego/web/response"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/json"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	"github.com/mobilemindtech/go-utils/v2/maps"
	"github.com/mobilemindtech/go-utils/v2/optional"
)

type WebJsonRenderV2 struct {
	base trait.WebBaseInterface
}

func (this *WebJsonRenderV2) InitWebJsonRenderV2(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebJsonRenderV2) RenderJson(opt interface{}) {

	var dataResult interface{}
	var statusCodeResult = 200

	switch opt.(type) {
	case *criteria.Page:
		dataResult = opt
		break
	case *optional.Some:
		someVal := opt.(*optional.Some).Item

		switch someVal.(type) {
		case *criteria.Page:
			dataResult = someVal
			break
		case *optional.Ok:
			dataResult = map[string]interface{}{}
			break
		default:
			if val, ok := criteria.TryExtractPageIfPegeOf(someVal); ok {
				this.RenderJson(val)
				return
			}
			dataResult = map[string]interface{}{
				"data": someVal,
			}
		}

		break
	case *optional.None:
		statusCodeResult = 404
		dataResult = maps.JSON("error", true, "message", "not found")
		break
	case *response.RawJson:
		dataResult = opt.(*response.RawJson).Value
		break
	case *optional.Fail:

		f := opt.(*optional.Fail)
		err := f.Error
		statusCode := 500

		data := maps.JSON("error", true, "message", fmt.Sprintf("%v", err))

		switch err.(type) {
		case *validator.ValidationError:
			data["validation"] = err.(*validator.ValidationError).Map
			data["validations"] = err.(*validator.ValidationError).List
			statusCode = 400
			break
		default:
			if err.Error() == "validation error" {
				data["validation"] = f.Item
				statusCode = 400
			}
			break
		}
		dataResult = data
		statusCodeResult = statusCode
		break
	case *validator.ValidationError:
		data := maps.JSON("error", true, "message", fmt.Sprintf("%v", opt))
		data["validation"] = opt.(*validator.ValidationError).Map
		data["validations"] = opt.(*validator.ValidationError).List
		dataResult = data
		statusCodeResult = 400
		break
	case *misc.NotFound:
		statusCodeResult = opt.(*misc.NotFound).Code
		dataResult = maps.JSON("error", true, "message", opt.(*misc.NotFound).Error())
		break
	case *misc.ServerError:
		statusCodeResult = opt.(*misc.ServerError).Code
		dataResult = maps.JSON("error", true, "message", opt.(*misc.ServerError).Error())
		break
	case *misc.Unauthorized:
		statusCodeResult = opt.(*misc.Unauthorized).Code
		dataResult = maps.JSON("error", true, "message", opt.(*misc.Unauthorized).Error())
		break
	case *misc.Forbidden:
		statusCodeResult = opt.(*misc.Forbidden).Code
		dataResult = maps.JSON("error", true, "message", opt.(*misc.Forbidden).Error())
		break
	case *misc.BadRequest:
		statusCodeResult = opt.(*misc.BadRequest).Code
		dataResult = maps.JSON("error", true, "message", opt.(*misc.BadRequest).Error(), "errors", opt.(*misc.BadRequest).Errors)
		break
	case error:
		statusCodeResult = 500
		dataResult = maps.JSON("error", true, "message", fmt.Sprintf("%v", opt.(error).Error()))
		break
	default:

		if val, ok := optional.TryExtractValIfOptional(opt); ok {
			this.RenderJson(val)
			return
		}

		if val, ok := criteria.TryExtractPageIfPegeOf(opt); ok {
			this.RenderJson(val)
			return
		}

		if val, ok := opt.(result.IResult); ok {
			if val.HasError() {
				this.RenderJson(val.GetError())
			} else {
				this.RenderJson(val.GetValue())
			}
			return
		}

		dataResult = maps.JSON("data", opt)
		break
	}

	encoder := option.
		Of(this.base.GetWebConfigs().NewJSON).
		Or(json.NewJSON)()

	j, err := encoder.Encode(dataResult)

	//logs.Debug("JSON = %v", string(j))

	if err != nil {
		logs.Error("ERROR JSON ENCODE: %v", err)
		this.base.GetCtx().Output.SetStatus(500)
		this.base.GetCtx().Output.Body([]byte(fmt.Sprint(`{ "error": true, "message": "%v" }`, err.Error())))
	} else {
		logs.Trace("REPONSE STATUS CODE = %v", statusCodeResult)
		this.base.GetCtx().Output.SetStatus(statusCodeResult)
		this.base.GetCtx().Output.Body(j)
	}

	//this.ServeJSON()
}
