package middleware

import "net/http"

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain applies middlewares in order (first middleware wraps outermost).
// The execution order will be: first middleware -> second middleware -> ... -> handler.
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// Apply middlewares in reverse order so the first middleware
	// in the list is the outermost (first to execute)
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
