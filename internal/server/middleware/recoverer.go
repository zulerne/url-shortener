package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

// Recoverer is a middleware that recovers from panics, logs the panic
// with a stack trace, and returns HTTP 500 (Internal Server Error).
// Based on go-chi/chi middleware.
func Recoverer(next http.Handler) http.Handler {
	slog.Debug("Recoverer middleware enabled")
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rvr := recover(); rvr != nil {
				// Don't recover http.ErrAbortHandler — it's used to abort
				// the response and should not be logged.
				if rvr == http.ErrAbortHandler {
					panic(rvr)
				}

				// Log the panic with stack trace
				slog.Error("Panic recovered",
					"error", rvr,
					"request_id", GetRequestID(r.Context()),
					"method", r.Method,
					"path", r.URL.Path,
					"stack", string(debug.Stack()),
				)

				// Don't write status if connection was upgraded (e.g., WebSocket)
				if r.Header.Get("Connection") != "Upgrade" {
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(w, r)
	})
}
