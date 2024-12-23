package v1

import (
	"context"
	"net/http"
)

// ListProjectSchedules implements the v1.ServerInterface.
func (a *API) ListProjectSchedules(_ context.Context, _ ListProjectSchedulesRequestObject) (ListProjectSchedulesResponseObject, error) {
	return ListProjectSchedules500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowProjectSchedule implements the v1.ServerInterface.
func (a *API) ShowProjectSchedule(_ context.Context, _ ShowProjectScheduleRequestObject) (ShowProjectScheduleResponseObject, error) {
	return ShowProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateProjectSchedule implements the v1.ServerInterface.
func (a *API) CreateProjectSchedule(_ context.Context, _ CreateProjectScheduleRequestObject) (CreateProjectScheduleResponseObject, error) {
	return CreateProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProjectSchedule implements the v1.ServerInterface.
func (a *API) UpdateProjectSchedule(_ context.Context, _ UpdateProjectScheduleRequestObject) (UpdateProjectScheduleResponseObject, error) {
	return UpdateProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectSchedule implements the v1.ServerInterface.
func (a *API) DeleteProjectSchedule(_ context.Context, _ DeleteProjectScheduleRequestObject) (DeleteProjectScheduleResponseObject, error) {
	return DeleteProjectSchedule500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
