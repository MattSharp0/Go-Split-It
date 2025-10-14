package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/handlers"
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
	} else {
		log.Println("Store created successfully")
	}

	// Initialize server
	s := server.NewServer(":8080", store)

	s.Mux().HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, World!"))
	})

	s.Mux().Handle("/users/", http.StripPrefix("/users", handlers.UserRoutes(s, store)))
	s.Mux().Handle("/groups/", http.StripPrefix("/groups", handlers.GroupRoutes(s, store)))
	s.Mux().Handle("/group_members/", http.StripPrefix("/group_members", handlers.GroupMemberRoutes(s, store)))
	s.Mux().Handle("/transactions/", http.StripPrefix("/transactions", handlers.TransactionRoutes(s, store)))
	s.Mux().Handle("/splits/", http.StripPrefix("/splits", handlers.SplitRoutes(s, store)))

	// Start server in goroutine
	if err := s.Start(); err != nil {
		log.Fatal(err)
	}

	log.Println("Server started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := s.Stop(); err != nil {
		log.Fatal(err)
	}
}
