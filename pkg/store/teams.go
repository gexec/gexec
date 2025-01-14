package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/dchest/uniuri"
	"github.com/genexec/genexec/pkg/model"
	"github.com/genexec/genexec/pkg/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/uptrace/bun"
)

// Teams provides all database operations related to teams.
type Teams struct {
	client *Store
}

// List implements the listing of all users.
func (s *Teams) List(ctx context.Context, params model.ListParams) ([]*model.Team, int64, error) {
	records := make([]*model.Team, 0)

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
func (s *Teams) Show(ctx context.Context, name string) (*model.Team, error) {
	record := &model.Team{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Relation("Auths").
		Where("id = ? OR slug = ?", name, name).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrTeamNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new team.
func (s *Teams) Create(ctx context.Context, record *model.Team) error {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			"",
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
		Model(&model.UserTeam{
			TeamID: record.ID,
			UserID: s.client.principal.ID,
			Perm:   model.UserTeamAdminPerm,
		}).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// Update implements the update of an existing team.
func (s *Teams) Update(ctx context.Context, record *model.Team) error {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			record.ID,
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

// Delete implements the deletion of a team.
func (s *Teams) Delete(ctx context.Context, name string) error {
	if _, err := s.client.handle.NewDelete().
		Model((*model.Team)(nil)).
		Where("id = ? OR slug = ?", name, name).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ListUsers implements the listing of all users for a team.
func (s *Teams) ListUsers(ctx context.Context, params model.UserTeamParams) ([]*model.UserTeam, int64, error) {
	records := make([]*model.UserTeam, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("User").
		Relation("Team").
		Where("team_id = ?", params.TeamID)

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

// AttachUser implements the attachment of a team to an user.
func (s *Teams) AttachUser(ctx context.Context, params model.UserTeamParams) error {
	team, err := s.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	assigned, err := s.isUserAssigned(ctx, team.ID, user.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.UserTeam{
		TeamID: team.ID,
		UserID: user.ID,
		Perm:   params.Perm,
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

// PermitUser implements the permission update for a user on a team.
func (s *Teams) PermitUser(ctx context.Context, params model.UserTeamParams) error {
	team, err := s.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	unassigned, err := s.isUserUnassigned(ctx, team.ID, user.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewUpdate().
		Model((*model.UserTeam)(nil)).
		Set("perm = ?", params.Perm).
		Where("team_id = ? AND user_id = ?", team.ID, user.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropUser implements the removal of a team from an user.
func (s *Teams) DropUser(ctx context.Context, params model.UserTeamParams) error {
	team, err := s.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	unassigned, err := s.isUserUnassigned(ctx, team.ID, user.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.UserTeam)(nil)).
		Where("team_id = ? AND user_id = ?", team.ID, user.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Teams) isUserAssigned(ctx context.Context, teamID, userID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserTeam)(nil)).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Teams) isUserUnassigned(ctx context.Context, teamID, userID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserTeam)(nil)).
		Where("team_id = ? AND user_id = ?", teamID, userID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

// ListProjects implements the listing of all projects for a team.
func (s *Teams) ListProjects(ctx context.Context, params model.TeamProjectParams) ([]*model.TeamProject, int64, error) {
	records := make([]*model.TeamProject, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Team").
		Relation("Project").
		Where("team_id = ?", params.TeamID)

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

// AttachProject implements the attachment of a team to a project.
func (s *Teams) AttachProject(ctx context.Context, params model.TeamProjectParams) error {
	team, err := s.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	assigned, err := s.isProjectAssigned(ctx, team.ID, project.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.TeamProject{
		TeamID:    team.ID,
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

	return nil
}

// PermitProject implements the permission update for a project on a team.
func (s *Teams) PermitProject(ctx context.Context, params model.TeamProjectParams) error {
	team, err := s.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	unassigned, err := s.isProjectUnassigned(ctx, team.ID, project.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewUpdate().
		Model((*model.TeamProject)(nil)).
		Set("perm = ?", params.Perm).
		Where("team_id = ? AND project_id = ?", team.ID, project.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropProject implements the removal of a team from a project.
func (s *Teams) DropProject(ctx context.Context, params model.TeamProjectParams) error {
	team, err := s.Show(ctx, params.TeamID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	unassigned, err := s.isProjectUnassigned(ctx, team.ID, project.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.TeamProject)(nil)).
		Where("team_id = ? AND project_id = ?", team.ID, project.ID).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Teams) isProjectAssigned(ctx context.Context, teamID, projectID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.TeamProject)(nil)).
		Where("team_id = ? AND project_id = ?", teamID, projectID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Teams) isProjectUnassigned(ctx context.Context, teamID, projectID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.TeamProject)(nil)).
		Where("team_id = ? AND project_id = ?", teamID, projectID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

func (s *Teams) validatePerm(perm string) error {
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

func (s *Teams) validate(ctx context.Context, record *model.Team, _ bool) error {
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
		validation.By(s.uniqueValueIsPresent(ctx, "name", record.ID)),
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

func (s *Teams) uniqueValueIsPresent(ctx context.Context, key, id string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Team)(nil)).
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

func (s *Teams) slugify(ctx context.Context, column, value, id string) string {
	var (
		slug string
	)

	for i := 0; true; i++ {
		if i == 0 {
			slug = slugify.Slugify(value)
		} else {
			slug = slugify.Slugify(
				fmt.Sprintf("%s-%s", value, uniuri.NewLen(6)),
			)
		}

		query := s.client.handle.NewSelect().
			Model((*model.Team)(nil)).
			Where("? = ?", bun.Ident(column), slug)

		if id != "" {
			query = query.Where(
				"id != ?",
				id,
			)
		}

		if count, err := query.Count(
			ctx,
		); err == nil && count == 0 {
			break
		}
	}

	return slug
}

func (s *Teams) validSort(val string) (string, bool) {
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

func (s *Teams) validUserSort(val string) (string, bool) {
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

func (s *Teams) validProjectSort(val string) (string, bool) {
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
