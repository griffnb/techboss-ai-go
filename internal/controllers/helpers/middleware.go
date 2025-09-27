package helpers

import (
	"context"
	"net/http"
	"time"
)

// NoTimeoutMiddleware clears per-connection read/write deadlines
// and removes any context deadline/cancelation. It doesn't hijack;
// the handler continues normally and decides how to respond.
func NoTimeoutMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		_ = rc.SetReadDeadline(time.Time{})
		_ = rc.SetWriteDeadline(time.Time{})

		// strip any router/server-imposed context timeout/cancelation
		// (Go 1.21+: WithoutCancel removes deadline/cancelation)
		ctx := context.WithoutCancel(r.Context())
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// StreamingResponseWriter is a wrapper that ensures the http.Flusher interface is preserved
type StreamingResponseWriter struct {
	http.ResponseWriter
	flusher http.Flusher
}

// NewStreamingResponseWriter creates a new streaming response writer
func NewStreamingResponseWriter(w http.ResponseWriter) *StreamingResponseWriter {
	flusher, ok := w.(http.Flusher)
	if !ok {
		// If the response writer doesn't support flushing, we can't stream
		// This should not happen with standard HTTP servers, but we handle it gracefully
		flusher = nil
	}
	return &StreamingResponseWriter{
		ResponseWriter: w,
		flusher:        flusher,
	}
}

// Flush implements the http.Flusher interface
func (w *StreamingResponseWriter) Flush() {
	if w.flusher != nil {
		w.flusher.Flush()
	}
}

// NoTimeoutStreamingMiddleware is like NoTimeoutMiddleware but preserves flushing capability for SSE
func NoTimeoutStreamingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rc := http.NewResponseController(w)
		_ = rc.SetReadDeadline(time.Time{})
		_ = rc.SetWriteDeadline(time.Time{})

		// Ensure we preserve the flushing capability
		streamingWriter := NewStreamingResponseWriter(w)

		// strip any router/server-imposed context timeout/cancelation
		ctx := context.WithoutCancel(r.Context())
		next.ServeHTTP(streamingWriter, r.WithContext(ctx))
	})
}
