package v1

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/store"
	"github.com/go-chi/chi/v5"
)

const (
	projectContext           contextKey = "project"
	executionContext         contextKey = "execution"
	scheduleContext          contextKey = "schedule"
	runnerContext            contextKey = "runner"
	credentialContext        contextKey = "credential"
	inventoryContext         contextKey = "inventory"
	repositoryContext        contextKey = "repository"
	environmentContext       contextKey = "environment"
	environmentSecretContext contextKey = "environment_secret"
	environmentValueContext  contextKey = "environment_value"
	templateContext          contextKey = "template"
	templateSurveyContext    contextKey = "template_survey"
	templateVaultContext     contextKey = "template_vault"
	groupContext             contextKey = "group"
	userContext              contextKey = "user"
)

// ProjectToContext is used to put the requested project into the context.
func (a *API) ProjectToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "project_id")

		record, err := a.storage.Projects.Show(
			ctx,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrProjectNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find project"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load project",
				slog.Any("error", err),
				slog.String("action", "ProjectToContext"),
				slog.String("project", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load project"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			projectContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectFromContext is used to get the requested project from the context.
func (a *API) ProjectFromContext(ctx context.Context) *model.Project {
	record, ok := ctx.Value(projectContext).(*model.Project)

	if !ok {
		return nil
	}

	return record
}

// ProjectExecutionToContext is used to put the requested execution into the context.
func (a *API) ProjectExecutionToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		project := a.ProjectFromContext(ctx)

		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "execution_id")

		record, err := a.storage.Executions.Show(
			ctx,
			project,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrExecutionNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find execution"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load execution",
				slog.Any("error", err),
				slog.String("action", "ProjectExecutionToContext"),
				slog.String("project", project.ID),
				slog.String("execution", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load execution"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			executionContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectExecutionFromContext is used to get the requested execution from the context.
func (a *API) ProjectExecutionFromContext(ctx context.Context) *model.Execution {
	record, ok := ctx.Value(executionContext).(*model.Execution)

	if !ok {
		return nil
	}

	return record
}

// ProjectScheduleToContext is used to put the requested schedule into the context.
func (a *API) ProjectScheduleToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		project := a.ProjectFromContext(ctx)

		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "schedule_id")

		record, err := a.storage.Schedules.Show(
			ctx,
			project,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrScheduleNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find schedule"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load schedule",
				slog.Any("error", err),
				slog.String("action", "ProjectScheduleToContext"),
				slog.String("project", project.ID),
				slog.String("schedule", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load schedule"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			scheduleContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectScheduleFromContext is used to get the requested schedule from the context.
func (a *API) ProjectScheduleFromContext(ctx context.Context) *model.Schedule {
	record, ok := ctx.Value(scheduleContext).(*model.Schedule)

	if !ok {
		return nil
	}

	return record
}

// ProjectRunnerToContext is used to put the requested runner into the context.
func (a *API) ProjectRunnerToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		project := a.ProjectFromContext(ctx)

		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "runner_id")

		record, err := a.storage.Runners.Show(
			ctx,
			project,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrRunnerNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find runner"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load runner",
				slog.Any("error", err),
				slog.String("action", "ProjectRunnerToContext"),
				slog.String("project", project.ID),
				slog.String("runner", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load runner"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			runnerContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectRunnerFromContext is used to get the requested runner from the context.
func (a *API) ProjectRunnerFromContext(ctx context.Context) *model.Runner {
	record, ok := ctx.Value(runnerContext).(*model.Runner)

	if !ok {
		return nil
	}

	return record
}

// ProjectCredentialToContext is used to put the requested credential into the context.
func (a *API) ProjectCredentialToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		project := a.ProjectFromContext(ctx)

		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "credential_id")

		record, err := a.storage.Credentials.Show(
			ctx,
			project,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrCredentialNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find credential"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load credential",
				slog.Any("error", err),
				slog.String("action", "ProjectCredentialToContext"),
				slog.String("project", project.ID),
				slog.String("credential", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load credential"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			credentialContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectCredentialFromContext is used to get the requested credential from the context.
func (a *API) ProjectCredentialFromContext(ctx context.Context) *model.Credential {
	record, ok := ctx.Value(credentialContext).(*model.Credential)

	if !ok {
		return nil
	}

	return record
}

// ProjectInventoryToContext is used to put the requested inventory into the context.
func (a *API) ProjectInventoryToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		project := a.ProjectFromContext(ctx)

		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "inventory_id")

		record, err := a.storage.Inventories.Show(
			ctx,
			project,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrInventoryNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find inventory"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load inventory",
				slog.Any("error", err),
				slog.String("action", "ProjectInventoryToContext"),
				slog.String("project", project.ID),
				slog.String("inventory", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load inventory"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			inventoryContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectInventoryFromContext is used to get the requested inventory from the context.
func (a *API) ProjectInventoryFromContext(ctx context.Context) *model.Inventory {
	record, ok := ctx.Value(inventoryContext).(*model.Inventory)

	if !ok {
		return nil
	}

	return record
}

// ProjectRepositoryToContext is used to put the requested repository into the context.
func (a *API) ProjectRepositoryToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		project := a.ProjectFromContext(ctx)

		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "repository_id")

		record, err := a.storage.Repositories.Show(
			ctx,
			project,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrRepositoryNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find repository"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load repository",
				slog.Any("error", err),
				slog.String("action", "ProjectRepositoryToContext"),
				slog.String("project", project.ID),
				slog.String("repository", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load repository"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			repositoryContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectRepositoryFromContext is used to get the requested repository from the context.
func (a *API) ProjectRepositoryFromContext(ctx context.Context) *model.Repository {
	record, ok := ctx.Value(repositoryContext).(*model.Repository)

	if !ok {
		return nil
	}

	return record
}

// ProjectEnvironmentToContext is used to put the requested environment into the context.
func (a *API) ProjectEnvironmentToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		project := a.ProjectFromContext(ctx)

		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "environment_id")

		record, err := a.storage.Environments.Show(
			ctx,
			project,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrEnvironmentNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find environment"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load environment",
				slog.Any("error", err),
				slog.String("action", "ProjectEnvironmentToContext"),
				slog.String("project", project.ID),
				slog.String("environment", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load environment"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			environmentContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectEnvironmentFromContext is used to get the requested environment from the context.
func (a *API) ProjectEnvironmentFromContext(ctx context.Context) *model.Environment {
	record, ok := ctx.Value(environmentContext).(*model.Environment)

	if !ok {
		return nil
	}

	return record
}

// ProjectEnvironmentSecretToContext is used to put the requested environment secret into the context.
func (a *API) ProjectEnvironmentSecretToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		project := a.ProjectFromContext(ctx)
		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		environment := a.ProjectEnvironmentFromContext(ctx)
		if environment == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find environment"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "secret_id")

		record, err := a.storage.Environments.ShowSecret(
			ctx,
			environment,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrEnvironmentSecretNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find environment secret"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load environment secret",
				slog.Any("error", err),
				slog.String("action", "ProjectEnvironmentSecretToContext"),
				slog.String("project", project.ID),
				slog.String("environment", environment.ID),
				slog.String("secret", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load environment secret"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			environmentSecretContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectEnvironmentSecretFromContext is used to get the requested environment secret from the context.
func (a *API) ProjectEnvironmentSecretFromContext(ctx context.Context) *model.EnvironmentSecret {
	record, ok := ctx.Value(environmentSecretContext).(*model.EnvironmentSecret)

	if !ok {
		return nil
	}

	return record
}

// ProjectEnvironmentValueToContext is used to put the requested environment value into the context.
func (a *API) ProjectEnvironmentValueToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		project := a.ProjectFromContext(ctx)
		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		environment := a.ProjectEnvironmentFromContext(ctx)
		if environment == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find environment"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "value_id")

		record, err := a.storage.Environments.ShowValue(
			ctx,
			environment,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrEnvironmentValueNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find environment value"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load environment value",
				slog.Any("error", err),
				slog.String("action", "ProjectEnvironmentValueToContext"),
				slog.String("project", project.ID),
				slog.String("environment", environment.ID),
				slog.String("value", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load environment value"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			environmentValueContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectEnvironmentValueFromContext is used to get the requested environment value from the context.
func (a *API) ProjectEnvironmentValueFromContext(ctx context.Context) *model.EnvironmentValue {
	record, ok := ctx.Value(environmentValueContext).(*model.EnvironmentValue)

	if !ok {
		return nil
	}

	return record
}

// ProjectTemplateToContext is used to put the requested template into the context.
func (a *API) ProjectTemplateToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		project := a.ProjectFromContext(ctx)

		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "template_id")

		record, err := a.storage.Templates.Show(
			ctx,
			project,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrTemplateNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find template"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load template",
				slog.Any("error", err),
				slog.String("action", "ProjectTemplateToContext"),
				slog.String("project", project.ID),
				slog.String("template", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load template"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			templateContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectTemplateFromContext is used to get the requested template from the context.
func (a *API) ProjectTemplateFromContext(ctx context.Context) *model.Template {
	record, ok := ctx.Value(templateContext).(*model.Template)

	if !ok {
		return nil
	}

	return record
}

// ProjectTemplateSurveyToContext is used to put the requested template survey into the context.
func (a *API) ProjectTemplateSurveyToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		project := a.ProjectFromContext(ctx)
		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		template := a.ProjectTemplateFromContext(ctx)
		if template == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find template"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "survey_id")

		record, err := a.storage.Templates.ShowSurvey(
			ctx,
			template,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrTemplateSurveyNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find template survey"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load template survey",
				slog.Any("error", err),
				slog.String("action", "ProjectTemplateSurveyToContext"),
				slog.String("project", project.ID),
				slog.String("template", template.ID),
				slog.String("survey", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load template survey"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			templateSurveyContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectTemplateSurveyFromContext is used to get the requested template survey from the context.
func (a *API) ProjectTemplateSurveyFromContext(ctx context.Context) *model.TemplateSurvey {
	record, ok := ctx.Value(templateSurveyContext).(*model.TemplateSurvey)

	if !ok {
		return nil
	}

	return record
}

// ProjectTemplateVaultToContext is used to put the requested template vault into the context.
func (a *API) ProjectTemplateVaultToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		project := a.ProjectFromContext(ctx)
		if project == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find project"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		template := a.ProjectTemplateFromContext(ctx)
		if template == nil {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to find template"),
				Status:  ToPtr(http.StatusNotFound),
			})

			return
		}

		id := chi.URLParam(r, "vault_id")

		record, err := a.storage.Templates.ShowVault(
			ctx,
			template,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrTemplateVaultNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find template vault"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load template vault",
				slog.Any("error", err),
				slog.String("action", "ProjectTemplateVaultToContext"),
				slog.String("project", project.ID),
				slog.String("template", template.ID),
				slog.String("vault", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load template vault"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			templateVaultContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ProjectTemplateVaultFromContext is used to get the requested template vault from the context.
func (a *API) ProjectTemplateVaultFromContext(ctx context.Context) *model.TemplateVault {
	record, ok := ctx.Value(templateVaultContext).(*model.TemplateVault)

	if !ok {
		return nil
	}

	return record
}

// GlobalRunnerToContext is used to put the requested runner into the context.
func (a *API) GlobalRunnerToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "runner_id")

		record, err := a.storage.Runners.Show(
			ctx,
			&model.Project{},
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrRunnerNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find runner"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load runner",
				slog.Any("error", err),
				slog.String("action", "GlobalRunnerToContext"),
				slog.String("runner", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load runner"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			runnerContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GlobalRunnerFromContext is used to get the requested runner from the context.
func (a *API) GlobalRunnerFromContext(ctx context.Context) *model.Runner {
	record, ok := ctx.Value(runnerContext).(*model.Runner)

	if !ok {
		return nil
	}

	return record
}

// GroupToContext is used to put the requested group into the context.
func (a *API) GroupToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "group_id")

		record, err := a.storage.Groups.Show(
			ctx,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrGroupNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find group"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load group",
				slog.Any("error", err),
				slog.String("action", "GroupToContext"),
				slog.String("group", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load group"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			groupContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GroupFromContext is used to get the requested group from the context.
func (a *API) GroupFromContext(ctx context.Context) *model.Group {
	record, ok := ctx.Value(groupContext).(*model.Group)

	if !ok {
		return nil
	}

	return record
}

// UserToContext is used to put the requested user into the context.
func (a *API) UserToContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := chi.URLParam(r, "user_id")

		record, err := a.storage.Users.Show(
			ctx,
			id,
		)

		if err != nil {
			if errors.Is(err, store.ErrUserNotFound) {
				a.RenderNotify(w, r, Notification{
					Message: ToPtr("Failed to find user"),
					Status:  ToPtr(http.StatusNotFound),
				})

				return
			}

			slog.Error(
				"Failed to load user",
				slog.Any("error", err),
				slog.String("action", "UserToContext"),
				slog.String("user", id),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to load user"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		ctx = context.WithValue(
			ctx,
			userContext,
			record,
		)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// UserFromContext is used to get the requested user from the context.
func (a *API) UserFromContext(ctx context.Context) *model.User {
	record, ok := ctx.Value(userContext).(*model.User)

	if !ok {
		return nil
	}

	return record
}
