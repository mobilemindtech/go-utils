package task

import (
	"sync"

	"github.com/beego/beego/v2/core/logs"
	"github.com/mobilemindtec/go-utils/v2/ctx"
	"github.com/mobilemindtec/go-utils/v2/optional"
	"github.com/sirsean/go-pool"
)

type TaskType int

type TaskSource[T any] struct {
	Data T
}

func NewTaskSource[T any]() *TaskSource[T] {
	return &TaskSource[T]{}
}

type TaskResult[T any] struct {
	Result T
	Error  error
}

func NewTaskResult[T any]() *TaskResult[T] {
	return &TaskResult[T]{}
}

func (this *TaskResult[T]) HasError() bool {
	return this.Error != nil
}

type Runnable[DS any, R any] struct {
	TaskR     func(DS, *TaskGroup[DS, R]) R
	Task      func(DS, *TaskGroup[DS, R])
	TaskGroup *TaskGroup[DS, R]
	Data      DS
}

func (this *Runnable[DS, R]) Perform() {

	defer func() {
		if r := recover(); r != nil {
			logs.Error("error to perform task: %v", r)
		}
	}()

	if this.Task != nil {
		this.Task(this.Data, this.TaskGroup)
	} else if this.TaskR != nil {
		r := this.TaskR(this.Data, this.TaskGroup)
		this.TaskGroup.AddResult(r)
	} else {
		logs.Error("no task to perform")
	}

}

type TaskGroup[DS any, R any] struct {
	ResultChan chan R
	DataSource []DS
	results    []R
	Ctx        *ctx.Ctx
	Mutex      *sync.Mutex
	WaitGroup  *sync.WaitGroup
	PoolSize   int
	onReceive  func(R)
	taskR      func(DS, *TaskGroup[DS, R]) R
	task       func(DS, *TaskGroup[DS, R])
}

func New[DS any, R any]() *TaskGroup[DS, R] {
	var wg sync.WaitGroup
	return &TaskGroup[DS, R]{Ctx: ctx.New(),
		Mutex:     new(sync.Mutex),
		WaitGroup: &wg,
		PoolSize:  10,
		results:   []R{}}
}

func (this *TaskGroup[DS, R]) SetPoolSize(i int) *TaskGroup[DS, R] {
	this.PoolSize = i
	return this
}

func (this *TaskGroup[DS, R]) SetTask(r func(DS, *TaskGroup[DS, R])) *TaskGroup[DS, R] {
	this.task = r
	return this
}

func (this *TaskGroup[DS, R]) SetTaskR(r func(DS, *TaskGroup[DS, R]) R) *TaskGroup[DS, R] {
	this.taskR = r
	return this
}

func (this *TaskGroup[DS, R]) SetOnReceive(r func(R)) *TaskGroup[DS, R] {
	this.onReceive = r
	return this
}

func (this *TaskGroup[DS, R]) SetDataSource(ds []DS) *TaskGroup[DS, R] {
	this.DataSource = append(this.DataSource, ds...)
	return this
}

func (this *TaskGroup[DS, R]) AddItem(item DS) *TaskGroup[DS, R] {
	this.DataSource = append(this.DataSource, item)
	return this
}

func (this *TaskGroup[DS, R]) AddResult(r R) {
	this.Mutex.Lock()
	this.results = append(this.results, r)
	this.Mutex.Unlock()
	if this.onReceive != nil {
		go this.onReceive(r)
	}
}

func (this *TaskGroup[DS, R]) AddCtx(key string, val interface{}) *TaskGroup[DS, R] {
	this.Mutex.Lock()
	this.Ctx.Put(key, val)
	this.Mutex.Unlock()
	return this
}

func (this *TaskGroup[DS, R]) GetCtxOpt(key string) interface{} {
	if this.Ctx.Has(key) {
		return optional.NewSome(this.Ctx.Get(key))
	}
	return optional.NewNone()
}

func (this *TaskGroup[DS, R]) GetCtx(key string) (interface{}, bool) {
	if this.Ctx.Has(key) {
		return this.Ctx.Get(key), true
	}
	return nil, false
}

func (this *TaskGroup[DS, R]) HasCtx(key string) bool {
	return this.Ctx.Has(key)
}

func (this *TaskGroup[DS, R]) GetResults() []R {
	return this.results
}

func (this *TaskGroup[DS, R]) Start() *TaskGroup[DS, R] {

	p := pool.NewPool(10, 20)
	p.Start()

	for _, data := range this.DataSource {
		p.Add(&Runnable[DS, R]{
			TaskR:     this.taskR,
			Task:      this.task,
			TaskGroup: this,
			Data:      data})
	}

	p.Close()

	/*

			this.ResultChan = make(chan R) // armazena dos erros do processamento async

			for _, data := range this.DataSource {
				this.WaitGroup.Add(1)    // novo trabalho em espera
				go func() {
		      defer this.WaitGroup.Done()
		      this.task(data, this) // inicia trabalho
		    }
			}

			// espera pelo fim do processamento em background, para que a leitura do canal funciona na thread principal
			go func() {
				logs.Debug("wait by workers..")
				this.WaitGroup.Wait()
				close(this.ResultChan)
				logs.Debug("workers is done")
			}()*/

	return this
}
