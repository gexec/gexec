package v1

import (
	"context"
	"net/http"
)

// ListProjectTasks implements the v1.ServerInterface.
func (a *API) ListProjectTasks(_ context.Context, _ ListProjectTasksRequestObject) (ListProjectTasksResponseObject, error) {
	return ListProjectTasks500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowProjectTask implements the v1.ServerInterface.
func (a *API) ShowProjectTask(_ context.Context, _ ShowProjectTaskRequestObject) (ShowProjectTaskResponseObject, error) {
	return ShowProjectTask500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateProjectTask implements the v1.ServerInterface.
func (a *API) CreateProjectTask(_ context.Context, _ CreateProjectTaskRequestObject) (CreateProjectTaskResponseObject, error) {
	return CreateProjectTask500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectTask implements the v1.ServerInterface.
func (a *API) DeleteProjectTask(_ context.Context, _ DeleteProjectTaskRequestObject) (DeleteProjectTaskResponseObject, error) {
	return DeleteProjectTask500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// StopProjectTask implements the v1.ServerInterface.
func (a *API) StopProjectTask(_ context.Context, _ StopProjectTaskRequestObject) (StopProjectTaskResponseObject, error) {
	return StopProjectTask500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// OutputProjectTask implements the v1.ServerInterface.
func (a *API) OutputProjectTask(_ context.Context, _ OutputProjectTaskRequestObject) (OutputProjectTaskResponseObject, error) {
	return OutputProjectTask500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
