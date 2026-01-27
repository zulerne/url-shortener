package server

import (
	"net/http"

	"github.com/zulerne/url-shortener/internal/server/middleware"
	"github.com/zulerne/url-shortener/internal/storage"
)

// Handler holds all dependencies for HTTP handlers
type Handler struct {
	storage storage.URLStorage
}

// NewHandler creates a new Handler with the given dependencies
func NewHandler(storage storage.URLStorage) http.Handler {
	h := &Handler{
		storage: storage,
	}

	mux := http.NewServeMux()

	// Register routes
	mux.HandleFunc("GET /health", h.healthCheck)
	mux.HandleFunc("POST /url", h.createShortURL)
	mux.HandleFunc("GET /{alias}", h.redirect)
	mux.HandleFunc("DELETE /url/{alias}", h.deleteURL)

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

func (h *Handler) createShortURL(w http.ResponseWriter, r *http.Request) {

	//url := r.FormValue("url")

	//alias

	//err:=h.storage.SaveURL()

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	// alias := r.PathValue("alias")
	// url, err := h.storage.GetURL(r.Context(), alias)
	w.WriteHeader(http.StatusNotImplemented)
}

func (h *Handler) deleteURL(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
	// alias := r.PathValue("alias")
	// err := h.storage.DeleteURL(r.Context(), alias)
	w.WriteHeader(http.StatusNotImplemented)
}
