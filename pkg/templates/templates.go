package templates

import (
	"bytes"
	"embed"
	"html/template"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/drone/funcmap"
	"github.com/gexec/gexec/pkg/config"
	"github.com/rs/zerolog/log"
)

var (
	//go:embed files/*
	assets embed.FS
)

// String renders the template and returns a string.
func String(cfg *config.Config, name string, data any) string {
	tmpls := Load(cfg)
	buffer := bytes.NewBufferString("")

	if err := tmpls.ExecuteTemplate(
		buffer,
		name,
		data,
	); err != nil {
		log.Warn().
			Err(err).
			Str("file", name).
			Msg("failed to parse template content")
	}

	return buffer.String()
}

// Read simply provides an io.Reader for direct usage.
func Read(cfg *config.Config, name string, data any) io.Reader {
	tmpls := Load(cfg)
	buffer := bytes.NewBufferString("")

	if err := tmpls.ExecuteTemplate(
		buffer,
		name,
		data,
	); err != nil {
		log.Warn().
			Err(err).
			Str("file", name).
			Msg("failed to parse template content")
	}

	return bytes.NewReader(
		buffer.Bytes(),
	)
}

// Load loads the template to make it parseable.
func Load(cfg *config.Config) *template.Template {
	tpls := template.New(
		"",
	).Funcs(
		Funcs(),
	)

	files, err := allEmbeddedTemplates(
		&assets,
		"",
	)

	if err != nil {
		log.Warn().
			Err(err).
			Msg("failed to get builtin template list")
	} else {
		for _, name := range files {
			file, err := assets.ReadFile(name)

			if err != nil {
				log.Warn().
					Err(err).
					Str("file", name).
					Msg("failed to read builtin template")
			}

			if _, err := tpls.New(
				strings.TrimPrefix(
					name,
					"files/",
				),
			).Parse(string(file)); err != nil {
				log.Warn().
					Err(err).
					Str("file", name).
					Msg("failed to parse builtin template")
			}
		}
	}

	if cfg.Server.Assets != "" {
		if stat, err := os.Stat(cfg.Server.Assets); err == nil && stat.IsDir() {
			files := []string{}

			_ = filepath.Walk(cfg.Server.Assets, func(path string, f os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				if f.IsDir() {
					return nil
				}

				if !strings.HasSuffix(path, ".html") {
					return nil
				}

				files = append(
					files,
					path,
				)

				return nil
			})

			for _, name := range files {
				file, err := os.ReadFile(name)

				if err != nil {
					log.Warn().
						Err(err).
						Str("file", name).
						Msg("failed to read custom template")
				}

				tplName := strings.TrimPrefix(
					strings.TrimPrefix(
						name,
						cfg.Server.Assets,
					),
					"/",
				)

				if _, err := tpls.New(tplName).Parse(string(file)); err != nil {
					log.Warn().
						Err(err).
						Str("file", name).
						Msg("failed to parse custom template")
				}
			}
		} else {
			log.Warn().
				Msg("custom assets directory doesn't exist")
		}
	}

	return tpls
}

// Funcs provides some general usefule template helpers.
func Funcs() template.FuncMap {
	return funcmap.Funcs
}

func allEmbeddedTemplates(fs *embed.FS, dir string) ([]string, error) {
	if len(dir) == 0 {
		dir = "."
	}

	result := []string{}
	entries, err := fs.ReadDir(dir)

	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		fp := path.Join(
			dir,
			entry.Name(),
		)

		if entry.IsDir() {
			res, err := allEmbeddedTemplates(
				fs,
				fp,
			)

			if err != nil {
				return nil, err
			}

			result = append(
				result,
				res...,
			)

			continue
		}

		if !strings.HasSuffix(fp, ".tmpl") {
			continue
		}

		result = append(
			result,
			fp,
		)
	}

	return result, nil
}
