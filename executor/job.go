package executor

import (
	"encoding/json"
	"time"
)

// Status representa o estado do job de pagamento.
type Status string

const (
	StatusPending    Status = "pending"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
)

// Job é a representação persistida de um processamento de pagamento.
// order_id é a chave de idempotência: um pedido só pode ter um job.
type Job struct {
	OrderID         string
	Status          Status
	RequestPayload  json.RawMessage
	ResponsePayload json.RawMessage
	ErrorMessage    *string
	Attempts        int
	MaxAttempts     int
	LockedBy        *string
	LockedAt        *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
	CompletedAt     *time.Time
}

// IsTerminal indica se o job já chegou a um estado final (sucesso ou falha definitiva).
func (j *Job) IsTerminal() bool {
	return j.Status == StatusCompleted || j.Status == StatusFailed
}
