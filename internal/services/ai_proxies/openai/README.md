# OpenAI Proxy Service

This package provides a proxy service for the OpenAI API that allows you to forward requests from your Next.js AI SDK frontend through your Go backend. This enables you to keep your API keys secure on the backend while allowing the frontend to use the AI SDK seamlessly.

## Features

- **Non-streaming proxy**: Forward regular chat completion requests
- **Streaming proxy**: Forward streaming chat completion requests with Server-Sent Events (SSE)
- **Request/Response piping**: Complete request and response forwarding with headers
- **Error handling**: Proper error wrapping and handling
- **Configurable**: Customizable timeouts and base URLs for testing

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "log"
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/griffnb/techboss-ai-go/internal/services/ai_proxies/openai"
)

func main() {
    // Create the OpenAI service
    service, err := openai.NewServiceFromEnv()
    if err != nil {
        log.Fatal("Failed to create OpenAI service:", err)
    }

    // Create router
    r := chi.NewRouter()

    // Add routes
    r.Post("/api/ai/chat", handleNonStreamingChat(service))
    r.Post("/api/ai/chat/stream", handleStreamingChat(service))

    // Start server
    log.Fatal(http.ListenAndServe(":8080", r))
}

func handleNonStreamingChat(service *openai.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        err := service.ProxyNonStreaming(r.Context(), r, w)
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            log.Printf("Proxy error: %v", err)
        }
    }
}

func handleStreamingChat(service *openai.Service) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        err := service.ProxyStreaming(r.Context(), r, w)
        if err != nil {
            http.Error(w, "Internal server error", http.StatusInternalServerError)
            log.Printf("Streaming proxy error: %v", err)
        }
    }
}
```

### Environment Variables

Set your OpenAI API key:

```bash
export OPENAI_API_KEY="your-api-key-here"
```

### Frontend Integration

On your Next.js frontend, configure the AI SDK to use your backend endpoints:

```typescript
// For non-streaming
import { generateText } from 'ai';

const result = await generateText({
  model: 'openai:gpt-4', // This will be forwarded to OpenAI
  prompt: 'Hello, world!',
  // Configure to use your backend
  baseURL: 'http://localhost:8080/api/ai', // Your Go backend
});

// For streaming  
import { useChat } from 'ai/react';
import { DefaultChatTransport } from 'ai';

export default function Chat() {
  const { messages, sendMessage } = useChat({
    transport: new DefaultChatTransport({
      api: 'http://localhost:8080/api/ai/chat/stream', // Your Go backend streaming endpoint
    }),
  });
  
  // ... rest of your component
}
```

### Advanced Usage

You can also use the client directly for more control:

```go
import "github.com/griffnb/techboss-ai-go/internal/services/ai_proxies/openai"

// Create client with custom configuration
client := openai.NewClient("your-api-key").
    WithTimeout(60 * time.Second).
    WithBaseURL("https://custom-openai-endpoint.com/v1")

// Use client directly
err := client.ProxyRequest(ctx, requestBody, responseWriter)
```

## API Reference

### Service

#### `NewService(apiKey string) *Service`
Creates a new OpenAI service with the given API key.

#### `NewServiceFromEnv() (*Service, error)`
Creates a new OpenAI service using the `OPENAI_API_KEY` environment variable.

#### `ProxyNonStreaming(ctx context.Context, request *http.Request, responseWriter http.ResponseWriter) error`
Proxies a non-streaming request to OpenAI. The request body is read and forwarded to OpenAI, and the response is piped back.

#### `ProxyStreaming(ctx context.Context, request *http.Request, responseWriter http.ResponseWriter) error`
Proxies a streaming request to OpenAI with Server-Sent Events. The request body is modified to include `stream: true` and the response is streamed back.

### Client

#### `NewClient(apiKey string) *Client`
Creates a new OpenAI client with the given API key.

#### `WithBaseURL(baseURL string) *Client`
Sets a custom base URL for the OpenAI API (useful for testing).

#### `WithTimeout(timeout time.Duration) *Client`
Sets a custom timeout for HTTP requests.

#### `ProxyRequest(ctx context.Context, requestBody []byte, responseWriter http.ResponseWriter) error`
Proxies a request with the given body to OpenAI.

#### `ProxyStreamRequest(ctx context.Context, requestBody []byte, responseWriter http.ResponseWriter) error`
Proxies a streaming request with the given body to OpenAI.

## Request/Response Flow

### Non-Streaming
1. Frontend sends request to your Go backend
2. Backend reads the request body
3. Backend forwards the request to OpenAI with your API key
4. OpenAI responds with the completion
5. Backend pipes the response back to the frontend

### Streaming
1. Frontend sends request to your Go backend streaming endpoint
2. Backend reads the request body and adds `stream: true`
3. Backend forwards the request to OpenAI with your API key
4. OpenAI responds with Server-Sent Events
5. Backend pipes the SSE stream back to the frontend in real-time

## Error Handling

The service handles various error scenarios:

- Invalid request bodies
- Network errors
- OpenAI API errors
- Streaming interruptions

All errors are properly wrapped with context using the `errors` package.

## Testing

Run the tests with:

```bash
go test ./internal/services/ai_proxies/openai -v
```

The tests include:
- Unit tests for client creation and configuration
- Mock server tests for both streaming and non-streaming scenarios
- Error handling verification