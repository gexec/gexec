package v1

import (
	"context"
	"net/http"
)

// ListProjectTemplates implements the v1.ServerInterface.
func (a *API) ListProjectTemplates(_ context.Context, _ ListProjectTemplatesRequestObject) (ListProjectTemplatesResponseObject, error) {
	return ListProjectTemplates500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowProjectTemplate implements the v1.ServerInterface.
func (a *API) ShowProjectTemplate(_ context.Context, _ ShowProjectTemplateRequestObject) (ShowProjectTemplateResponseObject, error) {
	return ShowProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateProjectTemplate implements the v1.ServerInterface.
func (a *API) CreateProjectTemplate(_ context.Context, _ CreateProjectTemplateRequestObject) (CreateProjectTemplateResponseObject, error) {
	return CreateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProjectTemplate implements the v1.ServerInterface.
func (a *API) UpdateProjectTemplate(_ context.Context, _ UpdateProjectTemplateRequestObject) (UpdateProjectTemplateResponseObject, error) {
	return UpdateProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectTemplate implements the v1.ServerInterface.
func (a *API) DeleteProjectTemplate(_ context.Context, _ DeleteProjectTemplateRequestObject) (DeleteProjectTemplateResponseObject, error) {
	return DeleteProjectTemplate500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
