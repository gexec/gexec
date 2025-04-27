package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/Machiel/slugify"
	"github.com/dchest/uniuri"
	"github.com/gexec/gexec/pkg/model"
	"github.com/gexec/gexec/pkg/validate"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/uptrace/bun"
)

// Groups provides all database operations related to groups.
type Groups struct {
	client *Store
}

// List implements the listing of all users.
func (s *Groups) List(ctx context.Context, params model.ListParams) ([]*model.Group, int64, error) {
	records := make([]*model.Group, 0)

	q := s.client.handle.NewSelect().
		Model(&records)

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
func (s *Groups) Show(ctx context.Context, name string) (*model.Group, error) {
	record := &model.Group{}

	if err := s.client.handle.NewSelect().
		Model(record).
		Where("id = ? OR slug = ?", name, name).
		Scan(ctx); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return record, ErrGroupNotFound
		}

		return record, err
	}

	return record, nil
}

// Create implements the create of a new group.
func (s *Groups) Create(ctx context.Context, record *model.Group) (*model.Group, error) {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			"",
		)
	}

	if err := s.validate(ctx, record, false); err != nil {
		return nil, err
	}

	if err := s.client.handle.RunInTx(ctx, &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		if _, err := tx.NewInsert().
			Model(record).
			Exec(ctx); err != nil {
			return err
		}

		if _, err := tx.NewInsert().
			Model(&model.UserGroup{
				GroupID: record.ID,
				UserID:  s.client.principal.ID,
				Perm:    model.UserGroupAdminPerm,
			}).
			Exec(ctx); err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      record.ID,
				ObjectDisplay: record.Name,
				ObjectType:    model.EventTypeGroup,
				Action:        model.EventActionCreate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, record.ID)
}

// Update implements the update of an existing group.
func (s *Groups) Update(ctx context.Context, record *model.Group) (*model.Group, error) {
	if record.Slug == "" {
		record.Slug = s.slugify(
			ctx,
			"slug",
			record.Name,
			record.ID,
		)
	}

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
				ObjectDisplay: record.Name,
				ObjectType:    model.EventTypeGroup,
				Action:        model.EventActionUpdate,
			},
		)).
		Exec(ctx); err != nil {
		return nil, err
	}

	return s.Show(ctx, record.ID)
}

// Delete implements the deletion of a group.
func (s *Groups) Delete(ctx context.Context, name string) error {
	record, err := s.Show(ctx, name)

	if err != nil {
		return err
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.Group)(nil)).
		Where("id = ? OR slug = ?", name, name).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      record.ID,
				ObjectDisplay: record.Name,
				ObjectType:    model.EventTypeGroup,
				Action:        model.EventActionDelete,
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// ListUsers implements the listing of all users for a group.
func (s *Groups) ListUsers(ctx context.Context, params model.UserGroupParams) ([]*model.UserGroup, int64, error) {
	records := make([]*model.UserGroup, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("User").
		Relation("Group").
		Where("group_id = ?", params.GroupID)

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

// AttachUser implements the attachment of a group to an user.
func (s *Groups) AttachUser(ctx context.Context, params model.UserGroupParams) error {
	group, err := s.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	assigned, err := s.isUserAssigned(ctx, group.ID, user.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.UserGroup{
		GroupID: group.ID,
		UserID:  user.ID,
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
				ObjectID:      group.ID,
				ObjectDisplay: group.Name,
				ObjectType:    model.EventTypeUserProject,
				Action:        model.EventActionCreate,
				Attrs: map[string]interface{}{
					"user_id":      user.ID,
					"user_display": user.Username,
					"perm":         params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// PermitUser implements the permission update for a user on a group.
func (s *Groups) PermitUser(ctx context.Context, params model.UserGroupParams) error {
	group, err := s.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	unassigned, err := s.isUserUnassigned(ctx, group.ID, user.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewUpdate().
		Model((*model.UserGroup)(nil)).
		Set("perm = ?", params.Perm).
		Where("group_id = ? AND user_id = ?", group.ID, user.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      group.ID,
				ObjectDisplay: group.Name,
				ObjectType:    model.EventTypeUserProject,
				Action:        model.EventActionUpdate,
				Attrs: map[string]interface{}{
					"user_id":      user.ID,
					"user_display": user.Username,
					"perm":         params.Perm,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

// DropUser implements the removal of a group from an user.
func (s *Groups) DropUser(ctx context.Context, params model.UserGroupParams) error {
	group, err := s.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	user, err := s.client.Users.Show(ctx, params.UserID)

	if err != nil {
		return err
	}

	unassigned, err := s.isUserUnassigned(ctx, group.ID, user.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.UserGroup)(nil)).
		Where("group_id = ? AND user_id = ?", group.ID, user.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ObjectID:      group.ID,
				ObjectDisplay: group.Name,
				ObjectType:    model.EventTypeUserProject,
				Action:        model.EventActionDelete,
				Attrs: map[string]interface{}{
					"user_id":      user.ID,
					"user_display": user.Username,
				},
			},
		)).
		Exec(ctx); err != nil {
		return err
	}

	return nil
}

func (s *Groups) isUserAssigned(ctx context.Context, groupID, userID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserGroup)(nil)).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Groups) isUserUnassigned(ctx context.Context, groupID, userID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.UserGroup)(nil)).
		Where("group_id = ? AND user_id = ?", groupID, userID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

// ListProjects implements the listing of all projects for a group.
func (s *Groups) ListProjects(ctx context.Context, params model.GroupProjectParams) ([]*model.GroupProject, int64, error) {
	records := make([]*model.GroupProject, 0)

	q := s.client.handle.NewSelect().
		Model(&records).
		Relation("Group").
		Relation("Project").
		Where("group_id = ?", params.GroupID)

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

// AttachProject implements the attachment of a group to a project.
func (s *Groups) AttachProject(ctx context.Context, params model.GroupProjectParams) error {
	group, err := s.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	assigned, err := s.isProjectAssigned(ctx, group.ID, project.ID)

	if err != nil {
		return err
	}

	if assigned {
		return ErrAlreadyAssigned
	}

	record := &model.GroupProject{
		GroupID:   group.ID,
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
				ObjectID:       group.ID,
				ObjectDisplay:  group.Name,
				ObjectType:     model.EventTypeGroupProject,
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

// PermitProject implements the permission update for a project on a group.
func (s *Groups) PermitProject(ctx context.Context, params model.GroupProjectParams) error {
	group, err := s.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	unassigned, err := s.isProjectUnassigned(ctx, group.ID, project.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewUpdate().
		Model((*model.GroupProject)(nil)).
		Set("perm = ?", params.Perm).
		Where("group_id = ? AND project_id = ?", group.ID, project.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       group.ID,
				ObjectDisplay:  group.Name,
				ObjectType:     model.EventTypeGroupProject,
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

// DropProject implements the removal of a group from a project.
func (s *Groups) DropProject(ctx context.Context, params model.GroupProjectParams) error {
	group, err := s.Show(ctx, params.GroupID)

	if err != nil {
		return err
	}

	project, err := s.client.Projects.Show(ctx, params.ProjectID)

	if err != nil {
		return err
	}

	unassigned, err := s.isProjectUnassigned(ctx, group.ID, project.ID)

	if err != nil {
		return err
	}

	if unassigned {
		return ErrNotAssigned
	}

	if _, err := s.client.handle.NewDelete().
		Model((*model.GroupProject)(nil)).
		Where("group_id = ? AND project_id = ?", group.ID, project.ID).
		Exec(ctx); err != nil {
		return err
	}

	if _, err := s.client.handle.NewInsert().
		Model(model.PrepareEvent(
			s.client.principal,
			&model.Event{
				ProjectID:      project.ID,
				ProjectDisplay: project.Name,
				ObjectID:       group.ID,
				ObjectDisplay:  group.Name,
				ObjectType:     model.EventTypeGroupProject,
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

func (s *Groups) isProjectAssigned(ctx context.Context, groupID, projectID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.GroupProject)(nil)).
		Where("group_id = ? AND project_id = ?", groupID, projectID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

func (s *Groups) isProjectUnassigned(ctx context.Context, groupID, projectID string) (bool, error) {
	count, err := s.client.handle.NewSelect().
		Model((*model.GroupProject)(nil)).
		Where("group_id = ? AND project_id = ?", groupID, projectID).
		Count(ctx)

	if err != nil {
		return false, err
	}

	return count < 1, nil
}

func (s *Groups) validatePerm(perm string) error {
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

func (s *Groups) validate(ctx context.Context, record *model.Group, _ bool) error {
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

func (s *Groups) uniqueValueIsPresent(ctx context.Context, key, id string) func(value interface{}) error {
	return func(value interface{}) error {
		val, _ := value.(string)

		q := s.client.handle.NewSelect().
			Model((*model.Group)(nil)).
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

func (s *Groups) slugify(ctx context.Context, column, value, id string) string {
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
			Model((*model.Group)(nil)).
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

func (s *Groups) validSort(val string) (string, bool) {
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

func (s *Groups) validUserSort(val string) (string, bool) {
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

func (s *Groups) validProjectSort(val string) (string, bool) {
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
