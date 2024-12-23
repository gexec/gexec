package v1

import (
	"context"
	"net/http"
)

// TokenProfile implements the v1.ServerInterface.
func (a *API) TokenProfile(_ context.Context, _ TokenProfileRequestObject) (TokenProfileResponseObject, error) {
	return TokenProfile500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ShowProfile implements the v1.ServerInterface.
func (a *API) ShowProfile(_ context.Context, _ ShowProfileRequestObject) (ShowProfileResponseObject, error) {
	return ShowProfile500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// UpdateProfile implements the v1.ServerInterface.
func (a *API) UpdateProfile(_ context.Context, _ UpdateProfileRequestObject) (UpdateProfileResponseObject, error) {
	return UpdateProfile500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
