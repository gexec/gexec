package v1

import (
	"context"
	"net/http"
)

// ListProjectEnvironments implements the v1.ServerInterface.
func (a *API) ListProjectEnvironments(_ context.Context, _ ListProjectEnvironmentsRequestObject) (ListProjectEnvironmentsResponseObject, error) {
	return ListProjectEnvironments500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowProjectEnvironment implements the v1.ServerInterface.
func (a *API) ShowProjectEnvironment(_ context.Context, _ ShowProjectEnvironmentRequestObject) (ShowProjectEnvironmentResponseObject, error) {
	return ShowProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateProjectEnvironment implements the v1.ServerInterface.
func (a *API) CreateProjectEnvironment(_ context.Context, _ CreateProjectEnvironmentRequestObject) (CreateProjectEnvironmentResponseObject, error) {
	return CreateProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProjectEnvironment implements the v1.ServerInterface.
func (a *API) UpdateProjectEnvironment(_ context.Context, _ UpdateProjectEnvironmentRequestObject) (UpdateProjectEnvironmentResponseObject, error) {
	return UpdateProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectEnvironment implements the v1.ServerInterface.
func (a *API) DeleteProjectEnvironment(_ context.Context, _ DeleteProjectEnvironmentRequestObject) (DeleteProjectEnvironmentResponseObject, error) {
	return DeleteProjectEnvironment500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
