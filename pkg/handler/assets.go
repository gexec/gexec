package handler

import (
	"io/fs"
	"log/slog"
	"net/http"
	"path"

	"github.com/gexec/gexec/pkg/frontend"
)

// Assets provides an handler to read all assets from embedded frontend.
func (h *Handler) Assets() http.Handler {
	content, err := fs.Sub(
		frontend.Load(h.config),
		"assets",
	)

	if err != nil {
		slog.Error(
			"Failed to load assets",
			slog.Any("error", err),
			slog.String("handler", "assets"),
		)
	}

	return http.StripPrefix(
		path.Join(
			h.config.Server.Root,
			"assets",
		)+"/",
		http.FileServer(
			http.FS(
				content,
			),
		),
	)
}
