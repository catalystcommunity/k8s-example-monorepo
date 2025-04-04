package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/catalystcommunity/k8s-example-monorepo/app_api/internal/store"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// BaseHandler provides common functionality for all handlers
type BaseHandler struct{}

// respondWithJSON writes a JSON response
func (h *BaseHandler) respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	response, err := json.Marshal(payload)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Internal Server Error","message":"Failed to marshal response"}`))  // Simple fallback
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	w.Write(response)
}

// respondWithError sends a standard error response
func (h *BaseHandler) respondWithError(w http.ResponseWriter, code int, err error) {
	var message string
	var errType string

	switch {
	case errors.Is(err, store.ErrNotFound):
		errType = "not_found"
		message = "Resource not found"
		code = http.StatusNotFound
	case errors.Is(err, store.ErrInvalidInput):
		errType = "invalid_input"
		message = "Invalid input data"
		code = http.StatusBadRequest
	case errors.Is(err, store.ErrAlreadyExists):
		errType = "already_exists"
		message = "Resource already exists"
		code = http.StatusConflict
	case errors.Is(err, store.ErrForbidden):
		errType = "forbidden"
		message = "Permission denied"
		code = http.StatusForbidden
	case errors.Is(err, store.ErrUnauthorized):
		errType = "unauthorized"
		message = "Unauthorized"
		code = http.StatusUnauthorized
	default:
		errType = "internal_error"
		message = "Internal server error"
		code = http.StatusInternalServerError
	}

	h.respondWithJSON(w, code, ErrorResponse{
		Error:   errType,
		Message: message,
	})
}

// getID gets a path parameter ID from the request context
func (h *BaseHandler) getID(r *http.Request, key string) string {
	return GetIDFromContext(r, key)
}
