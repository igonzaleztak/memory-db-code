package transport

import (
	"fmt"
	"log/slog"
	"memorydb/internal/db"
	"net/http"
	"strconv"
)

type Server struct {
	logger    *slog.Logger
	srv       *http.Server
	healthSrv *http.Server
}

// NewServer creates a new HTTP server with the provided logger, port, health port, and in-memory database.
//
// The server and the health server are listening to requests on different ports to allow for health checks
// without affecting the main application functionality.
func NewServer(logger *slog.Logger, port, healthPort int, db db.DBClient) *Server {
	srv := &http.Server{
		Addr:    ":" + strconv.Itoa(port),
		Handler: mountRouter(logger, db),
	}

	healthSrv := &http.Server{
		Addr:    ":" + strconv.Itoa(healthPort),
		Handler: mountHealthRouter(logger),
	}

	return &Server{
		logger:    logger,
		srv:       srv,
		healthSrv: healthSrv,
	}
}

// Start starts the HTTP server and listens for incoming requests on the specified port.
func (s *Server) Start() error {
	s.logger.Info("Starting HTTP server", "port", 8080)
	return s.srv.ListenAndServe()
}

// StartHealth starts the health HTTP server and listens for health check requests on the specified health port.
func (s *Server) StartHealth() error {
	s.logger.Info("Starting health HTTP server", "port", 8081)
	return s.healthSrv.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server and the health HTTP server.
func (s *Server) Shutdown() error {
	var errs []error

	s.logger.Info("Closing HTTP server")
	if err := s.srv.Close(); err != nil {
		s.logger.Error("Error closing main server", "error", err)
		errs = append(errs, err)
	}
	s.logger.Info("Closing health HTTP server")
	if err := s.healthSrv.Close(); err != nil {
		s.logger.Error("Error closing health server", "error", err)
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return fmt.Errorf("shutdown had errors: %v", errs)
	}
	s.logger.Info("HTTP servers closed successfully")
	return nil
}
