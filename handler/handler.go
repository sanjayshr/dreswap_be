// handlers.go
package handler

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/sanjayshr/event-outfitter-backend/gemini"
	"github.com/sanjayshr/event-outfitter-backend/models"
	"github.com/sanjayshr/event-outfitter-backend/server"
)

// maxUploadSize defines the maximum allowed file upload size (10 MB).
const maxUploadSize = 10 * 1024 * 1024 // 10 MB

// GenerateHandler handles the /api/v1/generate endpoint.
func GenerateHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Enforce a maximum request body size
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			s.Logger.Error("Failed to parse multipart form", "error", err)
			http.Error(w, "The uploaded file is too big. Please choose an image that is less than 10MB in size.", http.StatusBadRequest)
			return
		}

		// 1. Parse the JSON data part
		jsonData := r.FormValue("data")
		var reqData models.GenerateRequest
		if err := json.Unmarshal([]byte(jsonData), &reqData); err != nil {
			s.Logger.Error("Failed to unmarshal JSON data", "error", err)
			http.Error(w, "Invalid JSON data provided.", http.StatusBadRequest)
			return
		}
		s.Logger.Info("Received generation request", "data", reqData)

		// 2. Parse the image file part
		file, handler, err := r.FormFile("image")
		if err != nil {
			s.Logger.Error("Failed to get image from form", "error", err)
			http.Error(w, "Invalid image file provided.", http.StatusBadRequest)
			return
		}
		defer file.Close()

		imgData, err := io.ReadAll(file)
		if err != nil {
			s.Logger.Error("Failed to read image data", "error", err)
			http.Error(w, "Could not read image data.", http.StatusInternalServerError)
			return
		}

		// --- Replace with this robust logic ---
		var mimeType string

		// First, try to get the MIME type from the file extension.
		// This is often the most reliable method.
		mimeType = mime.TypeByExtension(filepath.Ext(handler.Filename))

		// If the extension is unknown, fall back to content detection.
		if mimeType == "" {
			mimeType = http.DetectContentType(imgData)
		}

		// FINAL CHECK: If the type is still generic, make an educated guess based on the extension.
		// This handles cases where system mime types are not configured for .jpg, etc.
		if mimeType == "application/octet-stream" {
			ext := strings.ToLower(filepath.Ext(handler.Filename))
			switch ext {
			case ".jpg", ".jpeg":
				mimeType = "image/jpeg"
			case ".png":
				mimeType = "image/png"
			case ".webp":
				mimeType = "image/webp"
				// Add other supported image types as needed
			}
		}

		s.Logger.Info("Image received", "filename", handler.Filename, "size", handler.Size, "mimeType", mimeType)
		// --- End of replacement ---

		// 3. Get style suggestions from Gemini (text-only call)
		styles, err := gemini.GetStyleSuggestions(r.Context(), s.Logger, reqData.EventType, reqData.Venue, reqData.Theme)
		if err != nil {
			s.Logger.Error("Failed to get style suggestions", "error", err)
			http.Error(w, "Failed to get style suggestions.", http.StatusInternalServerError)
			return
		}
		if len(styles) == 0 {
			s.Logger.Error("No style suggestions returned")
			http.Error(w, "No style suggestions could be generated.", http.StatusInternalServerError)
			return
		}

		// 4. Generate a session ID and store image data and styles in cache
		sessionID := uuid.New().String()
		// Save session ID to a file for easy access
		f, err := os.OpenFile("session.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			s.Logger.Error("Failed to open session log file", "error", err)
			// Do not fail the request, just log the error
		} else {
			if _, err := f.WriteString(sessionID + "\n"); err != nil {
				s.Logger.Error("Failed to write session ID to log file", "error", err)
			}
			f.Close()
		}
		sessionData := server.SessionData{
			Styles:      styles,
			ImageData:   imgData,
			MimeType:    mimeType,
			RequestData: reqData,
		}

		s.CacheMutex.Lock()
		s.SessionCache[sessionID] = sessionData
		s.CacheMutex.Unlock()

		// 5. Generate the first image using the first style
		generatedImg, generatedMimeType, err := gemini.GenerateImage(r.Context(), s.Logger, sessionData.ImageData, sessionData.MimeType, sessionData.RequestData.EventType, sessionData.RequestData.Venue, sessionData.RequestData.Theme, sessionData.Styles[0])
		if err != nil {
			s.Logger.Error("Failed to generate initial image via Gemini", "error", err)
			http.Error(w, "Failed to generate initial image.", http.StatusInternalServerError)
			return
		}

		// 6. Write the successful response with the first image and session ID
		w.Header().Set("Content-Type", generatedMimeType)
		w.Header().Set("X-Session-ID", sessionID) // Return session ID in header
		w.WriteHeader(http.StatusOK)
		w.Write(generatedImg)
	}
}

// SwapStyleHandler handles the /api/v1/swap-style endpoint.
func SwapStyleHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			s.Logger.Error("Missing X-Session-ID header")
			http.Error(w, "Missing X-Session-ID header.", http.StatusBadRequest)
			return
		}

		var swapReq models.SwapStyleRequest
		if err := json.NewDecoder(r.Body).Decode(&swapReq); err != nil {
			s.Logger.Error("Failed to decode swap style request", "error", err)
			http.Error(w, "Invalid request body.", http.StatusBadRequest)
			return
		}

		s.CacheMutex.Lock()
		sessionData, found := s.SessionCache[sessionID]
		s.CacheMutex.Unlock()

		if !found {
			s.Logger.Error("Session data not found", "sessionID", sessionID)
			http.Error(w, "Session expired or invalid.", http.StatusNotFound)
			return
		}

		s.Logger.Info("Found session data", "sessionID", sessionID, "styles", sessionData.Styles, "stylesCount", len(sessionData.Styles), "mimeType", sessionData.MimeType, "requestData", sessionData.RequestData)

		if swapReq.StyleIndex < 0 || swapReq.StyleIndex >= len(sessionData.Styles) {
			s.Logger.Error("Invalid style index", "sessionID", sessionID, "styleIndex", swapReq.StyleIndex, "numStyles", len(sessionData.Styles))
			http.Error(w, "Invalid style index.", http.StatusBadRequest)
			return
		}

		// Generate the new image using the selected style
		generatedImg, generatedMimeType, err := gemini.GenerateImage(
			r.Context(),
			s.Logger,
			sessionData.ImageData,
			sessionData.MimeType,
			sessionData.RequestData.EventType,
			sessionData.RequestData.Venue,
			sessionData.RequestData.Theme,
			sessionData.Styles[swapReq.StyleIndex],
		)
		if err != nil {
			s.Logger.Error("Failed to generate swapped image via Gemini", "error", err)
			http.Error(w, "Failed to generate swapped image.", http.StatusInternalServerError)
			return
		}

		// Write the successful response
		w.Header().Set("Content-Type", generatedMimeType)
		w.WriteHeader(http.StatusOK)
		w.Write(generatedImg)
	}
}

// GetStylesHandler handles the /api/v1/styles endpoint.
func GetStylesHandler(s *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		sessionID := r.Header.Get("X-Session-ID")
		if sessionID == "" {
			s.Logger.Error("Missing X-Session-ID header")
			http.Error(w, "Missing X-Session-ID header.", http.StatusBadRequest)
			return
		}

		s.CacheMutex.Lock()
		sessionData, found := s.SessionCache[sessionID]
		s.CacheMutex.Unlock()

		if !found {
			s.Logger.Error("Session data not found for styles request", "sessionID", sessionID)
			http.Error(w, "Session expired or invalid.", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(sessionData.Styles)
	}
}