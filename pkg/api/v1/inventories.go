package v1

import (
	"context"
	"net/http"
)

// ListProjectInventories implements the v1.ServerInterface.
func (a *API) ListProjectInventories(_ context.Context, _ ListProjectInventoriesRequestObject) (ListProjectInventoriesResponseObject, error) {
	return ListProjectInventories500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowProjectInventory implements the v1.ServerInterface.
func (a *API) ShowProjectInventory(_ context.Context, _ ShowProjectInventoryRequestObject) (ShowProjectInventoryResponseObject, error) {
	return ShowProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// CreateProjectInventory implements the v1.ServerInterface.
func (a *API) CreateProjectInventory(_ context.Context, _ CreateProjectInventoryRequestObject) (CreateProjectInventoryResponseObject, error) {
	return CreateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProjectInventory implements the v1.ServerInterface.
func (a *API) UpdateProjectInventory(_ context.Context, _ UpdateProjectInventoryRequestObject) (UpdateProjectInventoryResponseObject, error) {
	return UpdateProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// DeleteProjectInventory implements the v1.ServerInterface.
func (a *API) DeleteProjectInventory(_ context.Context, _ DeleteProjectInventoryRequestObject) (DeleteProjectInventoryResponseObject, error) {
	return DeleteProjectInventory500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
