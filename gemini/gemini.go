// gemini/gemini.go
package gemini

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/sanjayshr/event-outfitter-backend/models"
	"google.golang.org/genai"
)

// systemPromptTemplate is a detailed, professional prompt based on the prompt guide.
// It instructs the model to perform an image-to-image task, preserving the subject
// while transforming the context (outfit and background).
const systemPromptTemplate = `
A photorealistic close-up portrait of the peoples from the provided image.
Place them in a new context based on the following details:

- Event Type: %s
- Venue / Location: %s
- Theme / Style: %s

Please perform the following transformations:
1.  **Outfit Change:** Dress the people in a stylish, high-fashion outfit that is perfectly appropriate for the specified event and theme.
2.  **Background Replacement:** Completely replace the original background with the specified venue, ensuring it is highly detailed and photorealistic.
3.  **Lighting and Mood:** The scene should be illuminated by soft, golden hour light, creating a warm, celebratory, and masterful atmosphere.
4.  **Identity Preservation:** CRITICAL: The peoples face, features from the original photo must be preserved with high fidelity. Style and pose can be changed based on the event, venue, and theme.

The final image should be captured as if with an 85mm portrait lens, resulting in a soft, blurred background (bokeh), in a vertical portrait orientation.
`

// GenerateImage uses the Gemini API to generate a new image based on a user's photo and text inputs.
func GenerateImage(ctx context.Context, logger *slog.Logger, imgData []byte, mimeType string, reqData models.GenerateRequest) ([]byte, string, error) {
	logger.Info("Starting generare image")
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return nil, "", fmt.Errorf("GEMINI_API_KEY or GOOGLE_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, "", fmt.Errorf("failed to create genai client: %w", err)
	}

	// Construct the detailed prompt using our template
	prompt := fmt.Sprintf(systemPromptTemplate, reqData.EventType, reqData.Venue, reqData.Theme)
	logger.Info("Generated Gemini Prompt", "prompt", prompt)

	// Prepare the multi-modal content (image + text)
	parts := []*genai.Part{
		{Text: prompt},
		{InlineData: &genai.Blob{Data: imgData, MIMEType: mimeType}},
	}
	res, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-image-preview", []*genai.Content{{Parts: parts}}, nil)
	if err != nil {
		logger.Error("Gemini content generation failed", "error", err, "response", res)
		return nil, "", fmt.Errorf("failed to generate content: %w", err)
	}
	logger.Info("Gemini content generation successful", "response", res)

	// Extract the generated image data from the response
	if len(res.Candidates) > 0 && res.Candidates[0].Content != nil {
		for _, part := range res.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				logger.Info("Successfully generated image", "mimeType", part.InlineData.MIMEType, "size_bytes", len(part.InlineData.Data))
				return part.InlineData.Data, part.InlineData.MIMEType, nil
			}
		}
	}

	return nil, "", fmt.Errorf("no image data found in Gemini response")
}
