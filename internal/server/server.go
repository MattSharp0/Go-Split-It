package server

import (
	"context"
	"log"
	"net/http"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
)

type Server struct {
	Server *http.Server
	DB     *db.Store
	// Logger *log.Logger // TODO add specific logger to server
}

// Create new server at address with acompanying db pointer.
func NewServer(Addr string, db *db.Store) *Server {

	s := &Server{
		Server: &http.Server{
			Addr:         Addr,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 10 * time.Second,
			IdleTimeout:  120 * time.Second,
		},
		DB: db,
	}
	mux := http.NewServeMux()
	s.Server.Handler = mux

	return s
}

// Start the HTTP server if server is not shutdown, return error if fails
func (s *Server) Start() error {
	log.Printf("Server starting...")
	err := s.Server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		log.Printf("Server failed to start: %v", err)
		return err
	}
	log.Printf("Server listening on %s\n", s.Server.Addr)
	return nil
}

// Shutdown server *gracefully* or return error
func (s *Server) Stop() error {
	log.Printf("Server shutting down...")

	// TODO: Update Context
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.Server.Shutdown(ctx)
	if err != nil {
		log.Printf("Server failed to shutdown: %v", err)
		return err
	}
	log.Printf("Server shutdown complete")
	return nil
}

// Exposes server mux.
// Register routes via s.Mux().Handlefunc(<route>, <handler>)
func (s *Server) Mux() *http.ServeMux {
	return s.Server.Handler.(*http.ServeMux)
}
