package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tkahng/sticks/server"
	// Replace with your actual module path
)

// Main function with graceful shutdown
func main() {
	// Configuration
	const maxConcurrentGames = 1000
	const serverPort = ":8080"

	// Create and start game server
	srv := server.NewGameServer(maxConcurrentGames)
	srv.Start()

	// Create HTTP server
	// nolint:exhaustruct
	httpServer := &http.Server{
		Addr:    serverPort,
		Handler: server.Cors(srv.Hanlder()),
	}

	// Start HTTP server in goroutine
	go func() {
		log.Printf("Server starting on %s", serverPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown HTTP server
	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP server shutdown error: %v", err)
	}

	// Shutdown game server
	srv.Stop()

	log.Println("Server stopped")
}
