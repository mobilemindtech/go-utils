package v2

import (
	"fmt"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-io/option"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-io/types/unit"
	"github.com/mobilemindtech/go-utils/beego/validator"
	"github.com/mobilemindtech/go-utils/beego/web/response"
	"github.com/mobilemindtech/go-utils/beego/web/resps"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
	"github.com/mobilemindtech/go-utils/fault"
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

	mkValidationError := func(v *validator.ValidationError) {
		data := maps.JSON("error", true, "message", fmt.Sprintf("%v", v))
		data["validation"] = opt.(*validator.ValidationError).Map
		data["validations"] = opt.(*validator.ValidationError).List
		data["errors"] = opt.(*validator.ValidationError).List
		dataResult = data
		statusCodeResult = 400
	}

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
			mkValidationError(err.(*validator.ValidationError))
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
		mkValidationError(opt.(*validator.ValidationError))
		break
	case *resps.NotFound:
		statusCodeResult = opt.(*resps.NotFound).StatusCode()
		dataResult = maps.JSON("error", true, "message", opt.(*resps.NotFound).Error())
		break
	case *resps.ServerError:
		statusCodeResult = opt.(*resps.ServerError).StatusCode()
		dataResult = maps.JSON("error", true, "message", opt.(*resps.ServerError).Error())
		break
	case *resps.Unauthorized:
		statusCodeResult = opt.(*resps.Unauthorized).StatusCode()
		dataResult = maps.JSON("error", true, "message", opt.(*resps.Unauthorized).Error())
		break
	case *resps.Forbidden:
		statusCodeResult = opt.(*resps.Forbidden).StatusCode()
		dataResult = maps.JSON("error", true, "message", opt.(*resps.Forbidden).Error())
		break
	case *resps.BadRequest:
		statusCodeResult = opt.(*resps.BadRequest).StatusCode()
		dataResult = maps.JSON("error", true, "message", opt.(*resps.BadRequest).Error(), "errors", opt.(*resps.BadRequest).Errors)
		break
	case *resps.Ok:
		statusCodeResult = 200
		dataResult = opt.(*resps.Ok).Body
		break
	case *resps.Response:
		statusCodeResult = opt.(*resps.Response).StatusCode()
		dataResult = opt.(*resps.Response).Body()
		break
	case *fault.UserNotFoundError:
		statusCodeResult = 401
		v := opt.(*fault.UserNotFoundError)
		dataResult = maps.JSON("error", true, "message", v.Message)
		break
	case *fault.UserUnauthorizedError:
		statusCodeResult = 403
		v := opt.(*fault.UserUnauthorizedError)
		dataResult = maps.JSON("error", true, "message", v.Message)
		break
	case *fault.EntityNotFound:
		statusCodeResult = 404
		v := opt.(*fault.EntityNotFound)
		dataResult = maps.JSON("error", true, "message", v.Message)
		break
	case *fault.ValidationError:
		statusCodeResult = 400
		v := opt.(*fault.ValidationError)
		dataResult = maps.JSON("error", true, "message", v.Message, "errors", v.Errors)
		break
	case *unit.Unit:
		statusCodeResult = 200
		dataResult = maps.JSON("message", "successs")
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

		if val, ok := opt.(option.IOption); ok {
			if val.IsEmpty() {
				this.RenderJson(resps.NewNotFound())
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

	if err != nil {
		logs.Error("error on json encode: %v", err)
		this.base.GetCtx().Output.SetStatus(500)
		msg := fmt.Sprintf(`{ "error": true, "message": "%v" }`, err.Error())
		this.base.GetCtx().Output.Body([]byte(msg))
	} else {
		logs.Trace("responds with status code %v", statusCodeResult)
		this.base.GetCtx().Output.SetStatus(statusCodeResult)
		this.base.GetCtx().Output.Body(j)
	}
}

func (this *WebJsonRenderV2) RenderJsonWithValidationError() {
	this.base.SetErrorState()
	errors := this.base.GetData()["errors"].(map[string]string)
	var messages []string
	for k, v := range errors {
		messages = append(messages, fmt.Sprintf("%s: %s", k, v))
	}
	this.RenderJson(resps.NewBadRequest().WithErrors(messages))
}

func (this *WebJsonRenderV2) RenderJsonPage(results interface{}, totalCount int64) {
	this.RenderJson(&criteria.Page{Data: results, TotalCount: int(totalCount)})
}
