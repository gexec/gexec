package handler

import (
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/manifest"
	"github.com/go-chi/render"
)

// Manifest renders the manifest from embedded frontend.
func (h *Handler) Manifest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := manifest.Read(h.config)

		if err != nil {
			slog.Error(
				"Failed to load manifest",
				slog.Any("error", err),
				slog.String("handler", "manifest"),
			)

			http.Error(
				w,
				"Failed to load manifest",
				http.StatusInternalServerError,
			)

			return
		}

		render.Status(r, http.StatusOK)
		render.JSON(w, r, m)
	}
}
