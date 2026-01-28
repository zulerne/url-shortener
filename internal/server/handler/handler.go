package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zulerne/url-shortener/internal/server/middleware"
)

// Storage defines the interface for URL storage operations.
// This allows swapping implementations (sqlite, postgres, redis, etc.)
type Storage interface {
	SaveURL(url string, alias string) (int64, error)
	GetURL(alias string) (string, error)
	DeleteURL(alias string) error
}

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	storage     Storage
	validator   *validator.Validate
	aliasLength int
}

// NewHandler creates a new Handler with the given dependencies
func NewHandler(storage Storage, aliasLength int) http.Handler {
	h := &Handler{
		storage:     storage,
		validator:   validator.New(),
		aliasLength: aliasLength,
	}

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /health", h.healthCheck)
	mux.HandleFunc("POST /url", h.createURL)
	//mux.HandleFunc("GET /{alias}", h.redirect)
	//mux.HandleFunc("DELETE /url/{alias}", h.deleteURL)

	// Apply middleware chain (order: first listed = first executed)
	// Recoverer -> RequestID -> Logger -> handler
	return middleware.Chain(mux,
		middleware.Recoverer, // Recover from panics first
		middleware.RequestID,
		middleware.Logger,
	)
}

func (h *Handler) healthCheck(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK\n"))
}

func (h *Handler) renderJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}
