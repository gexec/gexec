package v1

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/gexec/gexec/pkg/authn"
	"github.com/gexec/gexec/pkg/config"
	"github.com/gexec/gexec/pkg/metrics"
	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/store"
	"github.com/gexec/gexec/pkg/token"
	"github.com/gexec/gexec/pkg/upload"
	"github.com/go-chi/render"
)

//go:generate go tool github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen --config=config.yaml ../../../openapi/v1.yaml

var (
	_ ServerInterface = (*API)(nil)

	// ErrUnsupportedImageFormat defines the error for unsupported image formats.
	ErrUnsupportedImageFormat = fmt.Errorf("unsupported avatar file format")
)

func init() {
	openapi3filter.RegisterBodyDecoder("image/jpeg", openapi3filter.FileBodyDecoder)
	openapi3filter.RegisterBodyDecoder("image/png", openapi3filter.FileBodyDecoder)
}

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

// RenderNotify is a helper to set a correct status for notifications.
func (a *API) RenderNotify(w http.ResponseWriter, r *http.Request, notify Notification) {
	render.Status(
		r,
		FromPtr(notify.Status),
	)

	render.JSON(
		w,
		r,
		notify,
	)
}

// AllowAdminAccessOnly defines a middleware to check permissions.
func (a *API) AllowAdminAccessOnly(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		principal := current.GetUser(ctx)

		if principal == nil {
			render.JSON(w, r, Notification{
				Message: ToPtr("Only admins can access this resource"),
				Status:  ToPtr(http.StatusForbidden),
			})

			return
		}

		if principal.Admin {
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		render.JSON(w, r, Notification{
			Message: ToPtr("Only admins can access this resource"),
			Status:  ToPtr(http.StatusForbidden),
		})
	})
}

// Authentication provides the authentication for the OpenAPI filter.
func (a *API) Authentication(ctx context.Context, input *openapi3filter.AuthenticationInput) error {
	authenticating := &model.User{}
	scheme := input.SecuritySchemeName
	operation := input.RequestValidationInput.Route.Operation.OperationID

	logger := slog.With(
		slog.String("scheme", scheme),
		slog.String("operation", operation),
	)

	switch scheme {
	case "Header":
		header := input.RequestValidationInput.Request.Header.Get(
			input.SecurityScheme.Name,
		)

		if header == "" {
			return fmt.Errorf("missing authorization header")
		}

		t, err := token.Verify(
			a.config.Token.Secret,
			strings.TrimSpace(
				header,
			),
		)

		if err != nil {
			return fmt.Errorf("failed to parse auth token")
		}

		user, err := a.storage.Auth.ByID(
			ctx,
			t.Ident,
		)

		if err != nil {
			logger.Error(
				"Failed to find user",
				slog.Any("error", err),
				slog.String("user", t.Ident),
			)

			return fmt.Errorf("failed to find user")
		}

		logger.Debug(
			"Authentication",
			slog.String("user", t.Login),
		)

		authenticating = user

	case "Bearer":
		header := input.RequestValidationInput.Request.Header.Get(
			"Authorization",
		)

		if header == "" {
			return fmt.Errorf("missing authorization bearer")
		}

		t, err := token.Verify(
			a.config.Token.Secret,
			strings.TrimSpace(
				strings.Replace(
					header,
					"Bearer",
					"",
					1,
				),
			),
		)

		if err != nil {
			return fmt.Errorf("failed to parse auth token")
		}

		user, err := a.storage.Auth.ByID(
			ctx,
			t.Ident,
		)

		if err != nil {
			logger.Error(
				"Failed to find user",
				slog.Any("error", err),
				slog.String("user", t.Ident),
			)

			return fmt.Errorf("failed to find user")
		}

		logger.Debug(
			"Authentication",
			slog.String("user", t.Login),
		)

		authenticating = user

	case "Basic":
		username, password, ok := input.RequestValidationInput.Request.BasicAuth()

		if !ok {
			return fmt.Errorf("missing credentials")
		}

		user, err := a.storage.Auth.ByCreds(
			ctx,
			username,
			password,
		)

		if err != nil {
			logger.Error(
				"Wrong credentials",
				slog.Any("error", err),
				slog.String("user", username),
			)

			return fmt.Errorf("wrong credentials")
		}

		logger.Debug(
			"Authentication",
			slog.String("user", username),
		)

		authenticating = user

	default:
		return fmt.Errorf("unknown security scheme: %s", scheme)
	}

	logger.Debug(
		"Authenticated",
		slog.String("user", authenticating.Username),
	)

	current.SetUser(
		input.RequestValidationInput.Request.Context(),
		authenticating,
	)

	return nil
}
