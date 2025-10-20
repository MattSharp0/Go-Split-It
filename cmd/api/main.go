package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/handlers"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/server"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables first
	err := godotenv.Load(".dev.env")
	if err != nil {
		slog.Error("Failed to load env settings", "error", err)
		os.Exit(1)
	}

	// Initialize logger from environment config
	logCfg := logger.LoadConfigFromEnv()
	log, err := logger.InitLogger(logCfg)
	if err != nil {
		slog.Error("Failed to initialize logger", "error", err)
		os.Exit(1)
	}

	log.Info("Application starting",
		slog.String("environment", os.Getenv("ENVIROMENT")),
		slog.String("version", os.Getenv("VERSION")),
		slog.String("log_level", string(logCfg.Level)),
		slog.String("log_output", string(logCfg.Output)),
	)

	dbAddress := os.Getenv("DATABASE_URL")

	ctx := context.Background()

	// Create a new database connection pool
	pool, err := pgxpool.New(ctx, dbAddress)
	if err != nil {
		log.Error("DB connection failed", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Test the connection to the database
	if err := pool.Ping(ctx); err != nil {
		log.Error("Failed to ping database", "error", err)
		os.Exit(1)
	}
	log.Info("Connected to the database successfully")

	// Create a new store using the connection pool
	store, err := db.NewStore(pool)
	if err != nil {
		log.Error("Failed to create store", "error", err)
		os.Exit(1)
	}
	log.Debug("Store created successfully")

	// Initialize server with logger
	s := server.NewServer(":8080", store, log)

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
		log.Error("Failed to start server", "error", err)
		os.Exit(1)
	}

	log.Info("Server started successfully")

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("Shutdown signal received")

	if err := s.Stop(); err != nil {
		log.Error("Error during shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("Application shutdown gracefully")
}
