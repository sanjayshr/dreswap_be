# Event Outfitter AI

Event Outfitter is an AI-powered Go application that reimagines a person's attire and surroundings for any event.

Users provide a portrait photo and describe an event (e.g., "Wedding in Goa," "Tech Conference," "Beach Party"). The application uses the Google Gemini Pro model to generate a new, photorealistic image, placing the person in a stylish, event-appropriate outfit and a context-relevant background, while preserving their identity.

## Features

*   **AI Image Generation**: Transforms a user's photo into a new, context-aware image.
*   **Dynamic Style Suggestions**: Generates multiple unique fashion suggestions for each request.
*   **Interactive Style Swapping**: Allows users to regenerate the image with different generated styles within the same session.
*   **Session-Based Workflow**: Caches user data and style suggestions for a seamless experience.

## How It Works

1.  The user uploads an image and provides event details to the `/generate` endpoint.
2.  The backend uses the Gemini text model to generate 5 creative apparel descriptions (styles).
3.  The backend uses the Gemini vision model to generate a new image based on the user's photo and the *first* style suggestion.
4.  A unique `session-id` is created and returned. This ID is the key to the user's uploaded photo and the list of 5 style suggestions.
5.  The user can then call the `/swap-style` endpoint with the `session-id` and a style index (0-4) to generate a new image with a different outfit.
6.  The user can also call the `/styles` endpoint to retrieve the list of all generated style descriptions for their session.

## Technology Stack

*   **Language**: Go
*   **AI Model**: Google Gemini (`gemini-2.5-flash-image-preview` for images, `gemini-2.5-flash` for text)
*   **Key Libraries**:
    *   `google.golang.org/genai` for Gemini API interaction.
    *   Standard Go libraries for the HTTP server (`net/http`).

## Getting Started

### Prerequisites

*   Go 1.21+ installed.
*   A Google Gemini API Key.

### Installation & Setup

1.  **Clone the repository:**
    ```bash
    git clone <repository-url>
    cd event-outfitter-backend
    ```

2.  **Install dependencies:**
    ```bash
    go mod tidy
    ```

3.  **Set the environment variable:**
    The application requires a `GEMINI_API_KEY` to be set.
    ```bash
    export GEMINI_API_KEY="your-gemini-api-key"
    ```

4.  **Run the application:**
    ```bash
    go run .
    ```
    The server will start on `http://localhost:8081`.

## API Reference

The server provides three main endpoints to interact with the service.

---

### 1. Generate Initial Image

Generates the first styled image and initiates a new session.

*   **URL**: `/api/v1/generate`
*   **Method**: `POST`
*   **Content-Type**: `multipart/form-data`

**Request Body:**

*   `image`: The user's portrait photo file (e.g., `.jpg`, `.png`).
*   `data`: A JSON string with the event details.
    *   `eventType` (string): The type of event.
    *   `venue` (string): The location or venue.
    *   `theme` (string): The theme of the event.

**Response:**

*   **On Success**:
    *   **Status**: `200 OK`
    *   **Headers**: `X-Session-ID: <your-new-session-id>`
    *   **Body**: The raw image data of the generated picture.
*   **On Failure**:
    *   **Status**: `4xx` or `5xx`
    *   **Body**: A JSON error message.

**Example `curl` Request:**

```bash
curl -X POST http://localhost:8081/api/v1/generate \
  -F "image=@/path/to/your/person.jpg" \
  -F 'data={"eventType": "Wedding", "venue": "Goa, India", "theme": "South style wedding"}' \
  --output output.jpg --dump-header -
```
*(Note: The session ID will be printed in the response headers)*

---

### 2. Get Style Suggestions

Retrieves the list of all generated style descriptions for a session.

*   **URL**: `/api/v1/styles`
*   **Method**: `GET`

**Request Headers:**

*   `X-Session-ID`: The session ID returned from the `/generate` request.

**Response:**

*   **On Success**:
    *   **Status**: `200 OK`
    *   **Content-Type**: `application/json`
    *   **Body**: A JSON array of strings, where each string is a style description.
      ```json
      [
        "a traditional silk saree in vibrant colors with intricate gold embroidery",
        "a modern lehenga with a minimalist design and pastel shades",
        ...
      ]
      ```

**Example `curl` Request:**

```bash
# Replace <your-session-id> with the actual ID from the generate step
curl -X GET http://localhost:8081/api/v1/styles \
  -H "X-Session-ID: <your-session-id>"
```

---

### 3. Swap Style

Generates a new image using a different style from the list of suggestions.

*   **URL**: `/api/v1/swap-style`
*   **Method**: `POST`
*   **Content-Type**: `application/json`

**Request Headers:**

*   `X-Session-ID`: The session ID returned from the `/generate` request.

**Request Body:**

*   `styleIndex` (integer): The index of the desired style from the list (0-4).

**Response:**

*   **On Success**:
    *   **Status**: `200 OK`
    *   **Body**: The raw image data of the newly generated picture.

**Example `curl` Request:**

```bash
# Replace <your-session-id> with the actual ID
curl -X POST http://localhost:8081/api/v1/swap-style \
  -H "Content-Type: application/json" \
  -H "X-Session-ID: <your-session-id>" \
  -d '{"styleIndex": 2}' \
  --output swapped_image_style_2.jpg
```

## Project Structure

```
/
├── gemini/       # Logic for interacting with the Gemini API.
├── handler/      # HTTP handlers for the API endpoints.
├── models/       # Go structs for API request/response models.
├── server/       # Server setup and session management.
├── main.go       # Main application entry point.
├── go.mod/go.sum # Go module dependency information.
└── README.md     # This file.
```
