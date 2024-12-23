package v1

import (
	"context"
	"net/http"
)

// CallbackProvider implements the v1.ServerInterface.
func (a *API) CallbackProvider(_ context.Context, _ CallbackProviderRequestObject) (CallbackProviderResponseObject, error) {
	return CallbackProvider500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// ListProvider implements the v1.ServerInterface.
func (a *API) ListProvider(_ context.Context, _ ListProviderRequestObject) (ListProviderResponseObject, error) {
	return ListProvider500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// LoginAuth implements the v1.ServerInterface.
func (a *API) LoginAuth(_ context.Context, _ LoginAuthRequestObject) (LoginAuthResponseObject, error) {
	return LoginAuth500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// RefreshAuth implements the v1.ServerInterface.
func (a *API) RefreshAuth(_ context.Context, _ RefreshAuthRequestObject) (RefreshAuthResponseObject, error) {
	return RefreshAuth500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}

// VerifyAuth implements the v1.ServerInterface.
func (a *API) VerifyAuth(_ context.Context, _ VerifyAuthRequestObject) (VerifyAuthResponseObject, error) {
	return VerifyAuth500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
