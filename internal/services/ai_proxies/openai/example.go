package openai

import (
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ExampleServer demonstrates how to use the OpenAI proxy service
func ExampleServer() {
	// Create the OpenAI service from environment variable
	service, err := NewServiceFromEnv()
	if err != nil {
		log.Fatal("Failed to create OpenAI service. Make sure OPENAI_API_KEY is set:", err)
	}

	// Create router
	r := chi.NewRouter()

	// Add middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Add CORS headers for frontend integration
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	// Health check endpoint
	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("OK")); err != nil {
			log.Printf("Failed to write health check response: %v", err)
		}
	})

	// Non-streaming chat endpoint
	// Frontend can call this with: fetch('/api/ai/chat', { method: 'POST', body: JSON.stringify({...}) })
	r.Post("/api/ai/chat", func(w http.ResponseWriter, r *http.Request) {
		err := service.ProxyNonStreaming(r.Context(), r, w)
		if err != nil {
			log.Printf("Non-streaming proxy error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	// Streaming chat endpoint
	// Frontend can use with AI SDK: useChat({ api: '/api/ai/chat/stream' })
	r.Post("/api/ai/chat/stream", func(w http.ResponseWriter, r *http.Request) {
		err := service.ProxyStreaming(r.Context(), r, w)
		if err != nil {
			log.Printf("Streaming proxy error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	})

	// Start server with proper timeout configuration
	server := &http.Server{
		Addr:         ":8080",
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("Starting server on :8080")
	log.Println("Non-streaming endpoint: POST /api/ai/chat")
	log.Println("Streaming endpoint: POST /api/ai/chat/stream")
	log.Fatal(server.ListenAndServe())
}

// ExampleHandler shows how to create a custom handler function
func ExampleHandler(service *Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Add any custom logic here (auth, rate limiting, logging, etc.)
		log.Printf("Received request from %s", r.RemoteAddr)

		// You can inspect or modify the request here if needed
		// For example, add user context, validate permissions, etc.

		// Proxy the request
		var err error
		if isStreamingRequest(r) {
			err = service.ProxyStreaming(r.Context(), r, w)
		} else {
			err = service.ProxyNonStreaming(r.Context(), r, w)
		}

		if err != nil {
			log.Printf("Proxy error: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
	}
}

// isStreamingRequest determines if a request should be streamed
// You can implement your own logic here
func isStreamingRequest(r *http.Request) bool {
	// Check query parameter
	if r.URL.Query().Get("stream") == "true" {
		return true
	}

	// Check if it's a streaming endpoint
	if r.URL.Path == "/api/ai/chat/stream" {
		return true
	}

	// You could also parse the request body to check for stream: true
	// but that would require reading the body twice

	return false
}

// ExampleWithCustomClient shows how to use a custom client configuration
func ExampleWithCustomClient() {
	// Create client with custom settings
	client := NewClient("your-api-key").
		WithTimeout(60 * time.Second).           // 60 second timeout
		WithBaseURL("https://api.openai.com/v1") // Custom base URL

	// Create service with custom client
	service := &Service{client: client}

	// Use the service...
	_ = service
}
