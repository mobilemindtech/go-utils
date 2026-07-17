package executor

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
)

// ErrTimeout é retornado por WaitResult quando o prazo de espera estoura
// antes do job terminar. Não é um erro de negócio: o job continua rodando
// em background e pode ser consultado depois via GetStatus.
var ErrTimeout = errors.New("agent: tempo de espera excedido, job continua em processamento")

// Processor é o único ponto que você precisa adaptar: a chamada real ao
// sistema de pagamento externo. Deve ser idempotente por si só quando possível,
// mas a garantia de "uma única execução por pedido" já é dada pelo Agent.
type Processor interface {
	Process(ctx context.Context, orderID string, payload json.RawMessage) (json.RawMessage, error)
}

type ProcessorFunc func(ctx context.Context, orderID string, payload json.RawMessage) (json.RawMessage, error)

func (f ProcessorFunc) Process(ctx context.Context, orderID string, payload json.RawMessage) (json.RawMessage, error) {
	return f(ctx, orderID, payload)
}

type Config struct {
	WorkerID     string        // prefixo do id deste processo/instância (ex: hostname)
	NumWorkers   int           // goroutines consumindo a fila
	PollInterval time.Duration // frequência de polling no banco
	ClaimBatch   int           // quantos jobs cada worker pega por vez
	StaleAfter   time.Duration // após quanto tempo um job "processing" é considerado travado
	MaxAttempts  int
}

func (c *Config) setDefaults() {
	if c.WorkerID == "" {
		c.WorkerID = "worker"
	}
	if c.NumWorkers <= 0 {
		c.NumWorkers = 4
	}
	if c.PollInterval <= 0 {
		c.PollInterval = 300 * time.Millisecond
	}
	if c.ClaimBatch <= 0 {
		c.ClaimBatch = 10
	}
	if c.StaleAfter <= 0 {
		c.StaleAfter = 2 * time.Minute
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = 3
	}
}

// Agent é o ponto de entrada único para submeter pagamentos e esperar o resultado.
type Agent struct {
	repo      *Repository
	processor Processor
	cfg       Config

	mu      sync.Mutex
	waiters map[string][]chan struct{} // notificação em memória, best-effort
}

func NewAgent(db_repo *Repository, processor Processor, cfg Config) *Agent {
	cfg.setDefaults()
	return &Agent{
		repo:      db_repo,
		processor: processor,
		cfg:       cfg,
		waiters:   make(map[string][]chan struct{}),
	}
}

// Start inicia os workers e a rotina de recuperação de jobs travados.
// Deve ser chamado uma vez, na inicialização da aplicação.
func (a *Agent) Start(ctx context.Context) {
	for i := 0; i < a.cfg.NumWorkers; i++ {
		go a.workerLoop(ctx, fmt.Sprintf("%s-%d", a.cfg.WorkerID, i))
	}
	go a.reaperLoop(ctx)
}

// Submit enfileira o pagamento de forma idempotente por orderID (UUID do pedido).
// Chamar Submit várias vezes para o mesmo orderID nunca cria processamento duplicado:
// a segunda chamada em diante apenas retorna o job já existente.
func (a *Agent) Submit(ctx context.Context, orderID string, payload json.RawMessage) (*Job, error) {
	job, _, err := a.repo.CreateIfNotExists(ctx, orderID, payload)
	if err != nil {
		return nil, fmt.Errorf("submit job %s: %w", orderID, err)
	}
	return job, nil
}

// WaitResult bloqueia até o job terminar ou o timeout estourar. Pode ser chamado
// a qualquer momento, inclusive por uma outra requisição/processo depois do
// Submit original — é assim que o "resumir o job" funciona.
func (a *Agent) WaitResult(ctx context.Context, orderID string, timeout time.Duration) (*Job, error) {
	job, err := a.repo.GetByOrderID(ctx, orderID)
	if err != nil {
		return nil, fmt.Errorf("job %s não encontrado: %w", orderID, err)
	}
	if job.IsTerminal() {
		return job, nil
	}

	notifyCh := a.registerWaiter(orderID)
	defer a.unregisterWaiter(orderID, notifyCh)

	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// fallback de polling: cobre o caso em que outro processo/instância
	// processou o job e a notificação em memória local nunca dispara
	pollTicker := time.NewTicker(250 * time.Millisecond)
	defer pollTicker.Stop()

	for {
		select {
		case <-waitCtx.Done():
			latest, _ := a.repo.GetByOrderID(context.Background(), orderID)
			if latest != nil && latest.IsTerminal() {
				return latest, nil
			}
			return latest, ErrTimeout

		case <-notifyCh:
			return a.repo.GetByOrderID(ctx, orderID)

		case <-pollTicker.C:
			latest, err := a.repo.GetByOrderID(ctx, orderID)
			if err == nil && latest.IsTerminal() {
				return latest, nil
			}
		}
	}
}

func (a *Agent) Delete(ctx context.Context, orderID string) (int64, error){
	return a.repo.DeleteByOrderID(ctx, orderID)
}

func (a *Agent) ListsJobs(ctx context.Context, search string, limit int, offset int) ([]*Job, int, error) {
	return a.repo.ListsJobs(ctx, search, limit, offset)
}

func (a *Agent) UpdatePayloads(ctx context.Context, orderID string, requestPayload json.RawMessage, responsePayload json.RawMessage) (int64, error){
	return a.repo.UpdatePayloadsByOrderID(ctx, orderID, requestPayload, responsePayload)
}

// GetStatus apenas consulta o estado atual, sem esperar — uso típico em um
// endpoint de consulta chamado depois que o cliente recebeu um timeout.
func (a *Agent) GetStatus(ctx context.Context, orderID string) (*Job, error) {
	return a.repo.GetByOrderID(ctx, orderID)
}

func (a *Agent) workerLoop(ctx context.Context, workerID string) {
	ticker := time.NewTicker(a.cfg.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			jobs, err := a.repo.ClaimBatch(ctx, workerID, a.cfg.ClaimBatch)
			if err != nil {
				log.Printf("agent: erro ao reivindicar jobs: %v", err)
				continue
			}
			for _, job := range jobs {
				a.processJob(ctx, job)
			}
		}
	}
}

func (a *Agent) processJob(ctx context.Context, job *Job) {
	result, procErr := a.safeProcess(ctx, job)

	if procErr != nil {
		retryable := job.Attempts < job.MaxAttempts
		if err := a.repo.Fail(ctx, job.OrderID, procErr.Error(), retryable); err != nil {
			log.Printf("agent: erro ao persistir falha do job %s: %v", job.OrderID, err)
		}
	} else {
		if err := a.repo.Complete(ctx, job.OrderID, result); err != nil {
			log.Printf("agent: erro ao persistir sucesso do job %s: %v", job.OrderID, err)
		}
	}

	a.notifyWaiters(job.OrderID)
}

// safeProcess isola qualquer panic do código de processamento (ex: chamada
// externa instável) para que nunca derrube o worker nem deixe o job travado.
func (a *Agent) safeProcess(ctx context.Context, job *Job) (result json.RawMessage, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic durante o processamento: %v", r)
		}
	}()
	return a.processor.Process(ctx, job.OrderID, job.RequestPayload)
}

func (a *Agent) reaperLoop(ctx context.Context) {
	ticker := time.NewTicker(a.cfg.StaleAfter / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			n, err := a.repo.RequeueStuck(ctx, a.cfg.StaleAfter)
			if err != nil {
				log.Printf("agent: erro no reaper: %v", err)
				continue
			}
			if n > 0 {
				log.Printf("agent: %d job(s) travado(s) recolocado(s) na fila", n)
			}
		}
	}
}

func (a *Agent) registerWaiter(orderID string) chan struct{} {
	ch := make(chan struct{}, 1)
	a.mu.Lock()
	a.waiters[orderID] = append(a.waiters[orderID], ch)
	a.mu.Unlock()
	return ch
}

func (a *Agent) unregisterWaiter(orderID string, target chan struct{}) {
	a.mu.Lock()
	defer a.mu.Unlock()
	chans := a.waiters[orderID]
	for i, ch := range chans {
		if ch == target {
			a.waiters[orderID] = append(chans[:i], chans[i+1:]...)
			break
		}
	}
	if len(a.waiters[orderID]) == 0 {
		delete(a.waiters, orderID)
	}
}

func (a *Agent) notifyWaiters(orderID string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, ch := range a.waiters[orderID] {
		select {
		case ch <- struct{}{}:
		default:
		}
	}
}
