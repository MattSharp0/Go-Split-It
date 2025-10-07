package db

import "github.com/jackc/pgx/v5/pgxpool"

// Interface that requires all methods from SQLC queries to be implemented
// Use interface to allow for dependancy injection, instead of calling SQLStore directly. Pass value, not pointer.
type Store interface {
	Querier
	// CreateSplitsTx eventually

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
