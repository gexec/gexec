package v1

import (
	"context"

	"github.com/genexec/genexec/pkg/authn"
	"github.com/genexec/genexec/pkg/config"
	"github.com/genexec/genexec/pkg/metrics"
	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/store"
	"github.com/genexec/genexec/pkg/upload"
)

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../../../openapi/v1.yaml

var (
	_ StrictServerInterface = (*API)(nil)
)

// New creates a new API that adds the handler implementations.
func New(
	cfg *config.Config,
	registry *metrics.Metrics,
	identity *authn.Authn,
	uploads upload.Upload,
	storage *store.Store,
) *API {
	return &API{
		config:   cfg,
		registry: registry,
		identity: identity,
		uploads:  uploads,
		storage:  storage,
	}
}

// API provides the http.Handler for the OpenAPI implementation.
type API struct {
	config   *config.Config
	registry *metrics.Metrics
	identity *authn.Authn
	uploads  upload.Upload
	storage  *store.Store
}

func (a *API) permitAdmin(_ context.Context, principal *model.User) bool {
	if principal == nil {
		return false
	}

	if principal.Admin {
		return true
	}

	return false
}
