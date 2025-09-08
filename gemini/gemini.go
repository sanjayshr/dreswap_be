// gemini/gemini.go
package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"google.golang.org/genai"
)

// systemPromptTemplate is a detailed, professional prompt based on the prompt guide.
// It instructs the model to perform an image-to-image task, preserving the subject
// while transforming the context (outfit and background).
const systemPromptTemplate = `
A photorealistic close-up portrait of the people from the provided image.
Place them in a new context for a '{eventType}' at '{venue}' with the theme '{theme}'.

**CRITICAL INSTRUCTION:** Dress the people in a very specific, stylish, high-fashion outfit that perfectly matches this detailed description: %s.

Ensure the background, lighting, and mood are photorealistic and match the event.
Preserve the people's faces and features from the original photo. Style and pose can be changed to fit the outfit.
The final image should be captured with an 85mm portrait lens with a soft, blurred background.
`

// GenerateImage uses the Gemini API to generate a new image based on a user's photo and text inputs.
func GenerateImage(ctx context.Context, logger *slog.Logger, imgData []byte, mimeType string, eventType, venue, theme, styleDescription string) ([]byte, string, error) {
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
	prompt := fmt.Sprintf(systemPromptTemplate, eventType, venue, theme, styleDescription)
	logger.Info("Generated Gemini Prompt", "prompt", prompt)

	// Prepare the multi-modal content (image + text)
	parts := []*genai.Part{
		{Text: prompt},
		{InlineData: &genai.Blob{Data: imgData, MIMEType: mimeType}},
	}

	// Define safety settings to block only high-probability harmful content.
	safetySettings := []*genai.SafetySetting{
		{
			Category:  genai.HarmCategoryHarassment,
			Threshold: genai.HarmBlockThresholdBlockNone,
		},
		{
			Category:  genai.HarmCategoryHateSpeech,
			Threshold: genai.HarmBlockThresholdBlockNone,
		},
		{
			Category:  genai.HarmCategorySexuallyExplicit,
			Threshold: genai.HarmBlockThresholdBlockNone,
		},
		{
			Category:  genai.HarmCategoryDangerousContent,
			Threshold: genai.HarmBlockThresholdBlockNone,
		},
	}

	// Use the correct GenerateContentConfig struct to pass the settings.
	config := &genai.GenerateContentConfig{
		SafetySettings: safetySettings,
	}

	res, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash-image-preview", []*genai.Content{{Parts: parts}}, config)
	if err != nil {
		logger.Error("Gemini text content generation failed", "error", err, "response", res)
		return nil, "", fmt.Errorf("failed to generate prmots(text): %w", err)
	}
	logger.Info("Gemini content generation successful")

	// Extract the generated image data from the response
	if len(res.Candidates) > 0 && res.Candidates[0].Content != nil {
		for _, part := range res.Candidates[0].Content.Parts {
			if part.InlineData != nil {
				logger.Info("Successfully generated image", "mimeType", part.InlineData.MIMEType, "size_bytes", len(part.InlineData.Data))
				return part.InlineData.Data, part.InlineData.MIMEType, nil
			}
		}
	}

	// If we reach here, no image data was found. Log the full response for debugging.
	logger.Error("No image data found in Gemini response", "full_response", res)
	return nil, "", fmt.Errorf("no image data found in Gemini response")
}

// GetStyleSuggestions uses the Gemini API to generate a list of style suggestions based on event details.
func GetStyleSuggestions(ctx context.Context, logger *slog.Logger, eventType, venue, theme string) ([]string, error) {
	apiKey := os.Getenv("GOOGLE_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("GEMINI_API_KEY")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY or GOOGLE_API_KEY environment variable not set")
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: apiKey})
	if err != nil {
		return nil, fmt.Errorf("failed to create genai client: %w", err)
	}
	prompt := fmt.Sprintf(`Based on the person in the user's photo, identify their likely gender. Then, for an event '%s' at location '%s' with the theme '%s', generate a JSON array of 5 distinct and creative fashion apparel descriptions for them.Be specific and evocative.Example for a man: ["a crisp white linen shirt with tailored khaki shorts and leather sandals", "a lightweight navy blazer over a crew-neck t-shirt and chinos"].Example for a woman: ["a vibrant tropical print maxi dress with woven sandals", "bohemian chic with a crochet top and a flowy tiered skirt"].`, eventType, venue, theme)
	// Construct the prompt for style suggestions
	logger.Info("Generated Style Suggestion Prompt", "prompt", prompt)

	res, err := client.Models.GenerateContent(ctx, "gemini-2.5-flash", genai.Text(prompt), nil)
	if err != nil {
		logger.Error("Gemini style suggestion generation failed", "error", err, "response", res)
		return nil, fmt.Errorf("failed to generate style suggestions: %w", err)
	}
	logger.Info("Gemini style suggestion generation successful", "response", res)

	if len(res.Candidates) > 0 && res.Candidates[0].Content != nil {
		var fullResponseText string
		for _, part := range res.Candidates[0].Content.Parts {
			if part.Text != "" {
				fullResponseText += part.Text
			}
		}

		if fullResponseText == "" {
			return nil, fmt.Errorf("no text content found in Gemini response")
		}

		// Now, proceed with your existing JSON parsing logic on the fullResponseText
		logger.Info("Received text response for style suggestions", "text", fullResponseText)

		startIndex := strings.Index(fullResponseText, "[")
		endIndex := strings.LastIndex(fullResponseText, "]")

		if startIndex == -1 || endIndex == -1 || endIndex < startIndex {
			return nil, fmt.Errorf("could not find a valid JSON array in the AI response: %s", fullResponseText)
		}

		jsonString := fullResponseText[startIndex : endIndex+1]

		var styles []string
		if err := json.Unmarshal([]byte(jsonString), &styles); err != nil {
			return nil, fmt.Errorf("failed to unmarshal style suggestions JSON: %w; raw response: %s", err, jsonString)
		}

		return styles, nil
	}

	return nil, fmt.Errorf("no style suggestions found in Gemini response")
}
