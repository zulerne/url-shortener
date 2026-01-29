package middleware

import (
	"log/slog"
	"net/http"
)

func BasicAuth(userPassMap map[string]string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			slog.Debug("authenticating request")

			authUser, authPass, ok := r.BasicAuth()

			pass, exists := userPassMap[authUser]

			if !ok || !exists || pass != authPass {
				slog.Warn("Unauthorized request")
				w.WriteHeader(http.StatusUnauthorized)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
