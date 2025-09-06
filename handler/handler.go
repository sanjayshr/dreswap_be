// handlers.go
package handler

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"path/filepath"
	"strings"

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

		// 3. Call the Gemini service
		generatedImg, generatedMimeType, err := gemini.GenerateImage(r.Context(), s.Logger, imgData, mimeType, reqData)
		if err != nil {
			s.Logger.Error("Failed to generate image via Gemini", "error", err)
			http.Error(w, "Failed to generate image.", http.StatusInternalServerError)
			return
		}

		// 4. Write the successful response
		w.Header().Set("Content-Type", generatedMimeType)
		w.WriteHeader(http.StatusOK)
		w.Write(generatedImg)
	}
}

