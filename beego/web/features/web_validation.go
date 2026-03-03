package features

import (
	"github.com/beego/beego/v2/core/validation"
	"github.com/mobilemindtech/go-utils/beego/validator"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type WebValidation struct {
	EntityValidator *validator.EntityValidator
	base            trait.WebBaseInterface
}

func (this *WebValidation) InitWebValidation(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebValidation) Validate(entity interface{}, f ...func(validator *validation.Validation)) bool {

	var fn func(validator *validation.Validation) = nil

	if len(f) > 0 {
		fn = f[0]
	}

	result, _ := this.EntityValidator.IsValid(entity, fn)

	if result.HasError {
		this.base.GetFlash().Error(this.base.GetMessage("cadastros.validacao"))
		this.EntityValidator.CopyErrorsToView(result, this.base.GetData())
	}

	return result.HasError == false

}

// Deprecated: use Validate
func (this *WebValidation) OnValidate(entity interface{}, f func(validator *validation.Validation)) bool {
	return this.Validate(entity, f)
}
