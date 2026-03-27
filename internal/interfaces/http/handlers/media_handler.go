package handlers

import (
	"net/http"

	"blog_api/internal/application"
	"blog_api/internal/application/media"
	"blog_api/internal/interfaces/http/middleware"
)

type MediaHandler struct {
	svc *media.Service
}

func NewMediaHandler(svc *media.Service) *MediaHandler {
	return &MediaHandler{svc: svc}
}

func (h *MediaHandler) Upload(w http.ResponseWriter, r *http.Request) {
	identity, ok := middleware.IdentityFromContext(r.Context())
	if !ok {
		writeError(w, application.ErrUnauthorized)
		return
	}

	// Limit request body to 10 MB + 1 KB overhead for multipart boundaries.
	r.Body = http.MaxBytesReader(w, r.Body, 10<<20+1024)

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	defer file.Close()

	created, err := h.svc.Upload(r.Context(), identity, file, header)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":            created.ID,
		"original_name": created.OriginalName,
		"stored_name":   created.StoredName,
		"content_type":  created.ContentType,
		"size_bytes":    created.SizeBytes,
		"created_at":    created.CreatedAt,
	})
}

func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(extractURLParam(r, "id"))
	if err != nil {
		writeError(w, application.ErrValidation)
		return
	}
	m, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":            m.ID,
		"original_name": m.OriginalName,
		"stored_name":   m.StoredName,
		"content_type":  m.ContentType,
		"size_bytes":    m.SizeBytes,
		"uploader_id":   m.UploaderID,
		"created_at":    m.CreatedAt,
	})
}
