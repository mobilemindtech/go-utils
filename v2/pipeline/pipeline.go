package pipeline

import (
	"errors"
	_ "fmt"

	"reflect"

	"runtime/debug"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-utils/v2/ctx"
	"github.com/mobilemindtec/go-utils/v2/foreach"
	"github.com/mobilemindtec/go-utils/v2/optional"
	"github.com/mobilemindtec/go-utils/v2/util"
)

type Action = func() interface{}
type ActionR1 = func(interface{}) interface{}

type SuccessFn = func()
type SuccessR1Fn = func(interface{})
type ActionR2 = func(interface{}, interface{}) interface{}
type ActionR3 = func(interface{}, interface{}, interface{}) interface{}
type ActionR4 = func(interface{}, interface{}, interface{}, interface{}) interface{}
type ActionR5 = func(interface{}, interface{}, interface{}, interface{}, interface{}) interface{}

type PipeState int

const (
	StateCreated PipeState = 1 + iota
	StateRunning
	StateSuccess
	StateExit
	StateError
)

type PipeStep struct {
	ctxName    string
	returnLast bool
	name       string
	action     interface{}
	debug      bool
	log        bool
	logMsg     string
	logArgs    []interface{}
}

type Pipe struct {
	errorHandler   interface{}
	exitHandler    func()
	successHandler interface{}
	finallyHandler func()
	startHandler   func()
	results        []interface{}
	steps          []*PipeStep
	ctx            *ctx.Ctx
	State          PipeState
	fail           *optional.Fail
}

func New() *Pipe {
	return &Pipe{steps: []*PipeStep{}, results: []interface{}{}, ctx: ctx.New(), State: StateCreated}
}

func (this *Pipe) IsSuccess() bool {
	return this.State == StateSuccess
}

func (this *Pipe) IsExit() bool {
	return this.State == StateExit
}

func (this *Pipe) GetResults() []interface{} {
	return this.results
}

func (this *Pipe) PutCtx(key string, v interface{}) *Pipe {
	this.ctx.Put(key, v)
	return this
}

func (this *Pipe) AddCtx(v interface{}) *Pipe {
	this.ctx.AddOrdered(v)
	return this
}

func (this *Pipe) GetCtx(key string) interface{} {
	return this.ctx.Get(key)
}

func (this *Pipe) GetOrElseCtx(key string, def interface{}) interface{} {
	if this.HasCtx(key) {
		return this.ctx.Get(key)
	}
	return def
}

func (this *Pipe) HasCtx(key string) bool {
	return this.ctx.Has(key)
}

func (this *Pipe) ErrorHandler(ac interface{}) *Pipe {
	this.errorHandler = ac
	return this
}

func (this *Pipe) ExitHandler(ac func()) *Pipe {
	this.exitHandler = ac
	return this
}

func (this *Pipe) NotFoundHandler(ac func()) *Pipe {
	this.exitHandler = ac
	return this
}

func (this *Pipe) SuccessHandler(ac interface{}) *Pipe {
	this.successHandler = ac
	return this
}

func (this *Pipe) StartHandler(ac func()) *Pipe {
	this.startHandler = ac
	return this
}

func (this *Pipe) FinallyHandler(ac func()) *Pipe {
	this.finallyHandler = ac
	return this
}

func (this *Pipe) Next(ac interface{}) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac})
	return this
}

// execute ac and save return with on ctx with ctxName
func (this *Pipe) NextN(ctxName string, ac interface{}) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac, ctxName: ctxName})
	return this
}

// execute ac with param with last return
func (this *Pipe) NextM(ac interface{}) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac, returnLast: true})
	return this
}

// merge NextN and NextM
func (this *Pipe) NextMN(ctxName string, ac interface{}) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac, returnLast: true, ctxName: ctxName})
	return this
}

func (this *Pipe) Log(msg string, args ...interface{}) *Pipe {
	this.steps = append(this.steps, &PipeStep{log: true, logMsg: msg, logArgs: args})
	return this
}

func (this *Pipe) DebugOn() *Pipe {
	this.steps = append(this.steps, &PipeStep{debug: true})
	return this
}

func (this *Pipe) GetResult() interface{} {
	l := len(this.results)
	if l > 0 {
		return optional.NewSome(this.results[l-1])
	}
	return optional.NewNone()
}

func (this *Pipe) configure() {

	fnErrorHandler := this.errorHandler
	fnExitHandler := this.exitHandler
	fnSuccessHandler := this.successHandler

	this.errorHandler = func(v *optional.Fail) {
		this.State = StateError

		this.fail = v

		if fnErrorHandler != nil {
			switch fnErrorHandler.(type) {
			case util.ErrorFn:
				fnErrorHandler.(util.ErrorFn)(v)
				break
			case util.FailFn:
				fnErrorHandler.(util.FailFn)(v)
				break
			default:
				panic("wrong error handler func")
			}
		}
	}

	this.exitHandler = func() {
		this.State = StateExit
		if fnExitHandler != nil {
			fnExitHandler()
		}
	}

	if fnSuccessHandler == nil {
		this.successHandler = func() {
			this.State = StateSuccess
		}
	} else {
		switch fnSuccessHandler.(type) {
		case SuccessFn:
			this.successHandler = func() {
				this.State = StateSuccess
				fnSuccessHandler.(SuccessFn)()
			}
			break
		case SuccessR1Fn:
			this.successHandler = func(v interface{}) {
				this.State = StateSuccess
				fnSuccessHandler.(SuccessR1Fn)(v)
			}
			break
		default:
			panic("wrong success handler func")
		}
	}

}

func (this *Pipe) addStepResult(step *PipeStep, result interface{}) {
	this.results = append(this.results, result)
	this.AddCtx(result)

	if len(step.ctxName) > 0 {
		this.PutCtx(step.ctxName, result)
	}
}

func (this *Pipe) Run() *Pipe {

	defer func() {

		if r := recover(); r != nil {
			logs.Info("Pipeline recover: %v, StackTrae: %v", r, string(debug.Stack()))

			this.executeErrorHandler(optional.NewFailStr("%v", r))
		}

		if this.finallyHandler != nil {
			this.finallyHandler()
		}

	}()

	this.configure()

	if this.startHandler != nil {
		this.startHandler()
	}

	for i, step := range this.steps {

		if step.log {
			logs.Info(step.logMsg, step.logArgs...)
			continue
		}

		if step.debug {
			logs.Info("Step: %v, results size: %v, results: %v", i, len(this.results), this.results)
			continue
		}

		//logs.Info("ACTION %v", reflect.ValueOf(step.action))

		var r interface{}

		if step.action != nil {

			l := len(this.results)

			switch step.action.(type) {
			case Action:
				r = step.action.(Action)()
				break
			case ActionR1:

				if l < 1 {
					this.executeErrorHandler(optional.NewFailStr("action %v, results len %v", 1, l))
					return this
				}

				var v interface{}

				if step.returnLast {
					lastIdx := len(this.results) - 1
					if lastIdx > -1 {
						v = this.results[lastIdx]
					}
				} else {
					v = this.results[0]
				}
				r = step.action.(ActionR1)(v)
				break

			case ActionR2:

				if l < 2 {
					this.executeErrorHandler(optional.NewFailStr("action r%v, results len %v", 2, l))
					return this
				}

				r = step.action.(ActionR2)(this.results[0], this.results[1])
				break
			case ActionR3:

				if l < 3 {
					this.executeErrorHandler(optional.NewFailStr("action r%v, results len %v", 3, l))
					return this
				}

				r = step.action.(ActionR3)(this.results[0], this.results[1], this.results[2])
				break
			case ActionR4:

				if l < 4 {
					this.executeErrorHandler(optional.NewFailStr("action r%v, results len %v", 4, l))
					return this
				}

				r = step.action.(ActionR4)(this.results[0], this.results[1], this.results[2], this.results[3])
				break
			case ActionR5:

				if l < 5 {
					this.executeErrorHandler(optional.NewFailStr("action r%v, results len %v", 5, l))
					return this
				}

				r = step.action.(ActionR5)(this.results[0], this.results[1], this.results[2], this.results[3], this.results[4])
				break

			default:
				this.executeErrorHandler(optional.NewFailStr("no action set"))
			}

		} else {
			this.executeErrorHandler(optional.NewFailStr("nil action"))
			return this
		}

		if r == nil {
			continue
		}

		switch r.(type) {
		case *Pipe:
			pipe := r.(*Pipe)
			pipe.
				ErrorHandler(this.errorHandler).
				ExitHandler(this.exitHandler).
				Run()

			if pipe.State != StateSuccess {
				return this
			}

			r = pipe.GetResult()
			if some, ok := r.(*optional.Some); ok {
				this.addStepResult(step, some.Item)
			}
			break

		case []*Pipe:
			pipes := r.([]*Pipe)

			for _, pipe := range pipes {
				pipe.
					ErrorHandler(this.errorHandler).
					ExitHandler(this.exitHandler).
					Run()

				if pipe.State != StateSuccess {
					return this
				}

				r = pipe.GetResult()
				if some, ok := r.(*optional.Some); ok {
					this.addStepResult(step, some.Item)
				}
			}
			break

		case *optional.Some:
			result := r.(*optional.Some).Item
			this.addStepResult(step, result)
			break
		case *optional.Fail:
			this.executeErrorHandler(r.(*optional.Fail))
			return this

		case *optional.None:
			this.exitHandler()
			return this
		case *optional.Empty:
			continue
		default:

			if iterable, ok := r.(foreach.Iterable); ok {
				iterable.SetErrorHandler(this.errorHandler)
				iterable.Execute()
			} else {
				this.executeErrorHandler(optional.NewFailStr("action return has be Some | Fail | None. type %v is not allowed", reflect.TypeOf(r)))
				return this
			}

		}
	}

	this.executeSuccessHandler()

	return this
}

func (this *Pipe) executeErrorHandler(v *optional.Fail) {
	this.errorHandler.(util.FailFn)(v)
}

func (this *Pipe) GetFailOrNil() *optional.Fail {
	return this.fail
}

func (this *Pipe) executeSuccessHandler() {
	if this.successHandler != nil {
		switch this.successHandler.(type) {
		case SuccessFn:
			this.successHandler.(SuccessFn)()
			break
		case SuccessR1Fn:
			this.successHandler.(SuccessR1Fn)(this.GetResult())
		}
	}
}

func GetCtx[T any](c interface{}, key string) T {
	switch c.(type) {
	case *Pipe:
		return c.(*Pipe).GetCtx(key).(T)
	case *ctx.Ctx:
		return c.(*ctx.Ctx).Get(key).(T)
	default:
		panic(errors.New("should be Pipe or Ctx"))
	}
}

func GetCtxPtr2[T1 any, T2 any](c interface{}, key1 string, key2 string) (*T1, *T2) {
	return GetCtxPtr[T1](c, key1), GetCtxPtr[T2](c, key2)
}

func GetCtxPtr[T any](c interface{}, key string) *T {

	switch c.(type) {
	case *Pipe:
		return c.(*Pipe).GetCtx(key).(*T)
	case *ctx.Ctx:
		return c.(*ctx.Ctx).Get(key).(*T)
	default:
		panic(errors.New("should be Pipe or PipeCtx"))
	}
}

func IfCtx[T any](p *Pipe, key string, fn func(T)) {
	if p.HasCtx(key) {
		fn(GetCtx[T](p, key))
	}
}

func IfCtxPtr[T any](p *Pipe, key string, fn func(*T)) {
	if p.HasCtx(key) {
		fn(GetCtxPtr[T](p, key))
	}
}

func GetCtxOrNil[T any](p *Pipe, key string) *T {
	var v *T

	if p.HasCtx(key) {
		return GetCtxPtr[T](p, key)
	}

	return v
}
