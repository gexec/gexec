package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/secret"
	"github.com/gexec/gexec/pkg/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/uptrace/bun"
)

// Users provides all database operations related to users.
type Users struct {
	client *Store
}

// List implements the listing of all users.
func (s *Users) List(ctx context.Context, params model.ListParams) ([]*model.User, int64, error) {
	records := make([]*model.User, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Auths")

	if val, ok := s.validSort(params.Sort); ok {
		q = q.Order(strings.Join(
			[]string{
				val,
				sortOrder(params.Order),
			},
			" ",
		))
	}

	if params.Search != "" {
		q = s.client.SearchQuery(q, params.Search)
	}

	counter, err := q.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if params.Limit > 0 {
		q = q.Limit(int(params.Limit))
	}

	if params.Offset > 0 {
		q = q.Offset(int(params.Offset))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, int64(counter), err
	}

	return records, int64(counter), nil
}

// Show implements the details for a specific user.
func (s *Users) Show(ctx context.Context, name string) (*model.User, error) {
	record := &model.User{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Relation("Auths").
		Where("id = ? OR username = ?", name, name).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrUserNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new user.
func (s *Users) Create(ctx context.Context, record *model.User) (*model.User, error) {
	if err := s.validate(ctx, record, false); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      record.ID,
				ObjectDisplay: record.Username,
				ObjectType:    model.EventTypeUser,
				Action:        model.EventActionCreate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, record.ID)
}

// Update implements the update of an existing user.
func (s *Users) Update(ctx context.Context, record *model.User) (*model.User, error) {
	if err := s.validate(ctx, record, true); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewUpdate().
		Model(record).
		Where("id = ?", record.ID).
		Exec(ctx); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      record.ID,
				ObjectDisplay: record.Username,
				ObjectType:    model.EventTypeUser,
				Action:        model.EventActionUpdate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, record.ID)
}

// Delete implements the deletion of an user.
func (s *Users) Delete(ctx context.Context, name string) error {
	record, err := s.Show(ctx, name)

	if err != nil {
		return err
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.User)(nil)).
		Where("id = ?", record.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      record.ID,
				ObjectDisplay: record.Username,
				ObjectType:    model.EventTypeUser,
				Action:        model.EventActionDelete,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ShowRedirectToken implements the details for a specific redirect token.
func (s *Users) ShowRedirectToken(ctx context.Context, token string) (*model.UserToken, error) {
	record := &model.UserToken{}
	expired := time.Now().UTC().Add(-5 * time.Minute)

	if err := s.client.handle.NewSelect().
		Model(record).
		Where("token = ? AND kind = ? AND created_at > ?", token, model.UserTokenKindRedirect, expired).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrTokenNotFound
		}

		return record, err
	}

	return record, nil
}

// CreateRedirectToken implements the create of a new redirect token.
func (s *Users) CreateRedirectToken(ctx context.Context, username string) (*model.UserToken, error) {
	user, err := s.Show(ctx, username)

	if err != nil {
		return nil, err
	}

	record := &model.UserToken{
		UserID: user.ID,
		Kind:   model.UserTokenKindRedirect,
		Token:  secret.Generate(32),
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return nil, err
	}

	return record, nil
}

// DeleteRedirectToken implements the deletion of a redirect token.
func (s *Users) DeleteRedirectToken(ctx context.Context, token string) error {
	if _, err := s.client.handle.NewDelete().
		Model((*model.UserToken)(nil)).
		Where("token = ? AND kind = ?", token, model.UserTokenKindRedirect).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// CleanupRedirectTokens implements the cleanup of expired redirect tokens.
func (s *Users) CleanupRedirectTokens(ctx context.Context) error {
	expired := time.Now().UTC().Add(-5 * time.Minute)

	if _, err := s.client.handle.NewDelete().
		Model((*model.UserToken)(nil)).
		Where("kind = ? AND created_at < ?", model.UserTokenKindRedirect, expired).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ListGroups implements the listing of all groups for an user.
func (s *Users) ListGroups(ctx context.Context, params model.UserGroupParams) ([]*model.UserGroup, int64, error) {
	records := make([]*model.UserGroup, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("User").
		Relation("Group").
		Where("user_id = ?", params.UserID)

	if val, ok := s.validGroupSort(params.Sort); ok {
		q = q.Order(strings.Join(
			[]string{
				val,
				sortOrder(params.Order),
			},
			" ",
		))
	}

	if params.Search != "" {
		q = s.client.SearchQuery(q, params.Search)
	}

	counter, err := q.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if params.Limit > 0 {
		q = q.Limit(int(params.Limit))
	}

	if params.Offset > 0 {
		q = q.Offset(int(params.Offset))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, int64(counter), err
	}

	return records, int64(counter), nil
}

// AttachGroup implements the attachment of an user to a group.
func (s *Users) AttachGroup(ctx context.Context, params model.UserGroupParams) error {
	user, err := s.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	group, err := s.client.Groups.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	assigned, err := s.isGroupAssigned(ctx, user.ID, group.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.UserGroup{
		UserID:  user.ID,
		GroupID: group.ID,
		Perm:    params.Perm,
	}

	if err := s.validatePerm(record.Perm); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      user.ID,
				ObjectDisplay: user.Username,
				ObjectType:    model.EventTypeUserGroup,
				Action:        model.EventActionCreate,
				Attrs: map[string]interface{}{
					"group_id":      group.ID,
					"group_display": group.Name,
					"perm":          params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// PermitGroup implements the permission update for a group on an user.
func (s *Users) PermitGroup(ctx context.Context, params model.UserGroupParams) error {
	user, err := s.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	group, err := s.client.Groups.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	unassigned, err := s.isGroupUnassigned(ctx, user.ID, group.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewUpdate().
		Model((*model.UserGroup)(nil)).
		Set("perm = ?", params.Perm).
		Where("user_id = ? AND group_id = ?", user.ID, group.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      user.ID,
				ObjectDisplay: user.Username,
				ObjectType:    model.EventTypeUserGroup,
				Action:        model.EventActionUpdate,
				Attrs: map[string]interface{}{
					"group_id":      group.ID,
					"group_display": group.Name,
					"perm":          params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropGroup implements the removal of an user from a group.
func (s *Users) DropGroup(ctx context.Context, params model.UserGroupParams) error {
	user, err := s.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	group, err := s.client.Groups.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	unassigned, err := s.isGroupUnassigned(ctx, user.ID, group.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.UserGroup)(nil)).
		Where("user_id = ? AND group_id = ?", user.ID, group.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      user.ID,
				ObjectDisplay: user.Username,
				ObjectType:    model.EventTypeUserGroup,
				Action:        model.EventActionDelete,
				Attrs: map[string]interface{}{
					"group_id":      group.ID,
					"group_display": group.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Users) isGroupAssigned(ctx context.Context, userID, groupID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserGroup)(nil)).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Users) isGroupUnassigned(ctx context.Context, userID, groupID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserGroup)(nil)).
		Where("user_id = ? AND group_id = ?", userID, groupID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

// ListProjects implements the listing of all projects for an user.
func (s *Users) ListProjects(ctx context.Context, params model.UserProjectParams) ([]*model.UserProject, int64, error) {
	records := make([]*model.UserProject, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("User").
		Relation("Project").
		Where("user_id = ?", params.UserID)

	if val, ok := s.validProjectSort(params.Sort); ok {
		q = q.Order(strings.Join(
			[]string{
				val,
				sortOrder(params.Order),
			},
			" ",
		))
	}

	if params.Search != "" {
		q = s.client.SearchQuery(q, params.Search)
	}

	counter, err := q.Count(ctx)

	if err != nil {
		return nil, 0, err
	}

	if params.Limit > 0 {
		q = q.Limit(int(params.Limit))
	}

	if params.Offset > 0 {
		q = q.Offset(int(params.Offset))
	}

	if err := q.Scan(ctx); err != nil {
		return nil, int64(counter), err
	}

	return records, int64(counter), nil
}

// AttachProject implements the attachment of an user to a project.
func (s *Users) AttachProject(ctx context.Context, params model.UserProjectParams) error {
	user, err := s.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	assigned, err := s.isProjectAssigned(ctx, user.ID, project.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.UserProject{
		UserID:    user.ID,
		ProjectID: project.ID,
		Perm:      params.Perm,
	}

	if err := s.validatePerm(record.Perm); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       user.ID,
				ObjectDisplay:  user.Username,
				ObjectType:     model.EventTypeUserProject,
				Action:         model.EventActionCreate,
				Attrs: map[string]interface{}{
					"project_id":      project.ID,
					"project_display": project.Name,
					"perm":            params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// PermitProject implements the permission update for a project on an user.
func (s *Users) PermitProject(ctx context.Context, params model.UserProjectParams) error {
	user, err := s.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	unassigned, err := s.isProjectUnassigned(ctx, user.ID, project.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewUpdate().
		Model((*model.UserProject)(nil)).
		Set("perm = ?", params.Perm).
		Where("user_id = ? AND project_id = ?", user.ID, project.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       user.ID,
				ObjectDisplay:  user.Username,
				ObjectType:     model.EventTypeUserProject,
				Action:         model.EventActionUpdate,
				Attrs: map[string]interface{}{
					"project_id":      project.ID,
					"project_display": project.Name,
					"perm":            params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropProject implements the removal of an user from a project.
func (s *Users) DropProject(ctx context.Context, params model.UserProjectParams) error {
	user, err := s.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	unassigned, err := s.isProjectUnassigned(ctx, user.ID, project.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.UserProject)(nil)).
		Where("user_id = ? AND project_id = ?", user.ID, project.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       user.ID,
				ObjectDisplay:  user.Username,
				ObjectType:     model.EventTypeUserProject,
				Action:         model.EventActionDelete,
				Attrs: map[string]interface{}{
					"project_id":      project.ID,
					"project_display": project.Name,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Users) isProjectAssigned(ctx context.Context, userID, projectID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserProject)(nil)).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Users) isProjectUnassigned(ctx context.Context, userID, projectID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserProject)(nil)).
		Where("user_id = ? AND project_id = ?", userID, projectID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

func (s *Users) validatePerm(perm string) error {
	if err := validation.Validate(
		perm,
		validation.In("user", "admin"),
	); err != nil {
		return validate.Errors{
			Errors: []validate.Error{
				{
					Field: "perm",
					Error: fmt.Errorf("invalid permission value"),
				},
			},
		}
	}

	return nil
}

func (s *Users) validate(ctx context.Context, record *model.User, _ bool) error {
	errs := validate.Errors{}

	if err := validation.Validate(
		record.Username,
		validation.Required,
		validation.Length(3, 255),
		validation.By(s.uniqueValueIsPresent(ctx, "username", record.ID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "username",
			Error: err,
		})
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Users) uniqueValueIsPresent(ctx context.Context, key, id string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.User)(nil)).
			Where("? = ?", bun.Ident(key), val)

		if id != "" {
			q = q.Where(
				"id != ?",
				id,
			)
		}

		exists, err := q.Exists(ctx)

		if err != nil {
			return err
		}

		if exists {
			return errors.New("is already taken")
		}

		return nil
	}
}

func (s *Users) validSort(val string) (string, bool) {
	if val == "" {
		return "username", true
	}

	val = strings.ToLower(val)

	for _, name := range []string{
		"username",
		"email",
		"fullname",
		"admin",
		"active",
	} {
		if val == name {
			return val, true
		}
	}

	return "username", true
}

func (s *Users) validGroupSort(val string) (string, bool) {
	if val == "" {
		return "group.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"slug": "group.slug",
		"name": "group.name",
	} {
		if val == key {
			return name, true
		}
	}

	return "group.name", true
}

func (s *Users) validProjectSort(val string) (string, bool) {
	if val == "" {
		return "project.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"slug": "project.slug",
		"name": "project.name",
	} {
		if val == key {
			return name, true
		}
	}

	return "project.name", true
}
