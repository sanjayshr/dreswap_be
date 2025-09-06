// main.go
package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/sanjayshr/event-outfitter-backend/handler"
	"github.com/sanjayshr/event-outfitter-backend/server"
)

func main() {
	// Initialize structured logger
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	s := server.NewServer(logger)

	// Use the new ServeMux for pattern-based routing
	mux := http.NewServeMux()

	// Register handlers
	mux.HandleFunc("POST /api/v1/generate", handler.GenerateHandler(s))
	mux.HandleFunc("POST /api/v1/swap-style", handler.SwapStyleHandler(s)) // New endpoint
	mux.HandleFunc("GET /api/v1/styles", handler.GetStylesHandler(s))      // New endpoint

	// A simple health check endpoint
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Configure the HTTP server
	srv := &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("Starting server", "address", srv.Addr)
	err := srv.ListenAndServe()
	if err != nil {
		logger.Error("Server failed to start", "error", err)
		os.Exit(1)
	}
}

