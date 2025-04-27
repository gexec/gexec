package v1

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/gexec/gexec/pkg/middleware/current"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/token"
	"github.com/gexec/gexec/pkg/validate"
	"github.com/go-chi/render"
)

// TokenProfile implements the v1.ServerInterface.
func (a *API) TokenProfile(w http.ResponseWriter, r *http.Request) {
	principal := current.GetUser(
		r.Context(),
	)

	result, err := token.Authed(
		a.config.Token.Secret,
		10*365*24*time.Hour,
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
			slog.String("username", principal.Username),
			slog.String("uid", principal.ID),
			slog.String("action", "TokenProfile"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to generate a token"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, TokenResponse(
		a.convertAuthToken(result),
	))
}

// ShowProfile implements the v1.ServerInterface.
func (a *API) ShowProfile(w http.ResponseWriter, r *http.Request) {
	record := current.GetUser(
		r.Context(),
	)

	render.JSON(w, r, ProfileResponse(
		a.convertProfile(
			record,
		),
	))
}

// UpdateProfile implements the v1.ServerInterface.
func (a *API) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	incoming := current.GetUser(
		r.Context(),
	)

	body := &UpdateProfileBody{}

	if err := json.NewDecoder(r.Body).Decode(body); err != nil {
		slog.Error(
			"Failed to decode request body",
			slog.Any("error", err),
			slog.String("username", incoming.Username),
			slog.String("uid", incoming.ID),
			slog.String("action", "UpdateProfile"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to decode request"),
			Status:  ToPtr(http.StatusBadRequest),
		})

		return
	}

	if body.Username != nil {
		incoming.Username = FromPtr(body.Username)
	}

	if body.Password != nil {
		incoming.Password = FromPtr(body.Password)
	}

	if body.Email != nil {
		incoming.Email = FromPtr(body.Email)
	}

	if body.Fullname != nil {
		incoming.Fullname = FromPtr(body.Fullname)
	}

	record, err := a.storage.Users.Update(
		r.Context(),
		incoming,
	)

	if err != nil {
		if v, ok := err.(validate.Errors); ok {
			errors := make([]Validation, 0)

			for _, verr := range v.Errors {
				errors = append(
					errors,
					Validation{
						Field:   ToPtr(verr.Field),
						Message: ToPtr(verr.Error.Error()),
					},
				)
			}

			a.RenderNotify(w, r, Notification{
				Message: ToPtr("Failed to validate profile"),
				Status:  ToPtr(http.StatusUnprocessableEntity),
				Errors:  ToPtr(errors),
			})

			return
		}

		slog.Error(
			"Failed to update profile",
			slog.Any("error", err),
			slog.String("username", record.Username),
			slog.String("uid", record.ID),
			slog.String("action", "UpdateProfile"),
		)

		a.RenderNotify(w, r, Notification{
			Message: ToPtr("Failed to update profile"),
			Status:  ToPtr(http.StatusInternalServerError),
		})

		return
	}

	render.JSON(w, r, ProfileResponse(
		a.convertProfile(
			record,
		),
	))
}

func (a *API) convertProfile(record *model.User) Profile {
	result := Profile{
		ID:        ToPtr(record.ID),
		Username:  ToPtr(record.Username),
		Email:     ToPtr(record.Email),
		Fullname:  ToPtr(record.Fullname),
		Profile:   ToPtr(gravatarFor(record.Email)),
		Active:    ToPtr(record.Active),
		Admin:     ToPtr(record.Admin),
		CreatedAt: ToPtr(record.CreatedAt),
		UpdatedAt: ToPtr(record.UpdatedAt),
	}

	if len(record.Auths) > 0 {
		auths := make([]UserAuth, 0)

		for _, auth := range record.Auths {
			auths = append(
				auths,
				a.convertUserAuth(auth),
			)
		}

		result.Auths = ToPtr(auths)
	}

	if len(record.Groups) > 0 {
		groups := make([]UserGroup, 0)

		for _, group := range record.Groups {
			groups = append(
				groups,
				a.convertUserGroup(group),
			)
		}

		result.Groups = ToPtr(groups)
	}

	if len(record.Projects) > 0 {
		projects := make([]UserProject, 0)

		for _, project := range record.Projects {
			projects = append(
				projects,
				a.convertUserProject(project),
			)
		}

		result.Projects = ToPtr(projects)
	}

	return result
}
