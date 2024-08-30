package postgresql

import (
	"context"
	"fmt"
	"users/internal/domain"
	"users/pkg/postgresql"
)

type transactionSupplier struct {
	db postgresql.Interface
}

// NewTransactionSupplier creates a new instance of transactionSupplier that satisfies the domain.Transaction interface
func NewTransactionSupplier(db postgresql.Interface) domain.Transaction {
	return &transactionSupplier{
		db: db,
	}
}

// BeginTx begins a transaction, injects it in the context and executes the fn function
// if fn fails, the transaction is rolledback
// if fn succeeds, the transaction is commited
func (u *transactionSupplier) BeginTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := u.db.GetPool().Begin(ctx)
	if err != nil {
		return err
	}
	ctx = context.WithValue(ctx, domain.TxKey, tx)

	if err = fn(ctx); err != nil {
		if err2 := tx.Rollback(ctx); err2 != nil {
			return fmt.Errorf("%w: %w", err2, err)
		}
		return err
	}
	return tx.Commit(ctx)
}
