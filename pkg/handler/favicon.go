package handler

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/frontend"
)

// Favicon returns the favicon for embedded frontend.
func (h *Handler) Favicon() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, err := frontend.Load(h.config).Open("favicon.svg")

		if err != nil {
			slog.Error(
				"Failed to load favicon",
				slog.Any("error", err),
				slog.String("handler", "favicon"),
			)

			http.Error(
				w,
				"Failed to load favicon",
				http.StatusInternalServerError,
			)

			return
		}

		defer func() { _ = file.Close() }()
		stat, err := file.Stat()

		if err != nil {
			slog.Error(
				"Failed to stat favicon",
				slog.Any("error", err),
				slog.String("handler", "assets"),
			)

			http.Error(
				w,
				"Failed to stat favicon",
				http.StatusInternalServerError,
			)

			return
		}

		http.ServeContent(
			w,
			r,
			"favicon.svg",
			stat.ModTime(),
			file.(io.ReadSeeker),
		)
	}
}
