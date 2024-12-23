package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"
	"time"

	v1 "github.com/genexec/genexec/pkg/api/v1"
	"github.com/genexec/genexec/pkg/authn"
	"github.com/genexec/genexec/pkg/config"
	"github.com/genexec/genexec/pkg/metrics"
	"github.com/genexec/genexec/pkg/middleware/current"
	"github.com/genexec/genexec/pkg/middleware/header"
	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/scim"
	"github.com/genexec/genexec/pkg/store"
	"github.com/genexec/genexec/pkg/token"
	"github.com/genexec/genexec/pkg/upload"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	oamw "github.com/go-openapi/runtime/middleware"
	cgmw "github.com/oapi-codegen/nethttp-middleware"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"
)

// Server initializes the routing of the server.
func Server(
	cfg *config.Config,
	registry *metrics.Metrics,
	identity *authn.Authn,
	uploads upload.Upload,
	storage *store.Store,
) *chi.Mux {
	mux := chi.NewRouter()

	mux.Use(hlog.NewHandler(log.Logger))
	mux.Use(hlog.RemoteAddrHandler("ip"))
	mux.Use(hlog.URLHandler("path"))
	mux.Use(hlog.MethodHandler("method"))
	mux.Use(hlog.RequestIDHandler("request_id", "Request-Id"))

	mux.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Debug().
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("Accesslog")
	}))

	mux.Use(middleware.Timeout(60 * time.Second))
	mux.Use(middleware.RealIP)
	mux.Use(header.Version)
	mux.Use(header.Cache)
	mux.Use(header.Secure)
	mux.Use(header.Options)
	mux.Use(current.Middleware)

	mux.Route(cfg.Server.Root, func(root chi.Router) {
		if cfg.Scim.Enabled {
			srv, err := scim.New(
				scim.WithRoot(
					path.Join(
						cfg.Server.Root,
						"scim",
						"v2",
					),
				),
				scim.WithStore(
					storage.Handle(),
				),
				scim.WithConfig(
					cfg.Scim,
				),
			).Server()

			if err != nil {
				log.Error().
					Err(err).
					Msg("Failed to linitialize scim server")
			}

			root.Mount("/scim/v2", srv)
		}

		root.Route("/v1", func(r chi.Router) {
			swagger, err := v1.GetSwagger()

			if err != nil {
				log.Error().
					Err(err).
					Str("version", "v1").
					Msg("Failed to load openapi spec")
			}

			swagger.Servers = openapi3.Servers{
				{
					URL: cfg.Server.Host + path.Join(
						cfg.Server.Root,
						"v1",
					),
				},
			}

			r.Get("/spec", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "application/json")

				b, err := json.Marshal(swagger)

				if err != nil {
					log.Error().
						Err(err).
						Msg("Failed to generate json response")

					http.Error(
						w,
						http.StatusText(http.StatusUnprocessableEntity),
						http.StatusUnprocessableEntity,
					)

					return
				}

				w.WriteHeader(http.StatusOK)
				_, _ = w.Write(b)
			})

			r.Handle("/docs", oamw.SwaggerUI(oamw.SwaggerUIOpts{
				Path: path.Join(
					cfg.Server.Root,
					"v1",
					"docs",
				),
				SpecURL: cfg.Server.Host + path.Join(
					cfg.Server.Root,
					"v1",
					"spec",
				),
			}, nil))

			r.With(cgmw.OapiRequestValidatorWithOptions(
				swagger,
				&cgmw.Options{
					SilenceServersWarning: true,
					Options: openapi3filter.Options{
						AuthenticationFunc: func(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
							authenticating := &model.User{}
							scheme := input.SecuritySchemeName
							operation := input.RequestValidationInput.Route.Operation.OperationID

							logger := log.With().
								Str("scheme", scheme).
								Str("operation", operation).
								Logger()

							switch scheme {
							case "Header":
								header := input.RequestValidationInput.Request.Header.Get(
									input.SecurityScheme.Name,
								)

								if header == "" {
									return fmt.Errorf("missing authorization header")
								}

								t, err := token.Verify(
									cfg.Token.Secret,
									strings.TrimSpace(
										header,
									),
								)

								if err != nil {
									return fmt.Errorf("failed to parse auth token")
								}

								user, err := storage.Auth.ByID(
									ctx,
									t.Ident,
								)

								if err != nil {
									logger.Error().
										Err(err).
										Str("user", t.Ident).
										Msg("Failed to find user")

									return fmt.Errorf("failed to find user")
								}

								logger.Trace().
									Str("user", t.Login).
									Msg("Authentication")

								authenticating = user

							case "Bearer":
								header := input.RequestValidationInput.Request.Header.Get(
									"Authorization",
								)

								if header == "" {
									return fmt.Errorf("missing authorization bearer")
								}

								t, err := token.Verify(
									cfg.Token.Secret,
									strings.TrimSpace(
										strings.Replace(
											header,
											"Bearer",
											"",
											1,
										),
									),
								)

								if err != nil {
									return fmt.Errorf("failed to parse auth token")
								}

								user, err := storage.Auth.ByID(
									ctx,
									t.Ident,
								)

								if err != nil {
									logger.Error().
										Err(err).
										Str("user", t.Ident).
										Msg("Failed to find user")

									return fmt.Errorf("failed to find user")
								}

								logger.Trace().
									Str("user", t.Login).
									Msg("Authentication")

								authenticating = user

							case "Basic":
								username, password, ok := input.RequestValidationInput.Request.BasicAuth()

								if !ok {
									return fmt.Errorf("missing credentials")
								}

								user, err := storage.Auth.ByCreds(
									ctx,
									username,
									password,
								)

								if err != nil {
									logger.Error().
										Err(err).
										Str("user", username).
										Msg("Wrong credentials")

									return fmt.Errorf("wrong credentials")
								}

								logger.Trace().
									Str("user", username).
									Msg("Authentication")

								authenticating = user

							default:
								return fmt.Errorf("unknown security scheme: %s", scheme)
							}

							log.Trace().
								Str("username", authenticating.Username).
								Str("operation", operation).
								Msg("Authenticated")

							current.SetUser(
								input.RequestValidationInput.Request.Context(),
								authenticating,
							)

							return nil
						},
					},
				},
			)).Mount("/", v1.Handler(
				v1.NewStrictHandler(
					v1.New(
						cfg,
						registry,
						identity,
						uploads,
						storage,
					),
					make([]v1.StrictMiddlewareFunc, 0),
				),
			))

			r.Handle("/storage/*", uploads.Handler(
				path.Join(
					cfg.Server.Root,
					"v1",
					"storage",
				),
			))
		})
	})

	return mux
}

// Metrics initializes the routing of metrics and health.
func Metrics(
	cfg *config.Config,
	registry *metrics.Metrics,
) *chi.Mux {
	mux := chi.NewRouter()

	mux.Use(middleware.Timeout(60 * time.Second))
	mux.Use(middleware.RealIP)
	mux.Use(header.Version)
	mux.Use(header.Cache)
	mux.Use(header.Secure)
	mux.Use(header.Options)

	mux.Route("/", func(root chi.Router) {
		root.Get("/metrics", registry.Handler())

		if cfg.Metrics.Pprof {
			root.Mount("/debug", middleware.Profiler())
		}

		root.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
		})

		root.Get("/readyz", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)

			_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
		})
	})

	return mux
}
