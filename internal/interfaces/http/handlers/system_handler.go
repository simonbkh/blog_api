package handlers

import (
	"net/http"
	"time"
)

type SystemHandler struct {
	startedAt time.Time
}

func NewSystemHandler(startedAt time.Time) *SystemHandler {
	return &SystemHandler{startedAt: startedAt}
}

func (h *SystemHandler) Health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"started_at": h.startedAt,
		"uptime":     time.Since(h.startedAt).String(),
	})
}
