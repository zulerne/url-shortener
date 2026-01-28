package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/zulerne/url-shortener/internal/lib/random"
	"github.com/zulerne/url-shortener/internal/server/middleware"
	"github.com/zulerne/url-shortener/internal/server/response"
	"github.com/zulerne/url-shortener/internal/storage"
)

type CreateURLRequest struct {
	URL   string `json:"url" validate:"required,url"`
	Alias string `json:"alias,omitempty"`
}

type CreateURLResponse struct {
	response.Response
	Alias string `json:"alias,omitempty"`
}

func (h *Handler) createURL(w http.ResponseWriter, r *http.Request) {
	const op = "handler.save.createURL"
	log := slog.With(
		"op", op,
		string(middleware.RequestIDKey), middleware.GetRequestID(r.Context()),
	)

	var req CreateURLRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Error("failed to decode request body", "error", err)
		h.renderJSON(w, http.StatusBadRequest, response.Error(err.Error()))
		return
	}
	log.Info("request received", "request", req)

	if err := h.validator.Struct(req); err != nil {
		log.Error("validation error", "error", err)

		var validationErr validator.ValidationErrors
		if !errors.As(err, &validationErr) {
			log.Error("unexpected validation error", "error", err)
			h.renderJSON(w, http.StatusInternalServerError, response.Error("internal error"))
			return
		}

		h.renderJSON(w, http.StatusBadRequest, response.ValidationError(validationErr))
		return
	}

	alias := req.Alias
	if alias == "" {
		alias = random.Alias(h.aliasLength)
	}

	id, err := h.storage.SaveURL(req.URL, alias)
	if err != nil {
		log.Error("failed to save url", "error", err)

		if errors.Is(err, storage.ErrAliasExists) {
			h.renderJSON(w, http.StatusConflict, response.Error(storage.ErrAliasExists.Error()))
			return
		}

		h.renderJSON(w, http.StatusInternalServerError, response.Error("failed to save url"))
		return
	}

	log.Info("url saved", "id", id, "alias", alias)

	h.renderJSON(w, http.StatusOK, CreateURLResponse{
		Response: response.Ok(),
		Alias:    alias,
	})
}
