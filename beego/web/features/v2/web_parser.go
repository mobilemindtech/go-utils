package v2

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mobilemindtech/go-utils/app/util"
	"github.com/mobilemindtech/go-utils/beego/web/trait"
)

type Param struct {
	Value string
}

func (this *Param) AsInt() int {
	val, _ := strconv.Atoi(this.Value)
	return val
}
func (this *Param) AsString() string {
	return this.Value
}

func (this *Param) AsInt64() int64 {
	val, _ := strconv.Atoi(this.Value)
	return int64(val)
}

func (this *Param) AsFloat32() float32 {
	val, _ := strconv.ParseFloat(this.Value, 32)
	return float32(val)
}

func (this *Param) AsFloat64() float64 {
	val, _ := strconv.ParseFloat(this.Value, 64)
	return float64(val)
}

func (this *Param) AsBool() bool {
	return this.Value == "true"
}

func (this *Param) AsCheckBox() bool {
	return len(this.Value) > 0
}

func (this *Param) AsTime(layout string) (time.Time, error) {
	return util.DateParse(layout, this.Value)
}

type WebParser struct {
	base trait.WebBaseInterface
}

func (this *WebParser) InitWebParser(base trait.WebBaseInterface) {
	this.base = base

}

func (this *WebParser) GetId() int64 {
	return this.Param(":id").AsInt64()
}

func (this *WebParser) Param(key string) *Param {

	if !strings.HasPrefix(key, ":") {
		key = fmt.Sprintf(":%v", key)
	}

	return &Param{Value: this.base.GetBeegoController().Ctx.Input.Param(key)}
}

func (this *WebParser) Query(key string) *Param {
	return &Param{Value: this.base.GetBeegoController().Ctx.Input.Query(key)}
}

func (this *WebParser) Form(key string) *Param {
	return &Param{Value: this.base.GetBeegoController().Ctx.Input.Query(key)}
}
