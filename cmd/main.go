package main

import (
	"context"
	"log"
	"memorydb/internal/config"
	"memorydb/internal/db"
	"memorydb/internal/logger"
	"memorydb/internal/transport"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "go.uber.org/automaxprocs"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	configuration, err := config.LoadConfig()
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	logger := logger.NewLogger(configuration.Verbose)

	// start the in-memory database
	logger.Info("Starting MemoryDB application", "version", "1.0.0")
	db := db.NewMemoryDB(
		db.WithCleanupInterval(configuration.DefaultCleanupInterval),
	)

	// Start the HTTP server with the loaded configuration and database instance
	httpServer := transport.NewServer(
		logger,
		*configuration.Port,
		*configuration.HealthPort,
		db,
	)

	go func() {
		if err := httpServer.StartHealth(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start health HTTP server", "error", err)
			cancel() // Cancel the context to trigger shutdown
		}
	}()

	go func() {
		if err := httpServer.Start(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to start HTTP server", "error", err)
			cancel() // Cancel the context to trigger shutdown
		}
	}()

	// Wait for shutdown signal and gracefully shut down the db and server
	<-ctx.Done()
	logger.Info("Received shutdown signal, shutting down...")
	db.Close() // Close the in-memory database
	if err := httpServer.Shutdown(); err != nil {
		logger.Error("Error shutting down HTTP server", "error", err)
		os.Exit(1) // Exit if server fails to shut down gracefully
	}
	logger.Info("Shut down complete, exiting application")
}
