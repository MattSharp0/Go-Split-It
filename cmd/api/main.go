package main

import (
	"context"
	"log"
	"log/slog"
	"os"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(textHandler)
	slog.SetDefault(logger)

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Failed to load env settings")
	}
	dbAddress := os.Getenv("DATABASE_URL")

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbAddress) // Create a new connection pool
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}
	defer pool.Close()              // Closed the pool when main exists / an error occurs
	store, err := db.NewStore(pool) // Create a new store using the connection pool
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Test the connection to the database
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	} else {
		log.Println("Connected to the database successfully")
	}

	// Start server
	srv := server.New(":8080", &store)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
