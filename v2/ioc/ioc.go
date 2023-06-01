package ioc

import (
	"os"
	"reflect"

	"github.com/beego/beego/v2/core/logs"
)

type FactotyFn = func(env string) interface{}

type Initializable interface {
	Init()
}

type Closeable interface {
	Close()
}

type Opts struct {
	EnvKey        string
	SupportedEnvs []string
}

func NewOpts() *Opts {
	return &Opts{
		EnvKey:        "BEEGO_MODE",
		SupportedEnvs: []string{"default", "test", "dev", "stage", "prod"},
	}
}

type Container struct {

	// env/type/value
	data      map[string]map[reflect.Type]interface{}
	factories map[string]map[reflect.Type]FactotyFn
	opts      *Opts
}

func New(opts ...*Opts) *Container {

	data := make(map[string]map[reflect.Type]interface{})
	factories := make(map[string]map[reflect.Type]FactotyFn)
	opt := NewOpts()

	if len(opts) > 0 {
		opt = opts[0]
	} else {
		opt = NewOpts()
	}

	for _, env := range opt.SupportedEnvs {
		data[env] = make(map[reflect.Type]interface{})
		factories[env] = make(map[reflect.Type]FactotyFn)
	}

	return &Container{data: data, factories: factories, opts: opt}
}

func (this *Container) GetCurrEnv() string {
	return os.Getenv(this.opts.EnvKey)
}

func (this *Container) Close() {
	//env := this.GetEnv()
	for _, denv := range this.data {
		for _, v := range denv {
			if m, ok := v.(Closeable); ok {
				m.Close()
			}
		}
	}
}

func (this *Container) AddFactory(t reflect.Type, fn FactotyFn, envs ...string) *Container {

	if len(envs) == 0 {
		envs = append(envs, "default")
	}

	for _, env := range envs {
		this.factories[env][t] = fn
	}

	return this
}

func (this *Container) AddDependency(val interface{}, envs ...string) *Container {
	t := reflect.TypeOf(val)

	if len(envs) == 0 {
		envs = append(envs, "default")
	}

	for _, env := range envs {
		this.data[env][t] = nil
	}

	return this
}

func (this *Container) Get(t reflect.Type) interface{} {

	var typeOf = t
	env := this.GetCurrEnv()

	if t.Kind() == reflect.Ptr {
		typeOf = t.Elem()
	}

	if val, ok := this.data[env][typeOf]; ok {
		return val
	}

	var r interface{}

	if factory, ok := this.factories[env][typeOf]; ok {
		r = factory(env)
	} else {
		r = reflect.New(t).Interface()
	}

	this.Inject(r)

	if v, ok := r.(Initializable); ok {
		v.Init()
	}

	this.data[env][typeOf] = r

	return r
}

func (this *Container) Inject(val interface{}) error {
	valueOf := reflect.ValueOf(val)
	typeOf := reflect.TypeOf(val)

	logs.Debug("valueOf = %v, typeOf  = %v, kind = %v", valueOf, typeOf, typeOf.Kind())

	if typeOf.Kind() == reflect.Ptr {
		valueOf = valueOf.Elem()
		typeOf = valueOf.Type()
	}

	logs.Debug("valueOf = %v, typeOf  = %v, kind = %v", valueOf, typeOf, typeOf.Kind())

	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if _, ok := field.Tag.Lookup("inject"); ok {
			fieldStruct := valueOf.FieldByName(field.Name)
			fieldValue := fieldStruct.Interface()
			logs.Debug("FIELD %v, fieldValue  = %v", field.Name, reflect.TypeOf(fieldValue))
			injected := this.Get(reflect.TypeOf(fieldValue))
			fieldStruct.Set(reflect.ValueOf(injected))
		}
	}

	return nil
}

func Get[T any](ctx *Container) *T {
	var t T
	return ctx.Get(reflect.TypeOf(t)).(*T)
}
