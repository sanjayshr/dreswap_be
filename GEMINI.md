# GEMINI.md

## Project Overview

This project is the Go-based backend for **Event Outfitter**, an AI-powered application that reimagines a person's attire and surroundings for a specific event.

The core functionality is an image-to-image generation service. The user provides a portrait photo along with details about an event (e.g., "Wedding," "Beach Party," "Tech Conference"). The backend then uses the Google Gemini 1.5 Pro model to generate a new, photorealistic image, placing the person in a stylish, event-appropriate outfit and a context-relevant background.

The system is designed to preserve the person's identity (face, expression, pose) while creatively transforming their clothing and environment.

### Key Technologies

- **Language:** Go
- **AI Model:**
  image: `gemini-2.5-flash-image-preview`
  text: `gemini-2.5-flash`
- **Key Libraries:**
  - `google.golang.org/genai` for interacting with the Gemini API.

## Project Structure

The project is organized into the following packages:

- `main`: The entry point of the application.
- `server`: Contains the `Server` struct and other server-related code.
- `handler`: Contains the HTTP handlers.
- `gemini`: Contains the code for interacting with the Gemini API.
- `models`: Contains the data models.

## Building and Running

As the project is in its early stages, a formal build and run script has not been established.

### Running the Application

To run the application, you will need to have Go installed and the necessary dependencies fetched.

1.  **Install Dependencies:**

    ```bash
    go mod tidy
    ```

2.  **Run the main application:**

    ```bash
    go run .
    ```

3.  **Set Environment Variables:**
    The application requires a `GEMINI_API_KEY` to be set in the environment.
    ```bash
    export GEMINI_API_KEY="your-api-key"
    ```

## Development Conventions

- **Logging:** The project uses the standard `log/slog` library for structured logging.
- **Configuration:** API keys and other sensitive information are configured via environment variables.
- **Modularity:** The project is organized into packages to promote modularity and separation of concerns.
- **Show code before applying changes:** Always show the code to be changed/updated/created to the user for approval before applying the changes.

## API

The main functionality is exposed via an API endpoint that accepts a multipart form request containing an image and a JSON payload.

### Request

- **Endpoint:** `/api/v1/generate`
- **Method:** `POST`
- **Body:**
  - `image`: The user's portrait photo file.
  - `data`: A JSON string with the following structure, matching the `GenerateRequest` model:
    ```json
    {
      "eventType": "string",
      "venue": "string",
      "theme": "string"
    }
    ```

### Response

The API returns the generated image data (e.g., as a JPEG or PNG) on success or a JSON error message on failure.

## `google.golang.org/genai` Package

When working with the `google.golang.org/genai` package, it is important to refer to the official documentation at [https://pkg.go.dev/google.golang.org/genai](https://pkg.go.dev/google.golang.org/genai) to get the correct methods, parameters, structs, and usage patterns. This will help to avoid compilation errors and ensure that the code is up-to-date with the latest version of the library.

- show the changes to the user first with diff and after approval apply the changes

