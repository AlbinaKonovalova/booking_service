package output

import "context"

// TxManager — исходящий порт для управления транзакциями.
// Application layer использует его, не зная о database/sql.
type TxManager interface {
	WithTx(ctx context.Context, fn func(ctx context.Context) error) error
}
