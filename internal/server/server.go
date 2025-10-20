package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
)

type Server struct {
	Server *http.Server
	mux    *http.ServeMux
	DB     *db.Store
	Logger *slog.Logger
}

// Create new server at address with accompanying db pointer and logger.
func NewServer(Addr string, db db.Store, log *slog.Logger) *Server {
	mux := http.NewServeMux()

	s := &Server{
		Server: &http.Server{
			Addr:         Addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
			Handler:      logger.HTTPMiddleware(mux), // Wrap with logging middleware
		},
		mux:    mux,
		DB:     &db,
		Logger: log,
	}

	return s
}

// Start the HTTP server if server is not shutdown, return error if fails
func (s *Server) Start() error {
	s.Logger.Debug("Server starting...", slog.String("address", s.Server.Addr))

	go func() {
		err := s.Server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			s.Logger.Error("Server failed to start", "error", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	s.Logger.Info("Server listening", slog.String("address", s.Server.Addr))

	return nil
}

// Shutdown server *gracefully* or return error
func (s *Server) Stop() error {
	s.Logger.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Server.Shutdown(ctx)
	if err != nil {
		s.Logger.Error("Server failed to shutdown gracefully", "error", err)
		return err
	}
	s.Logger.Info("Server shutdown complete")
	return nil
}

// Exposes server mux.
// Register routes via s.Mux().Handlefunc(<route>, <handler>)
func (s *Server) Mux() *http.ServeMux {
	return s.mux
}
