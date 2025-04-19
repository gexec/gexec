package v1

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"path"
	"strings"

	"github.com/gexec/gexec/pkg/authn"
	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/secret"
	"github.com/gexec/gexec/pkg/store"
	"github.com/gexec/gexec/pkg/templates"
	"github.com/gexec/gexec/pkg/token"
	"github.com/go-chi/render"
	"github.com/gobwas/glob"
	"golang.org/x/oauth2"
)

// RequestProvider implements the v1.ServerInterface.
func (a *API) RequestProvider(w http.ResponseWriter, r *http.Request, providerParam AuthProviderParam) {
	provider, ok := a.identity.Providers[providerParam]

	if !ok {
		slog.Error(
			"Failed to detect provider",
			slog.String("provider", providerParam),
		)

		render.Status(r, http.StatusPreconditionFailed)
		render.HTML(w, r, templates.String(
			a.config,
			"error.tmpl",
			struct {
				Error  string
				Status int
			}{
				Error:  "Failed to detect provider",
				Status: http.StatusPreconditionFailed,
			},
		))

		return
	}

	w.Header().Set(
		"Location",
		provider.OAuth2.AuthCodeURL(
			base64.URLEncoding.EncodeToString(
				secret.Bytes(64),
			),
			oauth2.AccessTypeOffline,
			oauth2.S256ChallengeOption(
				base64.RawURLEncoding.EncodeToString(
					[]byte(provider.Config.Verifier),
				),
			),
		),
	)

	w.Header().Set(
		"Content-Type",
		"text/html",
	)

	w.WriteHeader(
		http.StatusPermanentRedirect,
	)
}

// CallbackProvider implements the v1.ServerInterface.
func (a *API) CallbackProvider(w http.ResponseWriter, r *http.Request, providerParam AuthProviderParam, params CallbackProviderParams) {
	provider, ok := a.identity.Providers[providerParam]

	if !ok {
		slog.Error(
			"Failed to detect provider",
			slog.String("provider", providerParam),
		)

		render.Status(r, http.StatusPreconditionFailed)
		render.HTML(w, r, templates.String(
			a.config,
			"error.tmpl",
			struct {
				Error  string
				Status int
			}{
				Error:  "Failed to detect provider",
				Status: http.StatusPreconditionFailed,
			},
		))

		return
	}

	exchange, err := provider.OAuth2.Exchange(
		r.Context(),
		FromPtr(params.Code),
		oauth2.SetAuthURLParam(
			"code_verifier",
			base64.RawURLEncoding.EncodeToString(
				[]byte(provider.Config.Verifier),
			),
		),
	)

	if err != nil {
		slog.Error(
			"Failed to exchange token",
			slog.Any("error", err),
			slog.String("provider", providerParam),
		)

		render.Status(r, http.StatusPreconditionFailed)
		render.HTML(w, r, templates.String(
			a.config,
			"error.tmpl",
			struct {
				Error  string
				Status int
			}{
				Error:  "Failed to exchange token",
				Status: http.StatusPreconditionFailed,
			},
		))

		return
	}

	external, err := provider.Claims(
		r.Context(),
		exchange,
	)

	if err != nil {
		slog.Error(
			"Failed to parse claims",
			slog.Any("error", err),
			slog.String("provider", providerParam),
		)

		render.Status(r, http.StatusPreconditionFailed)
		render.HTML(w, r, templates.String(
			a.config,
			"error.tmpl",
			struct {
				Error  string
				Status int
			}{
				Error:  "Failed to parse claims",
				Status: http.StatusPreconditionFailed,
			},
		))

		return
	}

	user, err := a.storage.Auth.External(
		r.Context(),
		provider.Config.Name,
		external.Ident,
		external.Login,
		external.Email,
		external.Name,
		detectAdminFor(provider, external),
	)

	if err != nil {
		slog.Error(
			"Failed to create user",
			slog.Any("error", err),
			slog.String("provider", providerParam),
			slog.String("username", external.Login),
		)

		render.Status(r, http.StatusPreconditionFailed)
		render.HTML(w, r, templates.String(
			a.config,
			"error.tmpl",
			struct {
				Error  string
				Status int
			}{
				Error:  "Failed to create user",
				Status: http.StatusPreconditionFailed,
			},
		))

		return
	}

	slog.Debug(
		"Authenticated",
		slog.String("provider", providerParam),
		slog.String("username", user.Username),
		slog.String("uid", user.ID),
		slog.String("email", user.Email),
		slog.String("external", external.Ident),
	)

	redirect, err := a.storage.Users.CreateRedirectToken(
		r.Context(),
		user.ID,
	)

	if err != nil {
		slog.Error(
			"Failed to generate a token",
			slog.Any("error", err),
			slog.String("username", user.Username),
			slog.String("uid", user.ID),
		)

		render.Status(r, http.StatusPreconditionFailed)
		render.HTML(w, r, templates.String(
			a.config,
			"error.tmpl",
			struct {
				Error  string
				Status int
			}{
				Error:  "Failed to generate token",
				Status: http.StatusPreconditionFailed,
			},
		))

		return
	}

	slog.Info(
		"Successfully generated token",
		slog.String("username", user.Username),
		slog.String("uid", user.ID),
		slog.String("token", redirect.Token),
	)

	w.Header().Set(
		"Location",
		strings.Join([]string{
			a.config.Server.Host,
			path.Join(
				a.config.Server.Root,
				"auth",
				"callback",
				redirect.Token,
			),
		}, ""),
	)

	w.Header().Set(
		"Content-Type",
		"text/html",
	)

	w.WriteHeader(
		http.StatusPermanentRedirect,
	)
}

// ListProviders implements the v1.ServerInterface.
func (a *API) ListProviders(w http.ResponseWriter, r *http.Request) {
	records := make([]Provider, 0)

	for _, provider := range a.identity.Providers {
		records = append(
			records,
			Provider{
				Name:    ToPtr(provider.Config.Name),
				Driver:  ToPtr(provider.Config.Driver),
				Display: ToPtr(provider.Config.Display),
				Icon:    ToPtr(provider.Config.Icon),
			},
		)
	}

	render.JSON(w, r,
		ProvidersResponse{
			Total:     int64(len(records)),
			Providers: records,
		},
	)
}

// RedirectAuth implements the v1.ServerInterface.
func (a *API) RedirectAuth(w http.ResponseWriter, r *http.Request) {
	body := &RedirectAuthBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	redirect, err := a.storage.Users.ShowRedirectToken(
		r.Context(),
		body.Token,
	)

	if err != nil {
		if errors.Is(err, store.ErrTokenNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate token"),
				Status:  ToPtr(http.StatusUnauthorized),
			})

			return
		}

		slog.Error(
			"Failed to validate token",
			slog.Any("error", err),
			slog.String("token", body.Token),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to validate token"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	user, err := a.storage.Auth.ByID(
		r.Context(),
		redirect.UserID,
	)

	if err != nil {
		slog.Error(
			"Failed to authenticate",
			slog.Any("error", err),
			slog.String("token", body.Token),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to authenticate user"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	result, err := token.Authed(
		a.config.Token.Secret,
		a.config.Token.Expire,
		user.ID,
		user.Username,
		user.Email,
		user.Fullname,
		user.Admin,
	)

	if err != nil {
		slog.Error(
			"Failed to generate a token",
			slog.Any("error", err),
			slog.String("token", body.Token),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to generate a token"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	if err := a.storage.Users.DeleteRedirectToken(
		r.Context(),
		redirect.Token,
	); err != nil {
		slog.Error(
			"Failed to cleanup redirect",
			slog.Any("error", err),
			slog.String("token", body.Token),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to cleanup redirect"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r,
		a.convertAuthToken(result),
	)
}

// LoginAuth implements the v1.ServerInterface.
func (a *API) LoginAuth(w http.ResponseWriter, r *http.Request) {
	body := &LoginAuthBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	user, err := a.storage.Auth.ByCreds(
		r.Context(),
		body.Username,
		body.Password,
	)

	if err != nil {
		if errors.Is(err, store.ErrUserNotFound) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Wrong username or password"),
				Status:  ToPtr(http.StatusUnauthorized),
			})

			return
		}

		if errors.Is(err, store.ErrWrongCredentials) {
			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Wrong username or password"),
				Status:  ToPtr(http.StatusUnauthorized),
			})

			return
		}

		slog.Error(
			"Failed to authenticate",
			slog.Any("error", err),
			slog.String("username", body.Username),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to authenticate user"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	result, err := token.Authed(
		a.config.Token.Secret,
		a.config.Token.Expire,
		user.ID,
		user.Username,
		user.Email,
		user.Fullname,
		user.Admin,
	)

	if err != nil {
		slog.Error(
			"Failed to generate a token",
			slog.Any("error", err),
			slog.String("username", body.Username),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to generate a token"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r,
		a.convertAuthToken(result),
	)
}

// RefreshAuth implements the v1.ServerInterface.
func (a *API) RefreshAuth(w http.ResponseWriter, r *http.Request) {
	principal := current.GetUser(
		r.Context(),
	)

	result, err := token.Authed(
		a.config.Token.Secret,
		a.config.Token.Expire,
		principal.ID,
		principal.Username,
		principal.Email,
		principal.Fullname,
		principal.Admin,
	)

	if err != nil {
		slog.Error(
			"Failed to generate a token",
			slog.Any("error", err),
			slog.String("action", "RefreshAuth"),
			slog.String("username", principal.Username),
			slog.String("uid", principal.ID),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to generate a token"),
			Status:  ToPtr(http.StatusUnauthorized),
		})

		return
	}

	render.JSON(w, r,
		a.convertAuthToken(result),
	)
}

// VerifyAuth implements the v1.ServerInterface.
func (a *API) VerifyAuth(w http.ResponseWriter, r *http.Request) {
	principal := current.GetUser(
		r.Context(),
	)

	render.JSON(w, r,
		a.convertAuthVerify(principal),
	)
}

func (a *API) convertAuthToken(record string) AuthToken {
	return AuthToken{
		Token: ToPtr(record),
	}
}

func (a *API) convertAuthVerify(record *model.User) AuthVerify {
	return AuthVerify{
		Username:  ToPtr(record.Username),
		CreatedAt: ToPtr(record.CreatedAt),
	}
}

func detectAdminFor(provider *authn.Provider, external *authn.User) bool {
	for _, user := range provider.Config.Admins.Users {
		if user == external.Login {
			return true
		}
	}

	for _, email := range provider.Config.Admins.Emails {
		g, err := glob.Compile(email)

		if err != nil {
			slog.Error(
				"Failed to compile email glob",
				slog.String("provider", provider.Config.Name),
				slog.String("glob", email),
			)

			continue
		}

		if g.Match(external.Email) {
			return true
		}
	}

	if provider.Config.Mappings.Role != "" {
		for _, checkRole := range provider.Config.Admins.Roles {
			for _, assignedRole := range external.Roles {
				if checkRole == assignedRole {
					return true
				}
			}
		}
	}

	return false
}
