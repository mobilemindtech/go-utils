package executor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// --- 1) O único ponto que você adapta: a chamada real ao sistema de pagamentos ---

type gatewayProcessor struct {
	// injete aqui o client HTTP/SDK do sistema externo de pagamento
}

func (g *gatewayProcessor) Process(ctx context.Context, orderID string, payload json.RawMessage) (json.RawMessage, error) {
	// TODO: chamar o sistema externo de pagamento aqui.
	// Se der erro, apenas retorne err — o Agent cuida de persistir e permitir retry.
	// Se der certo, retorne o payload de resposta que quer guardar.
	return json.RawMessage(`{"status":"approved"}`), nil
}

// --- 2) Wiring na inicialização da aplicação ---

var agent *Agent

func main() {
	db, err := sql.Open("mysql", "test:test@tcp(127.0.0.1:3306)/myapp?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	db.SetMaxOpenConns(20)

	repo := NewRepository(db, 3) // 3 tentativas por job
	agent = NewAgent(repo, &gatewayProcessor{}, Config{
		WorkerID:     "api-01", // use hostname/pod name em produção
		NumWorkers:   4,
		PollInterval: 300 * time.Millisecond,
		ClaimBatch:   10,
		StaleAfter:   2 * time.Minute,
		MaxAttempts:  3,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	agent.Start(ctx) // inicia os workers em background

	http.HandleFunc("/pedidos/pagamento", handleCreatePayment)
	http.HandleFunc("/pedidos/pagamento/status", handlePaymentStatus)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// --- 3) Handler que mantém a semântica atual (síncrono do ponto de vista do cliente) ---

type createPaymentRequest struct {
	OrderID string          `json:"order_id"` // UUID do pedido, gerado pelo cliente/checkout
	Payload json.RawMessage `json:"payload"`  // dados do pagamento
}

func handleCreatePayment(w http.ResponseWriter, r *http.Request) {
	var req createPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload inválido", http.StatusBadRequest)
		return
	}
	ctx := r.Context()

	// Idempotente: se esse order_id já foi submetido antes (ex: retry do
	// cliente, double-click, timeout de rede), não processa de novo.
	if _, err := agent.Submit(ctx, req.OrderID, req.Payload); err != nil {
		http.Error(w, "erro ao registrar pagamento", http.StatusInternalServerError)
		return
	}

	// Espera até 8s pelo resultado, preservando o comportamento síncrono atual.
	job, err := agent.WaitResult(ctx, req.OrderID, 8*time.Second)

	switch {
	case errors.Is(err, ErrTimeout):
		// Não terminou a tempo, mas está seguro rodando em background.
		// Cliente consulta depois pelo mesmo order_id.
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"order_id": req.OrderID,
			"status":   "processing",
		})
	case err != nil:
		http.Error(w, "erro ao processar pagamento", http.StatusInternalServerError)
	case job.Status == StatusCompleted:
		w.WriteHeader(http.StatusOK)
		w.Write(job.ResponsePayload)
	default: // failed
		w.WriteHeader(http.StatusUnprocessableEntity)
		json.NewEncoder(w).Encode(map[string]string{
			"order_id": req.OrderID,
			"status":   "failed",
			"error":    safeString(job.ErrorMessage),
		})
	}
}

// --- 4) Endpoint para "resumir" a consulta depois de um 202 Accepted ---

func handlePaymentStatus(w http.ResponseWriter, r *http.Request) {
	orderID := r.URL.Query().Get("order_id")
	job, err := agent.GetStatus(r.Context(), orderID)
	if err != nil {
		http.Error(w, "pedido não encontrado", http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"order_id": job.OrderID,
		"status":   job.Status,
		"response": job.ResponsePayload,
		"error":    safeString(job.ErrorMessage),
	})
}

func safeString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
