package v1

import (
	"context"
	"net/http"
)

// Websockets implements the v1.ServerInterface.
func (a *API) Websockets(_ context.Context, _ WebsocketsRequestObject) (WebsocketsResponseObject, error) {
	return Websockets500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
