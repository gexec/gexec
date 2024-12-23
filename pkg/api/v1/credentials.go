package v1

import (
	"context"
	"net/http"
)

// ListProjectCredentials implements the v1.ServerInterface.
func (a *API) ListProjectCredentials(_ context.Context, _ ListProjectCredentialsRequestObject) (ListProjectCredentialsResponseObject, error) {
	return ListProjectCredentials500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowProjectCredential implements the v1.ServerInterface.
func (a *API) ShowProjectCredential(_ context.Context, _ ShowProjectCredentialRequestObject) (ShowProjectCredentialResponseObject, error) {
	return ShowProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateProjectCredential implements the v1.ServerInterface.
func (a *API) CreateProjectCredential(_ context.Context, _ CreateProjectCredentialRequestObject) (CreateProjectCredentialResponseObject, error) {
	return CreateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProjectCredential implements the v1.ServerInterface.
func (a *API) UpdateProjectCredential(_ context.Context, _ UpdateProjectCredentialRequestObject) (UpdateProjectCredentialResponseObject, error) {
	return UpdateProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectCredential implements the v1.ServerInterface.
func (a *API) DeleteProjectCredential(_ context.Context, _ DeleteProjectCredentialRequestObject) (DeleteProjectCredentialResponseObject, error) {
	return DeleteProjectCredential500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
