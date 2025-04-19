package frontend

import (
	"embed"
	"io/fs"
	"log/slog"
	"os"
	"path"

	"github.com/gexec/gexec/pkg/config"
)

var (
	//go:embed files/*
	files embed.FS
)

// Load initializes the frontend files including custom path.
func Load(cfg *config.Config) fs.FS {
	return Chained{
		config: cfg,
	}
}

// Chained is a simple HTTP filesystem including custom path.
type Chained struct {
	config *config.Config
}

// Open just implements the HTTP filesystem interface.
func (c Chained) Open(origPath string) (fs.File, error) {
	if c.config.Server.Frontend != "" {
		if stat, err := os.Stat(c.config.Server.Frontend); err == nil && stat.IsDir() {
			customPath := path.Join(
				c.config.Server.Frontend,
				origPath,
			)

			info, err := os.Stat(customPath)

			if !os.IsNotExist(err) {
				if info.IsDir() {
					return nil, os.ErrPermission
				}

				f, err := os.Open(customPath)

				if err != nil {
					return nil, err
				}

				return f, nil
			}
		} else {
			slog.Warn(
				"Custom frontend directory does not exist",
			)
		}
	}

	content, err := fs.Sub(
		files,
		"files",
	)

	if err != nil {
		return nil, err
	}

	f, err := content.Open(
		origPath,
	)

	if err != nil {
		return nil, err
	}

	info, err := f.Stat()

	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, os.ErrPermission
	}

	return f, nil
}
