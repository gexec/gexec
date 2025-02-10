package router

import (
	"encoding/json"
	"io"
	"net/http"
	"path"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	v1 "github.com/gexec/gexec/pkg/api/v1"
	"github.com/gexec/gexec/pkg/authn"
	"github.com/gexec/gexec/pkg/config"
	"github.com/gexec/gexec/pkg/handler"
	"github.com/gexec/gexec/pkg/metrics"
	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/middleware/header"
	"github.com/gexec/gexec/pkg/scim"
	"github.com/gexec/gexec/pkg/store"
	"github.com/gexec/gexec/pkg/upload"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
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

	mux.Use(render.SetContentType(render.ContentTypeJSON))
	mux.Use(middleware.Timeout(60 * time.Second))
	mux.Use(middleware.RealIP)
	mux.Use(middleware.Recoverer)
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
						"api",
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

		root.Route("/api/v1", func(r chi.Router) {
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
						"api",
						"v1",
					),
				},
			}

			if cfg.Server.Docs {
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
						"api",
						"v1",
						"docs",
					),
					SpecURL: cfg.Server.Host + path.Join(
						cfg.Server.Root,
						"api",
						"v1",
						"spec",
					),
				}, nil))
			}

			apiv1 := v1.New(
				cfg,
				registry,
				identity,
				uploads,
				storage,
			)

			wrapper := v1.ServerInterfaceWrapper{
				Handler: apiv1,
				ErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
					apiv1.RenderNotify(w, r, v1.Notification{
						Message: v1.ToPtr(err.Error()),
						Status:  v1.ToPtr(http.StatusBadRequest),
					})
				},
			}

			r.With(cgmw.OapiRequestValidatorWithOptions(
				swagger,
				&cgmw.Options{
					SilenceServersWarning: true,
					Options: openapi3filter.Options{
						AuthenticationFunc: apiv1.Authentication,
					},
				},
			)).Route("/", func(r chi.Router) {
				r.Route("/auth", func(r chi.Router) {
					r.Group(func(r chi.Router) {
						r.Post("/login", wrapper.LoginAuth)
						r.Post("/refresh", wrapper.RefreshAuth)
						r.Post("/verify", wrapper.VerifyAuth)
					})

					r.Group(func(r chi.Router) {
						r.Get("/providers", wrapper.ListProviders)

						r.Route("/{provider}", func(r chi.Router) {
							r.Use(render.SetContentType(render.ContentTypeHTML))

							r.Get("/callback", wrapper.CallbackProvider)
							r.Get("/request", wrapper.RequestProvider)
						})
					})
				})

				r.Route("/profile", func(r chi.Router) {
					r.Get("/self", wrapper.ShowProfile)
					r.Put("/self", wrapper.UpdateProfile)
					r.Get("/token", wrapper.TokenProfile)
				})

				r.Route("/projects", func(r chi.Router) {
					r.Get("/", wrapper.ListProjects)
					r.Post("/", wrapper.CreateProject)

					r.Route("/{project_id}", func(r chi.Router) {
						r.Use(apiv1.ProjectToContext)

						r.Get("/", wrapper.ShowProject)
						r.With(apiv1.AllowOwnerProject).Delete("/", wrapper.DeleteProject)
						r.With(apiv1.AllowOwnerProject).Put("/", wrapper.UpdateProject)

						r.Route("/events", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)
							r.Get("/", wrapper.ListProjectEvents)
						})

						r.Route("/groups", func(r chi.Router) {
							r.Use(apiv1.AllowManageProject)

							r.Get("/", wrapper.ListProjectGroups)
							r.Delete("/", wrapper.DeleteProjectFromGroup)
							r.Post("/", wrapper.AttachProjectToGroup)
							r.Put("/", wrapper.PermitProjectGroup)
						})

						r.Route("/users", func(r chi.Router) {
							r.Use(apiv1.AllowManageProject)

							r.Get("/", wrapper.ListProjectUsers)
							r.Delete("/", wrapper.DeleteProjectFromUser)
							r.Post("/", wrapper.AttachProjectToUser)
							r.Put("/", wrapper.PermitProjectUser)
						})

						r.Route("/executions", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)

							r.Get("/", wrapper.ListProjectExecutions)
							r.With(apiv1.AllowManageProjectExecution).Post("/", wrapper.CreateProjectExecution)

							r.Route("/{execution_id}", func(r chi.Router) {
								r.Use(apiv1.ProjectExecutionToContext)

								r.With(apiv1.AllowShowProjectExecution).Get("/", wrapper.ShowProjectExecution)
								r.With(apiv1.AllowManageProjectExecution).Delete("/", wrapper.DeleteProjectExecution)
								r.With(apiv1.AllowManageProjectExecution).Get("/purge", wrapper.PurgeProjectExecution)
								r.With(apiv1.AllowShowProjectExecution).Get("/output", wrapper.OutputProjectExecution)
							})
						})

						r.Route("/schedules", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)

							r.Get("/", wrapper.ListProjectSchedules)
							r.With(apiv1.AllowManageProjectSchedule).Post("/", wrapper.CreateProjectSchedule)

							r.Route("/{schedule_id}", func(r chi.Router) {
								r.Use(apiv1.ProjectScheduleToContext)

								r.With(apiv1.AllowShowProjectSchedule).Get("/", wrapper.ShowProjectSchedule)
								r.With(apiv1.AllowManageProjectSchedule).Delete("/", wrapper.DeleteProjectSchedule)
								r.With(apiv1.AllowManageProjectSchedule).Put("/", wrapper.UpdateProjectSchedule)
							})
						})

						r.Route("/runners", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)

							r.Get("/", wrapper.ListProjectRunners)
							r.With(apiv1.AllowManageProjectRunner).Post("/", wrapper.CreateProjectRunner)

							r.Route("/{runner_id}", func(r chi.Router) {
								r.Use(apiv1.ProjectRunnerToContext)

								r.With(apiv1.AllowShowProjectRunner).Get("/", wrapper.ShowProjectRunner)
								r.With(apiv1.AllowManageProjectRunner).Delete("/", wrapper.DeleteProjectRunner)
								r.With(apiv1.AllowManageProjectRunner).Put("/", wrapper.UpdateProjectRunner)
							})
						})

						r.Route("/credentials", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)

							r.Get("/", wrapper.ListProjectCredentials)
							r.With(apiv1.AllowManageProjectCredential).Post("/", wrapper.CreateProjectCredential)

							r.Route("/{credential_id}", func(r chi.Router) {
								r.Use(apiv1.ProjectCredentialToContext)

								r.With(apiv1.AllowShowProjectCredential).Get("/", wrapper.ShowProjectCredential)
								r.With(apiv1.AllowManageProjectCredential).Delete("/", wrapper.DeleteProjectCredential)
								r.With(apiv1.AllowManageProjectCredential).Put("/", wrapper.UpdateProjectCredential)
							})
						})

						r.Route("/inventories", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)

							r.Get("/", wrapper.ListProjectInventories)
							r.With(apiv1.AllowManageProjectInventory).Post("/", wrapper.CreateProjectInventory)

							r.Route("/{inventory_id}", func(r chi.Router) {
								r.Use(apiv1.ProjectInventoryToContext)

								r.With(apiv1.AllowShowProjectInventory).Get("/", wrapper.ShowProjectInventory)
								r.With(apiv1.AllowManageProjectInventory).Delete("/", wrapper.DeleteProjectInventory)
								r.With(apiv1.AllowManageProjectInventory).Put("/", wrapper.UpdateProjectInventory)
							})
						})

						r.Route("/repositories", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)

							r.Get("/", wrapper.ListProjectRepositories)
							r.With(apiv1.AllowManageProjectRepository).Post("/", wrapper.CreateProjectRepository)

							r.Route("/{repository_id}", func(r chi.Router) {
								r.Use(apiv1.ProjectRepositoryToContext)

								r.With(apiv1.AllowShowProjectRepository).Get("/", wrapper.ShowProjectRepository)
								r.With(apiv1.AllowManageProjectRepository).Delete("/", wrapper.DeleteProjectRepository)
								r.With(apiv1.AllowManageProjectRepository).Put("/", wrapper.UpdateProjectRepository)
							})
						})

						r.Route("/environments", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)

							r.Get("/", wrapper.ListProjectEnvironments)
							r.With(apiv1.AllowManageProjectEnvironment).Post("/", wrapper.CreateProjectEnvironment)

							r.Route("/{environment_id}", func(r chi.Router) {
								r.Use(apiv1.ProjectEnvironmentToContext)

								r.With(apiv1.AllowShowProjectEnvironment).Get("/", wrapper.ShowProjectEnvironment)
								r.With(apiv1.AllowManageProjectEnvironment).Delete("/", wrapper.DeleteProjectEnvironment)
								r.With(apiv1.AllowManageProjectEnvironment).Put("/", wrapper.UpdateProjectEnvironment)

								r.Route("/secrets", func(r chi.Router) {
									r.Use(apiv1.AllowManageProjectEnvironment)
									r.Post("/", wrapper.CreateProjectEnvironmentSecret)

									r.Route("/{secret_id}", func(r chi.Router) {
										r.Use(apiv1.ProjectEnvironmentSecretToContext)

										r.Delete("/", wrapper.DeleteProjectEnvironmentSecret)
										r.Put("/", wrapper.UpdateProjectEnvironmentSecret)
									})
								})

								r.Route("/values", func(r chi.Router) {
									r.Use(apiv1.AllowManageProjectEnvironment)
									r.Post("/", wrapper.CreateProjectEnvironmentValue)

									r.Route("/{value_id}", func(r chi.Router) {
										r.Use(apiv1.ProjectEnvironmentValueToContext)

										r.Delete("/", wrapper.DeleteProjectEnvironmentValue)
										r.Put("/", wrapper.UpdateProjectEnvironmentValue)
									})
								})
							})
						})

						r.Route("/templates", func(r chi.Router) {
							r.Use(apiv1.AllowShowProject)

							r.Get("/", wrapper.ListProjectTemplates)
							r.With(apiv1.AllowManageProjectTemplate).Post("/", wrapper.CreateProjectTemplate)

							r.Route("/{template_id}", func(r chi.Router) {
								r.Use(apiv1.ProjectTemplateToContext)

								r.With(apiv1.AllowShowProjectTemplate).Get("/", wrapper.ShowProjectTemplate)
								r.With(apiv1.AllowManageProjectTemplate).Delete("/", wrapper.DeleteProjectTemplate)
								r.With(apiv1.AllowManageProjectTemplate).Put("/", wrapper.UpdateProjectTemplate)

								r.Route("/surveys", func(r chi.Router) {
									r.Use(apiv1.AllowManageProjectTemplate)
									r.Post("/", wrapper.CreateProjectTemplateSurvey)

									r.Route("/{survey_id}", func(r chi.Router) {
										r.Use(apiv1.ProjectTemplateSurveyToContext)

										r.Delete("/", wrapper.DeleteProjectTemplateSurvey)
										r.Put("/", wrapper.UpdateProjectTemplateSurvey)
									})
								})

								r.Route("/vaults", func(r chi.Router) {
									r.Use(apiv1.AllowManageProjectTemplate)
									r.Post("/", wrapper.CreateProjectTemplateVault)

									r.Route("/{vault_id}", func(r chi.Router) {
										r.Use(apiv1.ProjectTemplateVaultToContext)

										r.Delete("/", wrapper.DeleteProjectTemplateVault)
										r.Put("/", wrapper.UpdateProjectTemplateVault)
									})
								})
							})
						})

					})
				})

				r.Route("/events", func(r chi.Router) {
					r.Get("/", wrapper.ListGlobalEvents)
				})

				r.Route("/runners", func(r chi.Router) {
					r.Use(apiv1.AllowAdminAccessOnly)

					r.Get("/", wrapper.ListGlobalRunners)
					r.Post("/", wrapper.CreateGlobalRunner)

					r.Route("/{runner_id}", func(r chi.Router) {
						r.Use(apiv1.GlobalRunnerToContext)

						r.Get("/", wrapper.ShowGlobalRunner)
						r.Delete("/", wrapper.DeleteGlobalRunner)
						r.Put("/", wrapper.UpdateGlobalRunner)
					})
				})

				r.Route("/groups", func(r chi.Router) {
					r.Use(apiv1.AllowAdminAccessOnly)

					r.Get("/", wrapper.ListGroups)
					r.Post("/", wrapper.CreateGroup)

					r.Route("/{group_id}", func(r chi.Router) {
						r.Use(apiv1.GroupToContext)

						r.Get("/", wrapper.ShowGroup)
						r.Delete("/", wrapper.DeleteGroup)
						r.Put("/", wrapper.UpdateGroup)

						r.Route("/projects", func(r chi.Router) {
							r.Get("/", wrapper.ListGroupProjects)
							r.Delete("/", wrapper.DeleteGroupFromProject)
							r.Post("/", wrapper.AttachGroupToProject)
							r.Put("/", wrapper.PermitGroupProject)
						})

						r.Route("/users", func(r chi.Router) {
							r.Get("/", wrapper.ListGroupUsers)
							r.Delete("/", wrapper.DeleteGroupFromUser)
							r.Post("/", wrapper.AttachGroupToUser)
							r.Put("/", wrapper.PermitGroupUser)
						})
					})
				})

				r.Route("/users", func(r chi.Router) {
					r.Use(apiv1.AllowAdminAccessOnly)

					r.Get("/", wrapper.ListUsers)
					r.Post("/", wrapper.CreateUser)

					r.Route("/{user_id}", func(r chi.Router) {
						r.Use(apiv1.UserToContext)

						r.Get("/", wrapper.ShowUser)
						r.Delete("/", wrapper.DeleteUser)
						r.Put("/", wrapper.UpdateUser)

						r.Route("/projects", func(r chi.Router) {
							r.Get("/", wrapper.ListUserProjects)
							r.Delete("/", wrapper.DeleteUserFromProject)
							r.Post("/", wrapper.AttachUserToProject)
							r.Put("/", wrapper.PermitUserProject)
						})

						r.Route("/groups", func(r chi.Router) {
							r.Get("/", wrapper.ListUserGroups)
							r.Delete("/", wrapper.DeleteUserFromGroup)
							r.Post("/", wrapper.AttachUserToGroup)
							r.Put("/", wrapper.PermitUserGroup)
						})
					})
				})
			})

			r.Handle("/storage/*", uploads.Handler(
				path.Join(
					cfg.Server.Root,
					"v1",
					"storage",
				),
			))
		})

		handlers := handler.New(cfg)
		root.Get("/", handlers.Index())
		root.Get("/favicon.svg", handlers.Favicon())
		root.Get("/config.json", handlers.Config())
		root.Get("/manifest.json", handlers.Manifest())
		root.Handle("/assets/*", handlers.Assets())
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
