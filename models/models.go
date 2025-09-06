// models/models.go
package models

// GenerateRequest defines the structure for the JSON data sent from the frontend.
type GenerateRequest struct {
	EventType string `json:"eventType"`
	Venue     string `json:"venue"`
	Theme     string `json:"theme"`
}