package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Interface that requires all methods from SQLC queries to be implemented
// Use interface to allow for dependancy injection, instead of calling SQLStore directly. Pass value, not pointer.
type Store interface {
	Querier
	CreateSplitsTx(ctx context.Context, arg CreateSplitsTxParams) (CreateSplitsTxResult, error)
	UpdateTransactionSplitsTx(ctx context.Context, arg UpdateTransactionSplitsTxParams) (UpdateTransactionSplitsTxResult, error)
	DeleteTransactionWithSplitsTx(ctx context.Context, transactionID int64) error
	CreateGroupMembersTx(ctx context.Context, arg CreateGroupMemberTxParams) (CreateGroupMemberTxResult, error)
	UpdateGroupMembersTx(ctx context.Context, arg UpdateGroupMemberTxParams) (UpdateGroupMemberTxResult, error)
	DeleteGroupMembersTx(ctx context.Context, groupID int64) error
}

// Implementation of the Store interface
// Stores the connection pool and embeds SQL generated quries
// Concrete implementation of the Store interface, connects to the database. Pass pointer.
type SQLStore struct {
	connPool *pgxpool.Pool // Pointer to connection pool
	*Queries               // Embedding Queries to use SQLC methods
}

// Creates a new SQLStore
func NewStore(connPool *pgxpool.Pool) (*SQLStore, error) {
	// TODO return error on failure
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}, nil
}

func (store *SQLStore) execTx(ctx context.Context, fn func(*Queries) error) error {
	tx, err := store.connPool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	q := New(tx)
	err = fn(q)
	if err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			log.Printf("Transaction error: %v, rollback error: %v", err, rbErr)
			return fmt.Errorf("tx err: %v, rb err: %v", err, rbErr)
		}
		return err
	}

	return tx.Commit(ctx)
}
