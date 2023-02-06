package pipeline

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-utils/v2/optional"
)

type Action func() interface{}
type ActionR1 func(interface{}) interface{}
type ActionR2 func(interface{}, interface{}) interface{}
type ActionR3 func(interface{}, interface{}, interface{}) interface{}
type ActionR4 func(interface{}, interface{}, interface{}, interface{}) interface{}
type ActionR5 func(interface{}, interface{}, interface{}, interface{}, interface{}) interface{}

type PipeCtx struct {
	data  map[string]interface{}
	index int
}

func NewCtx() *PipeCtx {
	return &PipeCtx{data: map[string]interface{}{}}
}

func (this *PipeCtx) Put(key string, v interface{}) *PipeCtx {
	this.data[key] = v
	return this
}

func (this *PipeCtx) Add(v interface{}) *PipeCtx {
	key := fmt.Sprintf("$%v", this.index)
	this.index += 1
	this.data[key] = v
	return this
}

func (this *PipeCtx) Get(key string) interface{} {
	if v, ok := this.data[key]; ok {
		return v
	}
	return nil
}

func (this *PipeCtx) Has(key string) bool {
	_, ok := this.data[key]
	return ok
}

type PipeStep struct {
	name    string
	action  interface{}
	debug   bool
	log     bool
	logMsg  string
	logArgs []interface{}
}

type Pipe struct {
	onError   func(*optional.Fail)
	onExit    func()
	onSuccess func()
	Results   []interface{}
	Result    interface{}
	steps     []*PipeStep
	Ctx       *PipeCtx
}

func New() *Pipe {
	return &Pipe{steps: []*PipeStep{}, Results: []interface{}{}, Ctx: NewCtx()}
}

func (this *Pipe) PutCtx(key string, v interface{}) *Pipe {
	this.Ctx.Put(key, v)
	return this
}

func (this *Pipe) AddCtx(v interface{}) *Pipe {
	this.Ctx.Add(v)
	return this
}

func (this *Pipe) GetCtx(key string) interface{} {
	return this.Ctx.Get(key)
}

func (this *Pipe) HasCtx(key string) bool {
	return this.Ctx.Has(key)
}

func (this *Pipe) OnError(ac func(*optional.Fail)) *Pipe {
	this.onError = ac
	return this
}

func (this *Pipe) OnExit(ac func()) *Pipe {
	this.onExit = ac
	return this
}

func (this *Pipe) OnSuccess(ac func()) *Pipe {
	this.onSuccess = ac
	return this
}

func (this *Pipe) Next(ac Action) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac})
	return this
}

func (this *Pipe) NextR1(ac ActionR1) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac})
	return this
}

func (this *Pipe) NextR2(ac ActionR2) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac})
	return this
}

func (this *Pipe) NextR3(ac ActionR3) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac})
	return this
}

func (this *Pipe) NextR4(ac ActionR4) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac})
	return this
}

func (this *Pipe) NextR5(ac ActionR5) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac})
	return this
}

func (this *Pipe) Log(msg string, args ...interface{}) *Pipe {
	this.steps = append(this.steps, &PipeStep{log: true, logMsg: msg, logArgs: args})
	return this
}

func (this *Pipe) Debug() *Pipe {
	this.steps = append(this.steps, &PipeStep{debug: true})
	return this
}

func (this *Pipe) Run() {

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
			this.onError(optional.NewFailStr("%v", r))
		}
	}()

	for i, step := range this.steps {

		if step.log {
			logs.Info(step.logMsg, step.logArgs...)
			continue
		}

		if step.debug {
			logs.Info("Step: %v, Results: %v", i, len(this.Results))
			for j, r := range this.Results {
				logs.Info("Result %v: %v", j, r)
			}
			continue
		}

		//logs.Info("ACTION %v", reflect.ValueOf(step.action))

		var r interface{}

		if step.action != nil {

			l := len(this.Results)

			switch step.action.(type) {
			case Action:
				r = step.action.(Action)()
				break
			case ActionR1:

				if l < 1 {
					this.onError(optional.NewFailStr("action r%v, results len %v", 1, l))
					return
				}

				r = step.action.(ActionR1)(this.Results[0])
				break
			case ActionR2, func(interface{}, interface{}) interface{}:

				if l < 2 {
					this.onError(optional.NewFailStr("action r%v, results len %v", 2, l))
					return
				}

				r = step.action.(ActionR2)(this.Results[0], this.Results[1])
				break
			case ActionR3, func(interface{}, interface{}, interface{}) interface{}:

				if l < 3 {
					this.onError(optional.NewFailStr("action r%v, results len %v", 3, l))
					return
				}

				r = step.action.(ActionR3)(this.Results[0], this.Results[1], this.Results[2])
				break
			case ActionR4, func(interface{}, interface{}, interface{}, interface{}) interface{}:

				if l < 4 {
					this.onError(optional.NewFailStr("action r%v, results len %v", 4, l))
					return
				}

				r = step.action.(ActionR4)(this.Results[0], this.Results[1], this.Results[2], this.Results[3])
				break
			case ActionR5, func(interface{}, interface{}, interface{}, interface{}, interface{}) interface{}:

				if l < 5 {
					this.onError(optional.NewFailStr("action r%v, results len %v", 5, l))
					return
				}

				r = step.action.(ActionR5)(this.Results[0], this.Results[1], this.Results[2], this.Results[3], this.Results[4])
				break
			default:
				this.onError(optional.NewFailStr("no action set"))
			}

		} else {
			this.onError(optional.NewFailStr("nil action"))
			return
		}

		if r == nil {
			continue
		}

		switch r.(type) {
		case *optional.Some:
			this.Result = r.(*optional.Some).Item
			this.Results = append(this.Results, this.Result)
			this.AddCtx(this.Result)
			break
		case *optional.Fail:
			this.onError(r.(*optional.Fail))
			return

		case *optional.None:
			this.onExit()
			return
		case *optional.Empty:
			continue
		default:
			this.onError(optional.NewFailStr("action return has be Some | Fail | None. type %v is not allowed", reflect.TypeOf(r)))
			return
		}
	}

	if this.onSuccess != nil {
		this.onSuccess()
	}
}

func GetCtx[T any](ctx interface{}, key string) T {
	switch ctx.(type) {
	case *Pipe:
		return ctx.(*Pipe).GetCtx(key).(T)
	case *PipeCtx:
		return ctx.(*PipeCtx).Get(key).(T)
	default:
		panic(errors.New("should be Pipe or PipeCtx"))
	}
}

func GetCtxPtr[T any](ctx interface{}, key string) *T {

	switch ctx.(type) {
	case *Pipe:
		return ctx.(*Pipe).GetCtx(key).(*T)
	case *PipeCtx:
		return ctx.(*PipeCtx).Get(key).(*T)
	default:
		panic(errors.New("should be Pipe or PipeCtx"))
	}
}
