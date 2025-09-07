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

// enableCORS is a middleware that adds CORS headers to the response.
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		origin := r.Header.Get("Origin")
		if origin == "https://dreswap-ui.vercel.app" || origin == "http://localhost:3000" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Session-ID")
		w.Header().Set("Access-Control-Expose-Headers", "X-Session-ID")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Call the next handler
		next.ServeHTTP(w, r)
	})
}

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
		Handler:      enableCORS(mux),
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
