package server

import (
	"log"
	"net/http"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
)

type Server struct {
	Addr string
	DB   *db.Store
	Mux  *http.ServeMux
	// Logger *log.Logger
}

func New(Addr string, db *db.Store) *Server {
	mux := http.NewServeMux()

	// Inject dependencies into handlers as needed
	// mux.Handle("/users/", http.StripPrefix("/users", handlers.UserRoutes(db)))
	// mux.Handle("/products/", http.StripPrefix("/products", handlers.ProductRoutes(db)))

	return &Server{
		Addr: Addr,
		DB:   db,
		Mux:  mux,
		// Logger: logger,
	}
}

// Start starts the HTTP server
func (s *Server) Start() error {
	log.Printf("Server starting on %s\n", s.Addr)
	return http.ListenAndServe(s.Addr, s.Mux)
}
