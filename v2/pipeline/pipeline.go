package pipeline

import (
	"errors"
	"fmt"
	"reflect"

	"runtime/debug"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtech/go-io/result"
	"github.com/mobilemindtech/go-utils/v2/criteria"
	"github.com/mobilemindtech/go-utils/v2/ctx"
	"github.com/mobilemindtech/go-utils/v2/fn"
	"github.com/mobilemindtech/go-utils/v2/foreach"
	"github.com/mobilemindtech/go-utils/v2/optional"
)

type PipeState int

const (
	StateCreated PipeState = 1 + iota
	//StateRunning
	StateSuccess
	StateExit
	StateError
)

type Continue struct{}

func NewContinue() *Continue {
	return &Continue{}
}
func FailOrContinue(err error) interface{} {
	if err != nil {
		return optional.NewFail(err)
	}
	return NewContinue()
}

type PipeStep struct {
	ctxName     string
	returnLast  bool
	typeResolve bool
	name        string
	action      interface{}
	debug       bool
	log         bool
	logMsg      string
	logArgs     []interface{}
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
	debug          bool
}

type CtxItem struct {
	Key  string
	Item interface{}
}

func NewCtxItem(key string, val interface{}) *CtxItem {
	return &CtxItem{key, val}
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

func (this *Pipe) IsFail() bool {
	return this.State == StateError
}

func (this *Pipe) GetError() error {
	if this.IsFail() {
		return this.fail.Error
	}
	return nil
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

func (this *Pipe) IfCtx(key string, fn func(interface{})) {
	if this.ctx.Has(key) {
		fn(this.GetCtx(key))
	}
}

func (this *Pipe) CtxDump() {
	this.ctx.Dump()
}

func (this *Pipe) OnError(ac interface{}) *Pipe {
	return this.ErrorHandler(ac)
}
func (this *Pipe) ErrorHandler(ac interface{}) *Pipe {
	this.errorHandler = ac
	return this
}

func (this *Pipe) OnExit(ac func()) *Pipe {
	return this.ExitHandler(ac)
}

func (this *Pipe) OnNotFound(ac func()) *Pipe {
	return this.ExitHandler(ac)
}

func (this *Pipe) ExitHandler(ac func()) *Pipe {
	this.exitHandler = ac
	return this
}

func (this *Pipe) NotFoundHandler(ac func()) *Pipe {
	this.exitHandler = ac
	return this
}

func (this *Pipe) OnSuccess(ac interface{}) *Pipe {
	return this.SuccessHandler(ac)
}

func (this *Pipe) SuccessHandler(ac interface{}) *Pipe {
	this.successHandler = ac
	return this
}

func (this *Pipe) OnStart(ac func()) *Pipe {
	return this.StartHandler(ac)
}
func (this *Pipe) StartHandler(ac func()) *Pipe {
	this.startHandler = ac
	return this
}

func (this *Pipe) OnFinally(ac func()) *Pipe {
	return this.FinallyHandler(ac)
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

/**
 * Execute and resolve args types
 */
func (this *Pipe) NextR(ac interface{}) *Pipe {
	this.steps = append(this.steps, &PipeStep{action: ac, typeResolve: true})
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

func (this *Pipe) MergeRight(pipe *Pipe) *Pipe {

	this.ctx = pipe.ctx

	for _, step := range pipe.steps {
		this.steps = append(this.steps, step)
	}
	return this
}

func (this *Pipe) MergeLeft(pipe *Pipe) *Pipe {

	this.ctx = pipe.ctx

	newSteps := []*PipeStep{}

	for _, step := range pipe.steps {
		newSteps = append(newSteps, step)
	}

	for _, step := range this.steps {
		newSteps = append(newSteps, step)
	}

	this.steps = newSteps
	return this
}

func (this *Pipe) Log(msg string, args ...interface{}) *Pipe {
	this.steps = append(this.steps, &PipeStep{log: true, logMsg: msg, logArgs: args})
	return this
}

func (this *Pipe) DebugOn() *Pipe {
	this.debug = true
	return this
}

func (this *Pipe) SetDebug(b bool) *Pipe {
	this.debug = b
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
			info := fn.NewFuncInfo(fnErrorHandler)

			if info.ArgsCount == 0 { // onSuccess = func()
				info.CallEmpty()
			} else {

				if info.ArgsCount != 1 { // onSuccess = func(interface{})
					panic("step OnError: func must have one argument of error or Fail")
				}

				args := []reflect.Value{}
				if info.HasTypedArgs() {
					if info.ArgsTypes[0] == reflect.TypeOf(v.Error) {
						args = append(args, reflect.ValueOf(v.Error))
					} else {
						args = append(args, reflect.ValueOf(v))
					}
				} else {
					args = append(args, reflect.ValueOf(v))
				}

				info.Call(args)
			}

		}
	}

	this.exitHandler = func() {
		this.State = StateExit
		if fnExitHandler != nil {
			fnExitHandler()
		} else {
			panic("exit handler not found")
		}
	}

	this.successHandler = func() {
		this.State = StateSuccess

		if fnSuccessHandler != nil {
			fnSuccessInfo := fn.NewFuncInfo(fnSuccessHandler)

			if fnSuccessInfo.ArgsCount == 0 { // onSuccess = func()
				fnSuccessInfo.CallEmpty()
			} else {

				args := []reflect.Value{}

				if !fnSuccessInfo.HasTypedArgs() { // onSuccess = func(interface{})

					if fnSuccessInfo.ArgsCount > 1 {
						panic("step OnSucess: func should be only 1 arg")
					}

					v := this.GetResult()
					args = append(args, reflect.ValueOf(v))

				} else { // onSucess = func(Type1, Type2, ...)
					for _, typ := range fnSuccessInfo.ArgsTypes {
						args = append(args, this.findInCtxbyType(typ, "OnSuccess"))
					}
				}

				logs.Debug("success call with args: %v", len(args))
				fnSuccessInfo.Call(args)
			}
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

func (this *Pipe) findInCtxbyType(argType reflect.Type, step string) reflect.Value {
	for _, it := range this.results {
		rtype := reflect.TypeOf(it)
		if rtype == argType {
			return reflect.ValueOf(it)
		}
	}
	panic(fmt.Sprintf("step %v: arg type %v not found in results", step, argType))
}

func (this *Pipe) Run() *Pipe {

	currentStep := -1

	defer func() {

		if r := recover(); r != nil {
			logs.Error("Pipeline recover on step %v. Message %v. StackTrae: %v", currentStep, r, string(debug.Stack()))

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

		if step.action == nil {
			panic(fmt.Sprintf("step %v: no action setted", i))
		}

		if step.log {
			logs.Info(step.logMsg, step.logArgs...)
			continue
		}

		currentStep = i

		nextFnInfo := fn.NewFuncInfo(step.action)
		step.typeResolve = nextFnInfo.HasTypedArgs()

		if this.debug {
			logs.Info("step: %v, results: %v, typeResolve: %v", i, len(this.results), step.typeResolve)
		}

		fnParams := []reflect.Value{}
		resultsCount := len(this.results)
		var r interface{}

		handleResult := func(res []reflect.Value) {
			if len(res) > 1 {
				panic(fmt.Sprintf("step %v: return type count should be 0 or 1", i))
			}

			if len(res) > 0 {
				r = res[0].Interface()
			}
		}

		switch nextFnInfo.ArgsCount {
		case 0:
			//r = step.action.(Action)()
			handleResult(nextFnInfo.Call(fnParams))
			break
		case 1:

			if resultsCount < 1 {
				panic(fmt.Sprintf("step %v: not result found", i))
			}

			var v reflect.Value

			if step.returnLast {
				lastIdx := len(this.results) - 1
				v = reflect.ValueOf(this.results[lastIdx])
			} else if step.typeResolve {
				v = this.findInCtxbyType(nextFnInfo.ArgType(0), fmt.Sprintf("%v", i))
			} else {
				v = reflect.ValueOf(this.results[0])
			}

			fnParams = append(fnParams, v)
			handleResult(nextFnInfo.Call(fnParams))

			//r = step.action.(ActionR1)(v)
			break

		default:

			if resultsCount < nextFnInfo.ArgsCount {
				panic(fmt.Sprintf(
					"step %v: action args %v but results %v", nextFnInfo.ArgsCount, resultsCount))
			}

			//r = step.action.(ActionR2)(this.results[0], this.results[1])
			for i = 0; i < nextFnInfo.ArgsCount; i++ {
				if step.typeResolve {
					fnParams = append(fnParams, this.findInCtxbyType(nextFnInfo.ArgType(i), fmt.Sprintf("%v", i)))
				} else {
					fnParams = append(fnParams, reflect.ValueOf(this.results[i]))
				}
			}
			handleResult(nextFnInfo.Call(fnParams))
			break
		}

		if r == nil {
			continue
		}

		switch r.(type) {
		case *CtxItem:
			step.ctxName = r.(*CtxItem).Key
			r = r.(*CtxItem).Item
			break
		case CtxItem:
			step.ctxName = r.(CtxItem).Key
			r = r.(CtxItem).Item
			break
		case *criteria.Reactive:
			r = r.(*criteria.Reactive).Get()
			break
		}

		r, _ = optional.TryExtractValIfOptional(r)

		switch r.(type) {
		case *Pipe:
			pipe := r.(*Pipe)
			pipe.
				ErrorHandler(this.errorHandler).
				ExitHandler(this.exitHandler).
				SetDebug(this.debug).
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
					SetDebug(this.debug).
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

			switch result.(type) {
			case *optional.Ok:
				continue
			}

			this.addStepResult(step, result)
			break
		case *optional.Fail:
			this.executeErrorHandler(r.(*optional.Fail))
			return this

		case *optional.None:
			if this.debug {
				logs.Info("step: %v, exited", i)
			}
			this.exitHandler()
			return this
		case *Continue:
			continue
		default:

			if iterable, ok := r.(foreach.Iterable); ok {
				iterable.SetErrorHandler(this.errorHandler)
				iterable.Execute()
			} else {

				this.addStepResult(step, r)
				//return this
			}

		}
	}

	this.executeSuccessHandler()

	return this
}

func (this *Pipe) executeErrorHandler(v *optional.Fail) {
	this.errorHandler.(func(fail *optional.Fail))(v)
}

func (this *Pipe) GetFailOrNil() *optional.Fail {
	return this.fail
}

func (this *Pipe) executeSuccessHandler() {
	if this.successHandler != nil {
		this.successHandler.(func())()
	}
}

func (this *Pipe) MapToBool() *optional.Optional[bool] {
	switch this.State {
	case StateCreated, StateExit:
		return optional.OfNone[bool]()
	case StateError:
		return optional.OfFail[bool](this.fail)
	case StateSuccess:
		return optional.Of[bool](true)
	default:
		return optional.OfNone[bool]()
	}
}

func (this *Pipe) RunAsResult() *result.Result[bool] {
	this.Run()
	switch this.State {
	case StateSuccess:
		return result.OfValue(true)
	case StateError:
		return result.OfError[bool](this.fail.Error)
	default:
		return result.OfErrorf[bool]("undefined state")
	}
}

type PipeT[T any] struct {
	Pipe
}

func NewT[T any]() *PipeT[T] {
	return &PipeT[T]{}
}

func (this PipeT[T]) OptResult() *optional.Optional[T] {
	switch this.State {
	case StateCreated, StateExit:
		return optional.OfNone[T]()
	case StateError:
		return optional.OfFail[T](this.fail)
	case StateSuccess:
		l := len(this.results)
		if l > 0 {
			last := this.results[l-1]
			if val, ok := last.(T); ok {
				return optional.Of[T](val)
			}
		}
	}

	return optional.OfNone[T]()
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

func GetCtxVal[T any](pipe *Pipe) T {
	var t T
	typOf := reflect.TypeOf(t)
	rval := pipe.findInCtxbyType(typOf, "GetCtxVal")
	return rval.Interface().(T)
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
