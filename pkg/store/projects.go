package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/uptrace/bun"
)

// Projects provides all database operations related to projects.
type Projects struct {
	client *Store
}

// AllowedIDs returns a list of project IDs the user is allowed to access.
func (s *Projects) AllowedIDs() []string {
	result := []string{}

	if s.client.principal == nil {
		return result
	}

	for _, p := range s.client.principal.Projects {
		result = append(result, p.ProjectID)
	}

	for _, t := range s.client.principal.Teams {
		for _, p := range t.Team.Projects {
			result = append(result, p.ProjectID)
		}
	}

	return result
}

// List implements the listing of all users.
func (s *Projects) List(ctx context.Context, params model.ListParams) ([]*model.Project, int64, error) {
	records := make([]*model.Project, 0)

	q := s.client.handle.NewSelect().
		Model(&records)

	if !s.client.principal.Admin {
		q = q.Where(
			"id IN (?)",
			bun.In(s.AllowedIDs()),
		)
	}

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
func (s *Projects) Show(ctx context.Context, name string) (*model.Project, error) {
	record := &model.Project{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Where("id = ? OR slug = ?", name, name).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrProjectNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new project.
func (s *Projects) Create(ctx context.Context, record *model.Project) error {
	if record.Slug == "" {
		record.Slug = Slugify(
			ctx,
			s.client.handle.NewSelect().
				Model((*model.Project)(nil)),
			"slug",
			record.Name,
			"",
			false,
		)
	}

	if err := s.validate(ctx, record, false); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(record).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(&model.UserProject{
			ProjectID: record.ID,
			UserID:    s.client.principal.ID,
			Perm:      model.UserProjectAdminPerm,
		}).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Update implements the update of an existing project.
func (s *Projects) Update(ctx context.Context, record *model.Project) error {
	if record.Slug == "" {
		record.Slug = Slugify(
			ctx,
			s.client.handle.NewSelect().
				Model((*model.Project)(nil)),
			"slug",
			record.Name,
			record.ID,
			false,
		)
	}

	if err := s.validate(ctx, record, true); err != nil {
		return err
	}

	if _, err := s.client.handle.NewUpdate().
		Model(record).
		Where("id = ?", record.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Delete implements the deletion of a project.
func (s *Projects) Delete(ctx context.Context, name string) error {
	if _, err := s.client.handle.NewDelete().
		Model((*model.Project)(nil)).
		Where("id = ? OR slug = ?", name, name).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ListTeams implements the listing of all teams for a project.
func (s *Projects) ListTeams(ctx context.Context, params model.TeamProjectParams) ([]*model.TeamProject, int64, error) {
	records := make([]*model.TeamProject, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Project").
		Relation("Team").
		Where("project_id = ?", params.ProjectID)

	if val, ok := s.validTeamSort(params.Sort); ok {
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

// AttachTeam implements the attachment of a project to a team.
func (s *Projects) AttachTeam(ctx context.Context, params model.TeamProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	team, err := s.client.Teams.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	assigned, err := s.isTeamAssigned(ctx, project.ID, team.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.TeamProject{
		ProjectID: project.ID,
		TeamID:    team.ID,
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

	return nil
}

// PermitTeam implements the permission update for a team on a project.
func (s *Projects) PermitTeam(ctx context.Context, params model.TeamProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	team, err := s.client.Teams.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	unassigned, err := s.isTeamUnassigned(ctx, project.ID, team.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewUpdate().
		Model((*model.TeamProject)(nil)).
		Set("perm = ?", params.Perm).
		Where("project_id = ? AND team_id = ?", project.ID, team.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropTeam implements the removal of a project from a team.
func (s *Projects) DropTeam(ctx context.Context, params model.TeamProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	team, err := s.client.Teams.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	unassigned, err := s.isTeamUnassigned(ctx, project.ID, team.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.TeamProject)(nil)).
		Where("project_id = ? AND team_id = ?", project.ID, team.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Projects) isTeamAssigned(ctx context.Context, projectID, teamID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.TeamProject)(nil)).
		Where("project_id = ? AND team_id = ?", projectID, teamID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Projects) isTeamUnassigned(ctx context.Context, projectID, teamID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.TeamProject)(nil)).
		Where("project_id = ? AND team_id = ?", projectID, teamID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

// ListUsers implements the listing of all users for a project.
func (s *Projects) ListUsers(ctx context.Context, params model.UserProjectParams) ([]*model.UserProject, int64, error) {
	records := make([]*model.UserProject, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Project").
		Relation("User").
		Where("project_id = ?", params.ProjectID)

	if val, ok := s.validUserSort(params.Sort); ok {
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

// AttachUser implements the attachment of a project to an user.
func (s *Projects) AttachUser(ctx context.Context, params model.UserProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	assigned, err := s.isUserAssigned(ctx, project.ID, user.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.UserProject{
		ProjectID: project.ID,
		UserID:    user.ID,
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

	return nil
}

// PermitUser implements the permission update for an user on a project.
func (s *Projects) PermitUser(ctx context.Context, params model.UserProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	unassigned, err := s.isUserUnassigned(ctx, project.ID, user.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewUpdate().
		Model((*model.UserProject)(nil)).
		Set("perm = ?", params.Perm).
		Where("project_id = ? AND user_id = ?", project.ID, user.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropUser implements the removal of a project from an user.
func (s *Projects) DropUser(ctx context.Context, params model.UserProjectParams) error {
	project, err := s.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	unassigned, err := s.isUserUnassigned(ctx, project.ID, user.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.UserProject)(nil)).
		Where("project_id = ? AND user_id = ?", project.ID, user.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Projects) isUserAssigned(ctx context.Context, projectID, userID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserProject)(nil)).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Projects) isUserUnassigned(ctx context.Context, projectID, userID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserProject)(nil)).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

func (s *Projects) validatePerm(perm string) error {
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

func (s *Projects) validate(ctx context.Context, record *model.Project, _ bool) error {
	errs := validate.Errors{}

	if err := validation.Validate(
		record.Slug,
		validation.Required,
		validation.Length(3, 255),
		validation.By(s.uniqueValueIsPresent(ctx, "slug", record.ID)),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "slug",
			Error: err,
		})
	}

	if err := validation.Validate(
		record.Name,
		validation.Required,
		validation.Length(3, 255),
	); err != nil {
		errs.Errors = append(errs.Errors, validate.Error{
			Field: "name",
			Error: err,
		})
	}

	if len(errs.Errors) > 0 {
		return errs
	}

	return nil
}

func (s *Projects) uniqueValueIsPresent(ctx context.Context, key, id string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Project)(nil)).
			Where(fmt.Sprintf("%s = ?", key), val)

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

func (s *Projects) validSort(val string) (string, bool) {
	if val == "" {
		return "name", true
	}

	val = strings.ToLower(val)

	for _, name := range []string{
		"slug",
		"name",
	} {
		if val == name {
			return val, true
		}
	}

	return "name", true
}

func (s *Projects) validTeamSort(val string) (string, bool) {
	if val == "" {
		return "team.name", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"slug": "team.slug",
		"name": "team.name",
	} {
		if val == key {
			return name, true
		}
	}

	return "team.name", true
}

func (s *Projects) validUserSort(val string) (string, bool) {
	if val == "" {
		return "user.username", true
	}

	val = strings.ToLower(val)

	for key, name := range map[string]string{
		"username": "user.username",
		"email":    "user.email",
		"fullname": "user.fullname",
		"admin":    "user.admin",
		"active":   "user.active",
	} {
		if val == key {
			return name, true
		}
	}

	return "user.username", true
}
