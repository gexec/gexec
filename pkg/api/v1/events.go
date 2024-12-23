package v1

import (
	"context"
	"net/http"
)

// ListEvents implements the v1.ServerInterface.
func (a *API) ListEvents(_ context.Context, _ ListEventsRequestObject) (ListEventsResponseObject, error) {
	return ListEvents500JSONResponse{InternalServerErrorJSONResponse{
		Message: ToPtr("Not implemented"),
		Status:  ToPtr(http.StatusInternalServerError),
	}}, nil
}
