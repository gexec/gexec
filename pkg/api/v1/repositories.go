package v1

import (
	"context"
	"net/http"
)

// ListProjectRepositories implements the v1.ServerInterface.
func (a *API) ListProjectRepositories(_ context.Context, _ ListProjectRepositoriesRequestObject) (ListProjectRepositoriesResponseObject, error) {
	return ListProjectRepositories500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowProjectRepository implements the v1.ServerInterface.
func (a *API) ShowProjectRepository(_ context.Context, _ ShowProjectRepositoryRequestObject) (ShowProjectRepositoryResponseObject, error) {
	return ShowProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateProjectRepository implements the v1.ServerInterface.
func (a *API) CreateProjectRepository(_ context.Context, _ CreateProjectRepositoryRequestObject) (CreateProjectRepositoryResponseObject, error) {
	return CreateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProjectRepository implements the v1.ServerInterface.
func (a *API) UpdateProjectRepository(_ context.Context, _ UpdateProjectRepositoryRequestObject) (UpdateProjectRepositoryResponseObject, error) {
	return UpdateProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectRepository implements the v1.ServerInterface.
func (a *API) DeleteProjectRepository(_ context.Context, _ DeleteProjectRepositoryRequestObject) (DeleteProjectRepositoryResponseObject, error) {
	return DeleteProjectRepository500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
