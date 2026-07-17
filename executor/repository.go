package executor

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
)

type Repository struct {
	db          *sql.DB
	maxAttempts int
}

func NewRepository(db *sql.DB, maxAttempts int) *Repository {
	if maxAttempts <= 0 {
		maxAttempts = 3
	}
	return &Repository{db: db, maxAttempts: maxAttempts}
}

// CreateIfNotExists tenta criar o job. Se já existir (mesmo order_id), retorna o
// job já existente sem erro — é aqui que a duplicidade de transação é bloqueada,
// na origem, via UNIQUE KEY em order_id.
func (r *Repository) CreateIfNotExists(ctx context.Context, orderID string, payload json.RawMessage) (job *Job, created bool, err error) {
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO payment_jobs (order_id, status, request_payload, max_attempts)
		VALUES (?, ?, ?, ?)
	`, orderID, StatusPending, payload, r.maxAttempts)

	if err != nil {
		if mysqlErr, ok := errors.AsType[*mysql.MySQLError](err); ok && mysqlErr.Number == 1062 { // duplicate key
			existing, getErr := r.GetByOrderID(ctx, orderID)
			return existing, false, getErr
		}
		return nil, false, fmt.Errorf("insert job: %w", err)
	}

	job, err = r.GetByOrderID(ctx, orderID)
	return job, true, err
}

func (r *Repository) GetByOrderID(ctx context.Context, orderID string) (*Job, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT order_id, status, request_payload, response_payload, error_message,
		       attempts, max_attempts, locked_by, locked_at, created_at, updated_at, completed_at
		FROM payment_jobs WHERE order_id = ? 
	`, orderID)
	return scanJob(row)
}


func (r *Repository) ListsJobs(ctx context.Context, search string, limit int, offset int) ([]*Job, int, error) {

	where := ""
	var args []interface{}

	if len(search) > 0 {
		where = "WHERE upper(request_payload) LIKE upper(?)"
		args = append(args, fmt.Sprintf("%%%s%%", search))
	}

	args = append(args, limit, offset)

	sql := fmt.Sprintf(`
		SELECT order_id, status, request_payload, response_payload, error_message,
		       attempts, max_attempts, locked_by, locked_at, created_at, updated_at, completed_at
		FROM payment_jobs
		%s
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, where)

	rows, err := r.db.QueryContext(ctx, sql, args...)

	if err != nil {
		return nil, 0, err
	}

	var jobs []*Job
	var count int

	for rows.Next() {
		if rows.Err() != nil {
			return nil, 0, rows.Err()
		}
		job, err := scanJob(rows)

		if err != nil {
			return nil, 0, rows.Err()
		}

		jobs = append(jobs, job)
	}

	if len(search) > 0 {
		args = []interface{} { fmt.Sprintf("%%%v%%", search) }
	} else {
		args = []interface{} {}
	}

	row := r.db.QueryRowContext(ctx, fmt.Sprintf(`SELECT count(*) FROM payment_jobs %v`, where), args...)


	if row.Err() != nil {
		return nil, 0, row.Err()
	}

	if err := row.Scan(&count); err != nil {
		return nil, 0, err
	}

	return jobs, count, nil
}

func (r *Repository) DeleteByOrderID(ctx context.Context, orderID string) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM payment_jobs WHERE order_id = ?
	`, orderID)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

func (r *Repository) UpdatePayloadsByOrderID(ctx context.Context, orderID string, requestPayload json.RawMessage, responsePayload json.RawMessage) (int64, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE payment_jobs set request_payload = ?, response_payload = ? WHERE  order_id = ?
	`, requestPayload, responsePayload, orderID)

	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// ClaimBatch reivindica até `limit` jobs pendentes de forma atômica, segura
// para múltiplos workers e múltiplas instâncias da API rodando em paralelo.
// SKIP LOCKED garante que dois workers nunca peguem o mesmo job.
func (r *Repository) ClaimBatch(ctx context.Context, workerID string, limit int) ([]*Job, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	rows, err := tx.QueryContext(ctx, `
		SELECT order_id FROM payment_jobs
		WHERE status = ?
		ORDER BY created_at ASC
		LIMIT ?
		FOR UPDATE SKIP LOCKED
	`, StatusPending, limit)
	if err != nil {
		return nil, fmt.Errorf("select pending: %w", err)
	}

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	rows.Close()

	if len(ids) == 0 {
		return nil, tx.Commit()
	}

	placeholders := strings.TrimRight(strings.Repeat("?,", len(ids)), ",")

	updateArgs := make([]interface{}, 0, len(ids)+2)
	updateArgs = append(updateArgs, StatusProcessing, workerID)
	for _, id := range ids {
		updateArgs = append(updateArgs, id)
	}
	_, err = tx.ExecContext(ctx, fmt.Sprintf(`
		UPDATE payment_jobs
		SET status = ?, locked_by = ?, locked_at = NOW(), attempts = attempts + 1
		WHERE order_id IN (%s)
	`, placeholders), updateArgs...)
	if err != nil {
		return nil, fmt.Errorf("claim update: %w", err)
	}

	selectArgs := make([]interface{}, len(ids))
	for i, id := range ids {
		selectArgs[i] = id
	}
	claimedRows, err := tx.QueryContext(ctx, fmt.Sprintf(`
		SELECT order_id, status, request_payload, response_payload, error_message,
		       attempts, max_attempts, locked_by, locked_at, created_at, updated_at, completed_at
		FROM payment_jobs WHERE order_id IN (%s)
	`, placeholders), selectArgs...)
	if err != nil {
		return nil, fmt.Errorf("select claimed: %w", err)
	}
	defer claimedRows.Close()

	var jobs []*Job
	for claimedRows.Next() {
		j, err := scanJob(claimedRows)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit claim: %w", err)
	}
	return jobs, nil
}

// Complete marca o job como concluído com sucesso e grava a resposta.
func (r *Repository) Complete(ctx context.Context, orderID string, response json.RawMessage) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE payment_jobs
		SET status = ?, response_payload = ?, error_message = NULL, completed_at = NOW()
		WHERE order_id = ?
	`, StatusCompleted, response, orderID)
	return err
}

// Fail registra o erro. Se ainda há tentativas disponíveis, volta o job para
// pending (retry automático no próximo ciclo de polling); senão, marca failed.
func (r *Repository) Fail(ctx context.Context, orderID string, errMsg string, retryable bool) error {
	nextStatus := StatusFailed
	var completedAtClause string
	if retryable {
		nextStatus = StatusPending
		completedAtClause = ""
	} else {
		completedAtClause = ", completed_at = NOW()"
	}
	query := fmt.Sprintf(`
		UPDATE payment_jobs
		SET status = ?, error_message = ?, locked_by = NULL, locked_at = NULL%s
		WHERE order_id = ?
	`, completedAtClause)
	_, err := r.db.ExecContext(ctx, query, nextStatus, errMsg, orderID)
	return err
}

// RequeueStuck devolve para 'pending' jobs presos em 'processing' por mais
// tempo que staleAfter — cobre crash de worker no meio do processamento.
func (r *Repository) RequeueStuck(ctx context.Context, staleAfter time.Duration) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		UPDATE payment_jobs
		SET status = ?, locked_by = NULL, locked_at = NULL
		WHERE status = ? AND locked_at < ?
	`, StatusPending, StatusProcessing, time.Now().Add(-staleAfter))
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

type rowScanner interface {
	Scan(dest ...interface{}) error
}

func scanJob(row rowScanner) (*Job, error) {
	var j Job
	var responsePayload sql.NullString
	var errorMessage sql.NullString
	var lockedBy sql.NullString
	var lockedAt sql.NullTime
	var completedAt sql.NullTime
	var requestPayload []byte

	err := row.Scan(
		&j.OrderID, &j.Status, &requestPayload, &responsePayload, &errorMessage,
		&j.Attempts, &j.MaxAttempts, &lockedBy, &lockedAt, &j.CreatedAt, &j.UpdatedAt, &completedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
		return nil, fmt.Errorf("scan job: %w", err)
	}

	j.RequestPayload = requestPayload
	if responsePayload.Valid {
		j.ResponsePayload = json.RawMessage(responsePayload.String)
	}
	if errorMessage.Valid {
		j.ErrorMessage = new(errorMessage.String)
	}
	if lockedBy.Valid {
		j.LockedBy = new(lockedBy.String)
	}
	if lockedAt.Valid {
		j.LockedAt = &lockedAt.Time
	}
	if completedAt.Valid {
		j.CompletedAt = &completedAt.Time
	}
	return &j, nil
}
