// server/server.go
package server

import "log/slog"

// Server holds dependencies for our application, like the logger.
type Server struct {
	Logger *slog.Logger
}
