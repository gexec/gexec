package handler

import (
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/manifest"
	"github.com/gexec/gexec/pkg/templates"
	"github.com/go-chi/render"
)

// Index renders the template for embedded frontend.
func (h *Handler) Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		m, err := manifest.Read(h.config)

		if err != nil {
			slog.Error(
				"Failed to load manifest",
				slog.Any("error", err),
				slog.String("handler", "index"),
			)

			http.Error(
				w,
				"Failed to load manifest",
				http.StatusInternalServerError,
			)

			return
		}

		render.Status(r, http.StatusOK)
		render.HTML(w, r, templates.String(
			h.config,
			"index.tmpl",
			struct {
				Prefix      string
				Stylesheets []string
				Javascripts []string
			}{
				Prefix:      h.Prefix(),
				Stylesheets: m.Index().Stylehseets,
				Javascripts: []string{
					m.Index().File,
				},
			},
		))
	}
}
