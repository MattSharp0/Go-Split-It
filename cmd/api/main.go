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
	textHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	logger := slog.New(textHandler)
	slog.SetDefault(logger)

	err := godotenv.Load(".dev.env")
	if err != nil {
		log.Fatal("Failed to load env settings")
	}
	dbAddress := os.Getenv("DATABASE_URL")

	ctx := context.Background()

	// Create a new database connection pool
	pool, err := pgxpool.New(ctx, dbAddress)
	if err != nil {
		log.Fatalf("DB connection failed: %v", err)
	}

	// Close the pool when main exits / an error occurs
	defer pool.Close()

	// Test the connection to the database
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	} else {
		log.Println("Connected to the database successfully")
	}

	// Create a new store using the connection pool
	store, err := db.NewStore(pool)
	if err != nil {
		log.Fatalf("Failed to create store: %v", err)
	}

	// Start server
	s := server.NewServer(":8080", &store)
	if err := s.Start(); err != nil {
		log.Fatal(err)

	}
}
