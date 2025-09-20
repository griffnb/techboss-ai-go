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
