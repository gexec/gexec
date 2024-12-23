package v1

import (
	"context"
	"net/http"
)

// ListRunners implements the v1.ServerInterface.
func (a *API) ListRunners(_ context.Context, _ ListRunnersRequestObject) (ListRunnersResponseObject, error) {
	return ListRunners500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowRunner implements the v1.ServerInterface.
func (a *API) ShowRunner(_ context.Context, _ ShowRunnerRequestObject) (ShowRunnerResponseObject, error) {
	return ShowRunner500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateRunner implements the v1.ServerInterface.
func (a *API) CreateRunner(_ context.Context, _ CreateRunnerRequestObject) (CreateRunnerResponseObject, error) {
	return CreateRunner500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateRunner implements the v1.ServerInterface.
func (a *API) UpdateRunner(_ context.Context, _ UpdateRunnerRequestObject) (UpdateRunnerResponseObject, error) {
	return UpdateRunner500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteRunner implements the v1.ServerInterface.
func (a *API) DeleteRunner(_ context.Context, _ DeleteRunnerRequestObject) (DeleteRunnerResponseObject, error) {
	return DeleteRunner500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
