package features

import (
	"github.com/mobilemindtech/go-utils/beego/web/response"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type WebJson struct {
	base trait.WebBaseInterface
}

func (this *WebJson) InitWebJson(base trait.WebBaseInterface) {
	this.base = base
}

func (this *WebJson) NewRawJson(value interface{}) *response.RawJson {
	return &response.RawJson{value}
}
