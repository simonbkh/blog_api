package handlers

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"blog_api/internal/application"
)

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func decodeJSONBody(r *http.Request, dst any) error {
	defer r.Body.Close()
	limited := io.LimitReader(r.Body, 1<<20)
	decoder := json.NewDecoder(limited)
	decoder.DisallowUnknownFields()
	return decoder.Decode(dst)
}

func writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "internal server error"
	switch {
	case errors.Is(err, application.ErrUnauthorized):
		status = http.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, application.ErrForbidden):
		status = http.StatusForbidden
		message = "forbidden"
	case errors.Is(err, application.ErrNotFound):
		status = http.StatusNotFound
		message = "not found"
	case errors.Is(err, application.ErrValidation):
		status = http.StatusBadRequest
		message = "validation failed"
	case errors.Is(err, application.ErrConflict):
		status = http.StatusConflict
		message = "conflict"
	}
	writeJSON(w, status, map[string]string{"error": message})
}
