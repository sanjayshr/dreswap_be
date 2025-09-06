// server/server.go
package server

import (
	"log/slog"
	"sync"

	"github.com/sanjayshr/event-outfitter-backend/models"
)

// SessionData holds all relevant data for a user's style generation session.
type SessionData struct {
	Styles      []string
	ImageData   []byte
	MimeType    string
	RequestData models.GenerateRequest // Original request data
}

// Server holds dependencies for our application, like the logger and session cache.
type Server struct {
	Logger *slog.Logger

	// sessionCache stores all session data for active sessions.
	// Key: sessionID (string), Value: SessionData
	SessionCache map[string]SessionData
	CacheMutex   sync.Mutex
}

// NewServer creates and initializes a new Server instance.
func NewServer(logger *slog.Logger) *Server {
	return &Server{
		Logger:       logger,
		SessionCache: make(map[string]SessionData),
	}
}

