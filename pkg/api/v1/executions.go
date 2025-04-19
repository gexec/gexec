package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
)

// ListProjectExecutions implements the v1.ServerInterface.
func (a *API) ListProjectExecutions(w http.ResponseWriter, r *http.Request, _ ProjectID, params ListProjectExecutionsParams) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	sort, order, limit, offset, search := listExecutionsSorting(params)

	records, count, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.List(
		ctx,
		project.ID,
		model.ListParams{
			Sort:   sort,
			Order:  order,
			Limit:  limit,
			Offset: offset,
			Search: search,
		},
	)

	if err != nil {
		slog.Error(
			"Failed to load executions",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "ListProjectExecutions"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load executions"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Execution, len(records))
	for id, record := range records {
		if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
			slog.Error(
				"Failed to decrypt secrets",
				slog.Any("error", err),
				slog.String("project", project.ID),
				slog.String("action", "ListProjectExecutions"),
			)

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to decrypt secrets"),
				Status:  ToPtr(http.StatusInternalServerError),
			})

			return
		}

		payload[id] = a.convertExecution(record)
	}

	render.JSON(w, r, ProjectExecutionsResponse{
		Total:      count,
		Limit:      limit,
		Offset:     offset,
		Project:    ToPtr(a.convertProject(project)),
		Executions: payload,
	})
}

// ShowProjectExecution implements the v1.ServerInterface.
func (a *API) ShowProjectExecution(w http.ResponseWriter, r *http.Request, _ ProjectID, _ ExecutionID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectExecutionFromContext(ctx)

	if err := record.DeserializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to decrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("execution", project.ID),
			slog.String("action", "ShowProjectExecution"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectExecutionResponse(
		a.convertExecution(record),
	))
}

// CreateProjectExecution implements the v1.ServerInterface.
func (a *API) CreateProjectExecution(w http.ResponseWriter, r *http.Request, _ ProjectID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	body := &CreateProjectExecutionBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectExecution"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	record := &model.Execution{
		ProjectID: project.ID,
	}

	if body.TemplateID != nil {
		record.TemplateID = FromPtr(body.TemplateID)
	}

	if err := record.SerializeSecret(a.config.Encrypt.Passphrase); err != nil {
		slog.Error(
			"Failed to encrypt secrets",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectExecution"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to encrypt secrets"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.Create(
		ctx,
		project,
		record,
	); err != nil {
		if v, ok := err.(validate.Errors); ok {
			errors := make([]Validation, 0)

			for _, verr := range v.Errors {
				errors = append(
					errors,
					Validation{
						Field:   ToPtr(verr.Field),
						Message: ToPtr(verr.Error.Error()),
					},
				)
			}

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate execution"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to create execution",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("action", "CreateProjectExecutions"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to create execution"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProjectExecutionResponse(
		a.convertExecution(record),
	))
}

// DeleteProjectExecution implements the v1.ServerInterface.
func (a *API) DeleteProjectExecution(w http.ResponseWriter, r *http.Request, _ ProjectID, _ ExecutionID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectExecutionFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.Delete(
		ctx,
		project,
		record.ID,
	); err != nil {
		slog.Error(
			"Failed to delete execution",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("execution", record.ID),
			slog.String("action", "DeleteProjectExecution"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to delete execution"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully deleted execution"),
		Status:  ToPtr(http.StatusOK),
	})
}

// PurgeProjectExecution implements the v1.ServerInterface.
func (a *API) PurgeProjectExecution(w http.ResponseWriter, r *http.Request, _ ProjectID, _ ExecutionID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectExecutionFromContext(ctx)

	if err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.Purge(
		ctx,
		project,
		record.ID,
	); err != nil {
		slog.Error(
			"Failed to purge execution",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("execution", record.ID),
			slog.String("action", "PurgeProjectExecution"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to purge execution"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	a.RenderNotify(w, r, Notification{
		Message: ToPtr("Successfully purged execution"),
		Status:  ToPtr(http.StatusOK),
	})
}

// OutputProjectExecution implements the v1.ServerInterface.
func (a *API) OutputProjectExecution(w http.ResponseWriter, r *http.Request, _ ProjectID, _ ExecutionID) {
	ctx := r.Context()
	project := a.ProjectFromContext(ctx)
	record := a.ProjectExecutionFromContext(ctx)

	outputs, err := a.storage.WithPrincipal(
		current.GetUser(ctx),
	).Executions.Outputs(
		ctx,
		project,
		record,
	)

	if err != nil {
		slog.Error(
			"Failed to load output",
			slog.Any("error", err),
			slog.String("project", project.ID),
			slog.String("execution", record.ID),
			slog.String("action", "OutputProjectExecution"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to load output"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	payload := make([]Output, len(outputs))
	for id, output := range outputs {
		payload[id] = a.convertOutput(output)
	}

	render.JSON(w, r, ProjectOutputResponse(
		payload,
	))
}

func (a *API) convertExecution(record *model.Execution) Execution {
	result := Execution{
		ID:        ToPtr(record.ID),
		Name:      ToPtr(record.Name),
		Status:    ToPtr(record.Status),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	if record.Template != nil {
		result.TemplateID = ToPtr(record.TemplateID)

		result.Template = ToPtr(
			a.convertTemplate(
				record.Template,
			),
		)
	}

	return result
}

func (a *API) convertOutput(record *model.Output) Output {
	result := Output{
		Content:   ToPtr(record.Content),
		CreatedAt: ToPtr(record.CreatedAt),
	}

	if record.Execution != nil {
		result.ExecutionID = ToPtr(record.ExecutionID)

		result.Execution = ToPtr(
			a.convertExecution(
				record.Execution,
			),
		)
	}

	return result
}

// AllowShowProjectExecution defines a middleware to check permissions.
func (a *API) AllowShowProjectExecution(next http.Handler) http.Handler {
	return a.AllowShowProject(next)
}

// AllowManageProjectExecution defines a middleware to check permissions.
func (a *API) AllowManageProjectExecution(next http.Handler) http.Handler {
	return a.AllowManageProject(next)
}

func listExecutionsSorting(request ListProjectExecutionsParams) (string, string, int64, int64, string) {
	sort, limit, offset, search := toPageParams(
		request.Sort,
		request.Limit,
		request.Offset,
		request.Search,
	)

	order := ""

	if request.Order != nil {
		order = string(FromPtr(request.Order))
	}

	return sort, order, limit, offset, search
}
