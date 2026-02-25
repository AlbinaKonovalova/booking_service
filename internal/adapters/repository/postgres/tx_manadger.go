package postgres

import (
	"context"
	"database/sql"
	"fmt"
)

type ctxKeyTx struct{}

// TxManager — реализация управления транзакциями для PostgreSQL.
type TxManager struct {
	db *sql.DB
}

// NewTxManager создаёт новый TxManager.
func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTx выполняет функцию внутри транзакции.
func (m *TxManager) WithTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}

	txCtx := context.WithValue(ctx, ctxKeyTx{}, tx)

	if err := fn(txCtx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback failed: %w (original: %v)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tx: %w", err)
	}

	return nil
}

// queryExecutor — интерфейс, который реализуют и *sql.DB, и *sql.Tx.
type queryExecutor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// getExecutor возвращает *sql.Tx из контекста, или *sql.DB если транзакции нет.
func getExecutor(ctx context.Context, db *sql.DB) queryExecutor {
	if tx, ok := ctx.Value(ctxKeyTx{}).(*sql.Tx); ok {
		return tx
	}
	return db
}
