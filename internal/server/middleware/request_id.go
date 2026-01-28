package middleware

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"sync/atomic"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey contextKey = "request_id"
	// RequestIDHeader is the header name for request ID
	RequestIDHeader = "X-Request-ID"
)

var reqid atomic.Uint64

// RequestID middleware adds a unique request ID to each request context.
// If X-Request-ID header is present in the request, it uses that value.
// Otherwise, it generates a new unique ID.
func RequestID(next http.Handler) http.Handler {
	slog.Debug("RequestID middleware enabled")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			myid := reqid.Add(1)
			requestID = fmt.Sprintf("%016x", myid)
		}

		// Set the request ID in response header
		w.Header().Set(RequestIDHeader, requestID)

		ctx = context.WithValue(ctx, RequestIDKey, requestID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetRequestID extracts the request ID from context.
// Returns empty string if not found.
func GetRequestID(ctx context.Context) string {
	if id, ok := ctx.Value(RequestIDKey).(string); ok {
		return id
	}
	return ""
}
