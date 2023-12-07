package ctx

import (
	"fmt"

	"github.com/beego/beego/v2/core/logs"
)

type Ctx struct {
	data  map[string]interface{}
	index int
}

func New() *Ctx {
	return &Ctx{data: map[string]interface{}{}}
}

func (this *Ctx) Dump() {
	logs.Debug("====> CTX DUMP START")
	for k, v := range this.data {
		logs.Debug("==> %v = %v", k, v)
	}
	logs.Debug("====> CTX DUMP END")

}

func (this *Ctx) Put(key string, v interface{}) *Ctx {
	this.data[key] = v
	return this
}

func (this *Ctx) AddOrdered(v interface{}) *Ctx {
	key := fmt.Sprintf("$%v", this.index)
	this.index += 1
	this.data[key] = v
	return this
}

func (this *Ctx) Get(key string) interface{} {
	if v, ok := this.data[key]; ok {
		return v
	}
	return nil
}

func (this *Ctx) Has(key string) bool {
	_, ok := this.data[key]
	return ok
}
