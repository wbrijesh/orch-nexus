package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"nexus/internal/database"
	"nexus/internal/server"
)

var dbService database.Service

// init runs before main, it initializes the database schema
func init() {
	log.Println("Initializing application...")

	// Initialize database connection and schema
	dbService = database.New()

	// Initialize schema - create tables if they don't exist
	if err := dbService.InitSchema(); err != nil {
		log.Printf("Warning: Failed to initialize database schema: %v", err)
		// We're not exiting here as the app might still work with existing schema
	} else {
		log.Println("Database schema initialized successfully")
	}
}

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	log.Println("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown with error: %v", err)
	}

	// Close database connection
	if err := dbService.Close(); err != nil {
		log.Printf("Error closing database connection: %v", err)
	}

	log.Println("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	// Pass the database service to the server
	server := server.NewServer(dbService)

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	log.Printf("Starting server on port %s", server.Addr)
	err := server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	log.Println("Graceful shutdown complete.")
}
