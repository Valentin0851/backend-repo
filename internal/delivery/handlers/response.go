package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Valentin0851/avito-recap-backend/internal/delivery/middleware"
)

type errorEnvelope struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id,omitempty"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	writeJSON(w, status, errorEnvelope{Error: apiError{
		Code:      code,
		Message:   message,
		RequestID: middleware.RequestIDFromContext(r.Context()),
	}})
}

func writeServiceError(w http.ResponseWriter, r *http.Request, logger *slog.Logger, err error) {
	logger.ErrorContext(r.Context(), "request failed",
		slog.String("error", err.Error()),
		slog.String("request_id", middleware.RequestIDFromContext(r.Context())),
	)
	writeError(w, r, http.StatusInternalServerError, "internal_error", "internal server error")
}
