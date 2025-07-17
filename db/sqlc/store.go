package db

import "github.com/jackc/pgx/v5/pgxpool"

// Interface that requires all methods from SQLC queries to be implemented
type Store interface {
	Querier
	// CreateSplitsTx eventually

}

// Implementation of the Store interface
// Stores the connection pool and embeds SQL generated quries
type SQLStore struct {
	connPool *pgxpool.Pool // Pointer to connection pool
	*Queries               // Embedding Queries to use SQLC methods
}

// Creates a new SQLStore which is an implementation of the Store interface
func NewStore(connPool *pgxpool.Pool) (Store, error) {
	// TODO return error on failure
	return &SQLStore{
		connPool: connPool,
		Queries:  New(connPool),
	}, nil
}
